# txtidy

This is just a very, very simple helper fixes a couple of inconsistencies in all
text files within a given directory. It performs these changes:

* dos2unix
* Trailing spaces at the end of lines are removed
* All files are made to end with exactly one newline character
* UTF8-BOM is removed.

I'm using this whenever I get a bunch of HTML/CSS/JS files and want to commit
them to a repository. Having clean files prevent ugly diffs later on.

## Installation

```bash
go install go.xrstf.de/txtidy
```

## Usage

Just run the ``txtidy`` binary and give the filename patterns to match against
all files from the starting directory resursively.

```
Usage of txtidy:
  -a, --all          run on all files, i.e. do not exclude [.git .hg .svn node_modules bower_components vendor]
  -d, --dir string   root directory to search files in
  -v, --verbose      whether or not to print all visited files
  -V, --version      show version info and exit immediately
```

For example:

```bash
$ txtidy *.php
$ txtidy -v *.css *.less
```
