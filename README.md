# scream-go
[![Go Reference](https://pkg.go.dev/badge/github.com/mengelbart/scream-go.svg)](https://pkg.go.dev/github.com/mengelbart/scream-go)

scream-go is a CGO-wrapper for the [SCReAM-implementation](https://github.com/EricssonResearch/scream) by EricssonResearch.

SCReAM is a mobile optimised congestion control algorithm for realtime interactive media.

## Building

`go get` is all that is needed. The SCReAM sources are vendored into this
package and compiled by cgo along with the wrapper, so the package works on any
platform with a C++ toolchain. Building requires `CGO_ENABLED=1`.

The vendored files are `RtpQueue.h`, `ScreamRx.h`, `ScreamTx.h`, `ScreamRx.cpp`,
`ScreamTx.cpp`, `ScreamV2Tx.cpp` and `ScreamV2TxStream.cpp`. Each one carries a
header saying where it came from. They must not be edited by hand. The
wrapper's own C++ files are the ones with a `C` suffix: `ScreamTxC`, `ScreamRxC`
and `RtpQueueC`.

## Configuration

SCReAM's tuning knobs are exposed as functional options. `TxOption` covers the
sender, `RxOption` the receiver and `StreamOption` an individual stream:

```go
tx := scream.NewTx(
	scream.WithL4S(true),
	scream.WithPacketPacing(true),
	scream.WithCwndMinLow(5000),
)
defer tx.Close()

err := tx.RegisterNewStream(
	queue, ssrc, priority, minBitrate, startBitrate, maxBitrate,
	scream.WithHysteresis(0.1),
)
if err != nil {
	return err
}
```

Settings that are meant to change while running are methods instead:
`UpdateBitrateStream`, `SetTargetPriority`, `SetMaxTotalBitrate` and
`SetMssListMinPacketsInFlight`.

Note that packet pacing is disabled by default, which differs from SCReAM's own
default. Pass `WithPacketPacing(true)` to enable it.

## Updating SCReAM

The original SCReAM implementation is tracked as a Git submodule, which is the
source the vendored files are generated from. The submodule itself is not part
of the published Go module, so the sources have to be vendored into the
repository for downstream builds to see them.

To update:

```
git submodule update --remote scream
make vendor
make test
```

Then commit the result and create a new tag. `make check-vendor` verifies that
the vendored sources match the submodule.

## LICENSES

The vendored SCReAM sources are licensed under BSD 2-Clause with the original
Ericsson copyright. That license is retained in
[LICENSE_SCREAM](LICENSE_SCREAM), which [the vendoring
script](./scripts/vendor-scream.sh) copies from upstream so the two cannot
drift. Each vendored file names its origin and points at that file.

Everything else is licensed under the same BSD 2-Clause license but with a different copyright. The license can be found in the [LICENSE file](LICENSE).

