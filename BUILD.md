# Building MV/Em

This guide applies to the development version of MV/Em written in Go known as 'mvemg'.

## Prerequisites

* Go version **1.9** or greater must be installed
* A GNU-compatible `make` must be installed

### Obtain Required Packages

* `go get github.com/SMerrony/aosvs-tools/simhTape`
* `go get github.com/SMerrony/dgemug/...`
* Install the `dginstr` command provided by dgemug as per the instructions in its README.md, ensure it is available on your PATH

### Obtain MV/Em Source Code

* `cd ~/go/src`
* `git clone https://github.com/SMerrony/mvemg.git`

## Build

Simply type `make` in the source directory.  The mvemg binary should pass its tests and build without any errors.

See UserGuide.md for operating instructions.
