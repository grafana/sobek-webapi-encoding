package encoding

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/grafana/sobek"
)

// testScript is a helper struct holding the base path
// and the path of a test script.
type testScript struct {
	base string
	path string
}

// testSetup wraps a sobek runtime configured with the encoding Web API.
type testSetup struct {
	rt *sobek.Runtime
}

func computeRepoRoot() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("unable to determine repository root from runtime caller data")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(filename), ".."))
}

func wptPath(parts ...string) string {
	base := filepath.Join(computeRepoRoot(), "wpt")
	return filepath.Join(append([]string{base}, parts...)...)
}

func newTestSetup(t testing.TB) *testSetup {
	t.Helper()

	rt := sobek.New()
	rt.SetFieldNameMapper(sobek.TagFieldNameMapper("json", true))

	mustNoError(t, RegisterRuntime(rt))

	ts := &testSetup{rt: rt}
	mustNoError(t, testExecuteTestScripts(ts))
	return ts
}

func mustNoError(t testing.TB, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func mustError(t testing.TB, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("expected error")
	}
}

func mustEqual[T comparable](t testing.TB, expected, actual T) {
	t.Helper()

	if expected != actual {
		t.Fatalf("expected %#v, got %#v", expected, actual)
	}
}

func testExecuteTestScripts(ts *testSetup) error {
	scripts := []testScript{
		{base: wptPath("resources"), path: "testharness.js"},
		{base: wptPath("encoding", "resources"), path: "encodings.js"},
		{base: wptPath("common"), path: "sab.js"},
	}

	return executeTestScripts(ts, scripts)
}

func executeTestScripts(ts *testSetup, scripts []testScript) error {
	for _, script := range scripts {
		fullPath := filepath.Join(script.base, script.path)
		// #nosec G304 -- WPT test files are part of the repository and not user-supplied.
		//nolint:forbidigo // os.ReadFile is acceptable for locally vendored fixtures.
		contents, err := os.ReadFile(fullPath)
		if err != nil {
			return err
		}

		if _, err = ts.rt.RunScript(script.path, string(contents)); err != nil {
			return err
		}
	}

	return nil
}
