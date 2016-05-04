package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

var (
	dirFlag     = flag.String("dir", "", "root directory to search files in")
	verboseFlag = flag.Bool("v", false, "whether or not to print all visited files")
	allFlag     = flag.Bool("a", false, "run on all files, i.e. do not exclude .git, .hg, node_modules etc.")

	excludedDirs = []string{".git", ".hg", ".svn", "node_modules", "bower_components"}
)

func main() {
	flag.Parse()

	verbose := *verboseFlag
	visitAll := *allFlag

	// find root dir
	rootDir := *dirFlag
	if rootDir == "" {
		dir, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error: could not determine current working directory: %s\n", err.Error())
			os.Exit(1)
		}

		rootDir = dir
	}

	// turn into an absolute path
	abs, err := filepath.Abs(rootDir)
	if err != nil {
		fmt.Printf("Error: could not determine absolute path for '%s': %s\n", rootDir, err.Error())
		os.Exit(1)
	}

	rootDir = abs

	// check if the path actually exists
	stats, err := os.Stat(rootDir)
	if err != nil || !stats.IsDir() {
		fmt.Printf("Error: the given root directory '%s' could not be found.\n", rootDir)
		os.Exit(1)
	}

	// collect all file patterns
	patterns := flag.Args()
	if len(patterns) == 0 {
		fmt.Printf("Error: no file pattern have been given.\n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// check patterns
	for _, pattern := range patterns {
		_, err := filepath.Match(pattern, "dummy")
		if err != nil {
			fmt.Printf("Error: file pattern '%s' is invalid.\n", pattern)
			os.Exit(1)
		}
	}

	// let's walk the file tree

	trailingWhitespace := regexp.MustCompile(`(?m:[\t ]+$)`)

	filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		filename := filepath.Base(path)

		// determine if we reached a directory we ignore
		if !visitAll {
			for _, excl := range excludedDirs {
				if filename == excl {
					return filepath.SkipDir
				}
			}
		}

		// check if the filename matches the given file pattern(s)
		matched := false

		for _, pattern := range patterns {
			match, _ := filepath.Match(pattern, filename)
			if match {
				matched = true
				break
			}
		}

		if !matched {
			return nil
		}

		if verbose {
			fmt.Printf("%s...", path)
		}

		// read file into memory
		content, err := ioutil.ReadFile(path)
		if err != nil {
			if !verbose {
				fmt.Printf("%s...", path)
			}

			fmt.Printf(" error: %s\n", err.Error())
			return nil
		}

		original := content

		// do magic
		// turn Windows newlines into Unix newlines
		content = bytes.Replace(content, []byte{'\r'}, []byte{}, -1)

		// trim trailing whitespace in each line
		content = trailingWhitespace.ReplaceAllLiteral(content, []byte{})

		// trim leading and trailing file space
		content = bytes.TrimSpace(content)

		// and make sure the file ends with a newline character
		content = append(content, '\n')

		if bytes.Compare(content, original) != 0 {
			if !verbose {
				fmt.Printf("%s...", path)
			}

			stats, _ := os.Stat(path)

			err := ioutil.WriteFile(path, content, stats.Mode())
			if err != nil {
				fmt.Printf(" error: %s\n", err.Error())
			} else {
				fmt.Printf(" fixed.\n")
			}
		} else if verbose {
			fmt.Printf("\n")
		}

		return nil
	})
}
