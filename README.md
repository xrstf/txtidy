dos2unix
========

This is just a very, very simple helper that turns Windows into Unix newlines and
trims trailing spaces in each line. With ab it of shell scripting, one could
achieve similar results by doing

    find . -name '*.css' -exec dos2unix {} \;

But this program is easier to use and works well on Windows.

Usage
-----

First, ``go get https://github.com/xrstf/dos2unix-go``. Then you run it via

    dos2unix [-a] [-v] [-dir=<dir>] <pattern> <pattern> <pattern>

For example

    $ dos2unix *.php
    $ dos2unix -v *.css *.less
