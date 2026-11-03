# Changelog

## Unreleased
FEATURES
* [#23](https://github.com/mocachain/moca-cmd/pull/23) Add `--delegate` to `object put` and `object update`: the primary SP creates the object, or replaces its content, on chain on the user's behalf.

IMPROVEMENTS
* [#25](https://github.com/mocachain/moca-cmd/pull/25) Build against moca `v2.0.0-rc1`, moca-cosmos-sdk `v1.4.0-rc1`, moca-cometbft `v1.3.0-rc1` and moca-go-sdk `main`, with the dependency replaces aligned to moca `v2.0.0-rc1`. Building from source now requires Go 1.26.10.

BUG FIXES
* [#24](https://github.com/mocachain/moca-cmd/pull/24) Commands that fail now exit with status 1 instead of 0, and report the failure once as `run command error: ...`. Scripts that relied on a zero exit status after a failed command must be updated.

## v0.2.0
FEATURES
