# workon [![GoDoc](https://godoc.org/github.com/perillo/workon?status.svg)](http://godoc.org/github.com/perillo/workon)

`workon` is a simple tool designed to make it easy to start to work on a Go
project.

workon will:

 1. Start a new terminal with the specified path set to its working directory.
 2. Open all the .go files in the specified path, including nested packages, in
    a new editor.  The new editor will have its working directory set to the
    specified path.

## Usage

    workon your.project.path


## Limitations

Currently workon only support `urxvt` as terminal and `gvim` as editor.  Adding
support for other terminals and editor should be trivial.
