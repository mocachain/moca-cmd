# Changelog

## Unreleased
FEATURES
* [#23](https://github.com/mocachain/moca-cmd/pull/23) Add `--delegate` to `object put` and `object update`: the primary SP creates the object, or replaces its content, on chain on the user's behalf.

IMPROVEMENTS
* [#27](https://github.com/mocachain/moca-cmd/pull/27) Build with Go 1.26.11 (security fixes to cgo, the go command, the compiler, the linker and `net/http`): both `go.mod` files, `.tool-versions`, the Dockerfile, the Makefile `GO_TOOLCHAIN` pin and the CI `go-version` pins move from 1.26.10; both stages of the Dockerfile move from `bookworm` to `trixie` (`golang:1.26.11-trixie`); the goreleaser-cross release targets keep the `v1.26.10` image and set `GOTOOLCHAIN=go1.26.11` inside it until a `v1.26.11` image is published.
* [#25](https://github.com/mocachain/moca-cmd/pull/25) Build against moca `v2.0.0-rc1`, moca-cosmos-sdk `v1.4.0-rc1`, moca-cometbft `v1.3.0-rc1` and moca-go-sdk `main`, with the dependency replaces aligned to moca `v2.0.0-rc1`. Building from source now requires Go 1.26.10.

BUG FIXES
* [#24](https://github.com/mocachain/moca-cmd/pull/24) Commands that fail now exit with status 1 instead of 0, and report the failure once as `run command error: ...`. Scripts that relied on a zero exit status after a failed command must be updated.

## v0.2.0
FEATURES
