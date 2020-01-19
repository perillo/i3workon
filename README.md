# i3workon [![GoDoc](https://godoc.org/github.com/perillo/i3workon?status.svg)](http://godoc.org/github.com/perillo/i3workon)

`i3workon` is a simple tool designed to make it easy to start to work on a Go
project, when using the *i3* window manager.

`i3workon` will:

 1. If `-workspace` is specified, switch to the specified *i3* `workspace`.
 2. Start a new terminal with the specified path set to its working directory.
 3. Open all the .go files in the specified path, including nested packages, in
    a new editor.  The new editor will have its working directory set to the
    specified path.

## Usage

    i3workon -workspace -terminal -editor your.project.path

The terminal used will be determined from, in order:

 1. the `-terminal` flag
 2. the `TERMINAL` environment variable

The editor used will be determined from, in order:

 1. the `-editor` flag
 2. the `VISUAL` environment variable
 3. the `EDITOR` environment variable

## TODO

If no terminal or editor is available, select a default one based on the *OS*.

## NOTES

`i3workon` is a fork of https://github.com/perillo/workon with the support for
workspaces.

In theory it is possible to add support for virtual desktops to the original
`workon` command.  On *UNIX* systems with *Xorg*, the `wmctrl` command can be
used to switch to a specified virtual desktop (an integer starting from 0).
*MacOS* and *Windows* also seems to support virtual desktops.

Unfortunately `wmctrl` does not work as expected on *i3*, since *i3* workspaces
are identified by a generic string, and internally mapped to virtual desktops.
