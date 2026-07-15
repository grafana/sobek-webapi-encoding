package encoding

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

// TestMain fetches WPT fixtures via scripts/sync-wpt.sh before running any
// tests, if they aren't already present. This keeps `go test ./...` working
// on a fresh checkout without a separate manual sync step.
func TestMain(m *testing.M) {
	if !wptFixturesPresent() {
		if err := runSyncWPT(); err != nil {
			fmt.Fprintln(os.Stderr, "encoding: fetching WPT fixtures:", err)
			os.Exit(1)
		}
	}

	os.Exit(m.Run())
}

// wptFixturesPresent reports whether the WPT directories this package needs
// already exist locally, so only a fresh checkout pays the cost of running
// scripts/sync-wpt.sh.
func wptFixturesPresent() bool {
	for _, dir := range []string{"encoding", "common", "resources"} {
		info, err := os.Stat(wptPath(dir))
		if err != nil || !info.IsDir() {
			return false
		}
	}

	return true
}

func runSyncWPT() error {
	root := computeRepoRoot()

	//nolint:gosec // fixed, repo-relative script path; not influenced by user input.
	cmd := exec.Command("scripts/sync-wpt.sh")
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running scripts/sync-wpt.sh: %w", err)
	}

	return nil
}
