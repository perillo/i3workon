# i3workon [![GoDoc](https://godoc.org/github.com/perillo/i3workon?status.svg)](http://godoc.org/github.com/perillo/i3workon)

`i3workon` is a simple tool designed to make it easy to start to work on a *Go*
project, when using the *i3* window manager.

`i3workon pattern` will:

 1. Find the *module* matching `pattern`.

 2. Switch to a new *i3* *workspace*.

 3. Start a new terminal emulator with the *module* root directory set to its
    working directory.

 4. Open in a new editor all the `.go` source files in all the packages in the
    *module*.  The new editor will have its working directory set to the
    *module* root directory.

## Usage

    Usage: i3workon [flags] pattern
    Flags:
      -editor string
           editor to use (default "i3-sensible-editor")
      -terminal string
           terminal to use (default "i3-sensible-terminal")
      -workspace int
           workspace where to switch to

`pattern` has the same syntax as used in the `go` tool, with the difference
that absolute and relative directories are not supported.  The specified
`pattern` will be used to match modules in `$GOPATH`.

When the `-workspace` flag is not specified or is set to `0`, `i3workon` will
switch to the next available *workspace*.  Specifying `-1` or any other
negative number will instruct `i3workon` to remain on the current workspace.

`i3workon` will use a named *workspace* in the format `[n][:][NAME]`, where `n`
is set to the *workspace* number as specified in the `-workspace` flag and
`NAME` is set to the *module* short name (the last segment in the
*module path*).


## Configuring i3

Since `i3workon` uses named workspaces, it is necessary to modify the default
*i3* user configuration (in `~/.config/i3/config` or `~/.i3/config`) to be able
to switch to the new *workspace* by its number.

As an example:

```
bindsym $mod+[n] workspace [n] => bindsym $mod+[n] workspace number [n]
```
and
```
bindsym $mod+Shift+[n] move container to workspace [n] => bindsym $mod+Shift+[n] move container to workspace number [n]
```


## Examples

```
$ i3workon -workspace 10 .../i3workon
```

```
$ i3workon github.com/perillo/i3workon
```


## Screenshot

![screenshot](https://user-images.githubusercontent.com/6217088/74610387-12265b00-50f3-11ea-82c7-af0d58d42435.jpg)


## Notes

`i3workon` is a fork of https://github.com/perillo/workon with the support for
workspaces.

In theory it is possible to add support for virtual desktops to the original
`workon` command.  On *UNIX* systems with *Xorg*, the `wmctrl` command can be
used to switch to a specified virtual desktop (an integer starting from 0).
*MacOS* and *Windows* also seems to support virtual desktops.

Unfortunately `wmctrl` does not work as expected on *i3*, since *i3* workspaces
are identified by a generic string, and internally mapped to virtual desktops.
