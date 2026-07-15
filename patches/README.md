# WPT Patch Overrides

Store patch files here when upstream WPT tests require adjustments to run
inside Sobek. Fixtures are fetched and patched by
[`wptsync`](https://github.com/oleiade/wptsync), which also runs
automatically before this package's tests via `TestMain` -- plain
`go test ./...` is enough to get up-to-date, patched fixtures.

Install the CLI:

```bash
go install github.com/oleiade/wptsync/cmd/wptsync@latest
```

Workflow for editing a synced file's patch:

1. `wptsync edit <path>` restores the file to its pristine + patched state
   before you start editing.
2. Edit the file under `wpt/` directly.
3. `wptsync save <path>` diffs your edits against the pristine file and
   writes the result to the file's patch, registering it in `wpt.json` if
   it is new.

Run `wptsync sync` to fetch fixtures on demand (this is what `TestMain` does
under the hood), and `wptsync update` to bump the pinned commit and
re-sync, reporting any patches that no longer apply.

Patches are standard `git apply` format, so you can still craft or adjust
them by hand if you prefer.

