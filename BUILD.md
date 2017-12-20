# Building MV/Em

This guide currently applies to the development version of MV/Em written in Go.

## Prerequisites

* Go version **1.9** or greater must be installed
* A GNU-compatible `make` must be installed

### Obtain Required Packages

* `go get githb.com/SMerrony/aosvs-tools/simhTape`

### Obtain MV/Em Source Code

* `cd ~/go/src`
* `git clone http://stephenmerrony.co.uk:6000/steve/mvemg.git`

## Build

Simply type `make` in the source directory.  The mvemg binary should pass its tests and build without any errors.

See UserGuide.md for operating instructions.
