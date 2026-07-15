# WPT Patch Overrides

Prefer polyfilling missing runtime globals or overriding harness behavior
from the Go test setup (see `encoding/test_setup_test.go`) over patching a
vendored WPT file. This keeps the files `scripts/sync-wpt.sh` fetches
byte-for-byte identical to upstream.

Reserve patch files for this directory only when the vendored source itself
must change -- e.g. a genuine bug in the WPT test, or content no
polyfill/override could address. `scripts/sync-wpt.sh` applies every
`*.patch` file found under this directory, in sorted order, after
re-fetching the pinned directories -- no separate list to keep in sync.
Requires `git` only.

Patches should be created with `git diff` from the repository root to keep
paths consistent with `git apply`.

