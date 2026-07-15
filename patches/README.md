# WPT Patch Overrides

Prefer polyfilling missing runtime globals or overriding harness behavior
from the Go test setup (see `encoding/test_setup_test.go`) over patching a
vendored WPT file. This keeps the files `scripts/sync-wpt.sh` fetches
byte-for-byte identical to upstream.

Reserve patch files for this directory only when the vendored source itself
must change -- e.g. a genuine bug in the WPT test, or content no
polyfill/override could address. Add the patch path to the corresponding
entry in `wpt.json` and re-run `scripts/sync-wpt.sh` to apply it. Requires
`curl`, `jq`, and `git` (the last only for entries with a `patch` field).

Patches should be created with `git diff` from the repository root to keep
paths consistent with `git apply`.

