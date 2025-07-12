# scream-go
[![Go Reference](https://pkg.go.dev/badge/github.com/mengelbart/scream-go.svg)](https://pkg.go.dev/github.com/mengelbart/scream-go)

scream-go is a CGO-wrapper for the [SCReAM-implementation](https://github.com/EricssonResearch/scream) by EricssonResearch.

SCReAM is a mobile optimised congestion control algorithm for realtime interactive media.

## Building and Release

The SCReAM library is included as a static library. To make a new release of the wrapper package, clone the repository and its submodules and use the Makefile to build the static SCReAM library. Then commit the resulting [archive](libscream.a) to Git and create a new tag to release the new version.

## LICENSES

The original SCReAM implementation is included as a Git submodule. To allow the go build system to build this package as a dependency of other packages, some header files are needed. These header files are copied from the original SCReAM repository to the [include](./include/) directory. A copy of the original LICENSE and copyright is included in the header of each file and in the [LICENSE_SCREAM file](LICENSE_SCREAM).

Everything else is licensed under the same BSD 2-Clause license but with a different copyright. The license can be found in the [LICENSE file](LICENSE).

