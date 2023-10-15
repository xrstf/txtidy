// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/spf13/pflag"
)

// These variables get set by ldflags during compilation.
var (
	BuildTag    string
	BuildCommit string
	BuildDate   string // RFC3339 format ("2006-01-02T15:04:05Z07:00")
)

func printVersion() {
	fmt.Printf(
		"txtidy %s (%s), built with %s on %s\n",
		BuildTag,
		BuildCommit[:10],
		runtime.Version(),
		BuildDate,
	)
}

var (
	excludedDirs = []string{".git", ".hg", ".svn", "node_modules", "bower_components", "vendor"}

	dirFlag     = pflag.StringP("dir", "d", "", "root directory to search files in")
	verboseFlag = pflag.BoolP("verbose", "v", false, "whether or not to print all visited files")
	versionFlag = pflag.BoolP("version", "V", false, "show version info and exit immediately")
	allFlag     = pflag.BoolP("all", "a", false, fmt.Sprintf("run on all files, i.e. do not exclude %v", excludedDirs))
)

var trailingWhitespace = regexp.MustCompile(`(?m:[\t ]+$)`)

func main() {
	pflag.Parse()

	if *versionFlag {
		printVersion()
		return
	}

	verbose := *verboseFlag
	visitAll := *allFlag

	// find root dir
	rootDir := *dirFlag
	if rootDir == "" {
		dir, err := os.Getwd()
		if err != nil {
			log.Fatalf("Error: could not determine current working directory: %v", err)
		}

		rootDir = dir
	}

	// turn into an absolute path
	abs, err := filepath.Abs(rootDir)
	if err != nil {
		log.Fatalf("Error: could not determine absolute path for '%s': %v", rootDir, err)
	}

	rootDir = abs

	// check if the path actually exists
	stats, err := os.Stat(rootDir)
	if err != nil || !stats.IsDir() {
		log.Fatalf("Error: the given root directory '%s' could not be found.", rootDir)
	}

	// collect all file patterns
	patterns := pflag.Args()
	if len(patterns) == 0 {
		pflag.Usage()
		log.Fatalf("Error: no file pattern have been given.")
	}

	// check patterns
	for _, pattern := range patterns {
		_, err := filepath.Match(pattern, "dummy")
		if err != nil {
			log.Fatalf("Error: file pattern '%s' is invalid.", pattern)
		}
	}

	// let's walk the file tree

	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

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
			fmt.Printf("%s …", path)
		}

		// read file into memory
		content, err := os.ReadFile(path)
		if err != nil {
			if !verbose {
				fmt.Printf("%s …", path)
			}

			fmt.Printf(" error: %s\n", err.Error())
			return nil
		}

		// do magic
		original := content
		content = tidy(content)

		if !bytes.Equal(content, original) {
			if !verbose {
				fmt.Printf("%s …", path)
			}

			stats, _ := os.Stat(path)

			err := os.WriteFile(path, content, stats.Mode())
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
	if err != nil {
		log.Fatalf("Error: Failed to walk filesystem: %v", err)
	}
}

func tidy(content []byte) []byte {
	// turn Windows newlines into Unix newlines
	content = bytes.Replace(content, []byte{'\r'}, []byte{}, -1)

	// remove UTF BOMs
	if len(content) >= 3 && bytes.Equal(content[0:3], []byte("\xEF\xBB\xBF")) {
		content = content[3:]
	}

	// trim trailing whitespace in each line
	content = trailingWhitespace.ReplaceAllLiteral(content, []byte{})

	// trim leading and trailing file space
	content = bytes.TrimSpace(content)

	// and make sure the file ends with a newline character
	content = append(content, '\n')

	return content
}
