# Changelog

## Unreleased
FEATURES
* [#23](https://github.com/mocachain/moca-cmd/pull/23) Add `--delegate` to `object put` and `object update`: the primary SP creates the object, or replaces its content, on chain on the user's behalf.

BUG FIXES
* [#24](https://github.com/mocachain/moca-cmd/pull/24) Commands that fail now exit with status 1 instead of 0, and report the failure once as `run command error: ...`. Scripts that relied on a zero exit status after a failed command must be updated.

## v0.2.0
FEATURES
