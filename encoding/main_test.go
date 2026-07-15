package encoding

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestMain fetches WPT fixtures via scripts/sync-wpt.sh before running any
// tests, if they aren't already present. This keeps `go test ./...` working
// on a fresh checkout without a separate manual sync step.
//
//nolint:forbidigo // TestMain needs os.Exit to return the package test status.
func TestMain(m *testing.M) {
	if !wptFixturesPresent() {
		if err := runSyncWPT(); err != nil {
			log.Printf("encoding: fetching WPT fixtures: %v", err)
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
		matches, err := filepath.Glob(wptPath(dir, "*"))
		if err != nil || len(matches) == 0 {
			return false
		}
	}

	return true
}

func runSyncWPT() error {
	root := computeRepoRoot()

	// Invoke the script through bash explicitly rather than executing it
	// directly: Windows has no shebang support, so running the file itself
	// fails there even though Git for Windows' bash.exe (which GitHub's
	// windows-latest runners provide) is on PATH.
	cmd := exec.CommandContext(context.Background(), "bash", "scripts/sync-wpt.sh")
	cmd.Dir = root

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running scripts/sync-wpt.sh: %w", err)
	}

	return nil
}
