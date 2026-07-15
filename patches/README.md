# WPT Patch Overrides

Store patch files here when upstream WPT tests require adjustments to run
inside Sobek. Add the patch path to the corresponding entry in `wpt.json`
and run `scripts/sync-wpt.sh` to re-apply the changes after syncing new
tests. Requires `curl`, `jq`, and `git` (the last only for entries with a
`patch` field).

Patches should be created with `git diff` from the repository root to keep
paths consistent with `git apply`.

