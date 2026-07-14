package encoding

import (
	"errors"
	"fmt"
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

// wptRuntimeCompatPolyfill provides runtime globals the vendored WPT test
// harness expects but sobek does not implement, so the vendored files
// themselves can stay byte-for-byte identical to upstream.
const wptRuntimeCompatPolyfill = `
(function () {
	// testharness.js's outer IIFE is invoked with the browser/worker "self"
	// global (see the last line of the file); sobek has no such global.
	if (typeof globalThis.self === "undefined") {
		globalThis.self = globalThis;
	}

	// common/subset-tests.js reads location.search to support splitting a
	// test file into variants; sobek has no location global and none of the
	// vendored tests use variants, so an empty search string is sufficient.
	if (typeof globalThis.location === "undefined") {
		globalThis.location = {search: ""};
	}

	// common/sab.js detects SharedArrayBuffer support via
	// "new WebAssembly.Memory({shared: true, ...}).buffer.constructor".
	// sobek implements neither WebAssembly nor SharedArrayBuffer, so
	// sab.js's own try/catch would fall back to sabConstructor = null and
	// every createBuffer("SharedArrayBuffer", ...) call would throw.
	// Provide a SharedArrayBuffer-named ArrayBuffer subclass and a minimal
	// WebAssembly.Memory shim so sab.js's unmodified feature detection
	// succeeds on its own.
	if (typeof globalThis.SharedArrayBuffer === "undefined") {
		globalThis.SharedArrayBuffer = class SharedArrayBuffer extends ArrayBuffer {};
	}
	if (typeof globalThis.WebAssembly === "undefined") {
		globalThis.WebAssembly = {};
	}
	if (typeof globalThis.WebAssembly.Memory === "undefined") {
		globalThis.WebAssembly.Memory = function Memory(descriptor) {
			var shared = !!(descriptor && descriptor.shared);
			this.buffer = shared ? new globalThis.SharedArrayBuffer(0) : new ArrayBuffer(0);
		};
	}
})();
`

// wptTestOverridePolyfill replaces testharness.js's test() with a variant
// that throws synchronously on assertion failures, once testharness.js has
// defined and exposed the original via expose(test, 'test'). Upstream
// test() records subtest results on an internal harness object instead of
// throwing, which would hide failed assertions from this Go-driven runner.
const wptTestOverridePolyfill = `
(function () {
	globalThis.test = function (func, name) {
		try {
			func();
		} catch (e) {
			if (name) {
				throw "Test \"" + name + "\" failed: " + e;
			}
			throw e;
		}
	};
})();
`

// wptEncodingsFilterPolyfill prunes encodings.js's encodings_table down to
// the encodings this package supports, once encodings.js has populated it.
// Without this, WPT's encoding test files iterate every WHATWG-registered
// encoding and fail on every one this package doesn't implement.
const wptEncodingsFilterPolyfill = `
(function () {
	var supported = new Set(["UTF-8", "UTF-16BE", "UTF-16LE"]);
	var filtered = [];

	encodings_table.forEach(function (section) {
		var encodings = section.encodings.filter(function (encoding) {
			return supported.has(encoding.name);
		});

		if (encodings.length === 0) {
			return;
		}

		filtered.push({heading: section.heading, encodings: encodings});
	});

	encodings_table.length = 0;
	Array.prototype.push.apply(encodings_table, filtered);
})();
`

func testExecuteTestScripts(ts *testSetup) error {
	if err := executeInlineScript(ts, "wpt-runtime-compat.js", wptRuntimeCompatPolyfill); err != nil {
		return err
	}

	if err := executeTestScripts(ts, []testScript{{base: wptPath("resources"), path: "testharness.js"}}); err != nil {
		return err
	}

	if err := executeInlineScript(ts, "wpt-test-override.js", wptTestOverridePolyfill); err != nil {
		return err
	}

	if err := executeTestScripts(ts, []testScript{{base: wptPath("encoding", "resources"), path: "encodings.js"}}); err != nil {
		return err
	}

	if err := executeInlineScript(ts, "wpt-encodings-filter.js", wptEncodingsFilterPolyfill); err != nil {
		return err
	}

	return executeTestScripts(ts, []testScript{{base: wptPath("common"), path: "sab.js"}})
}

func executeInlineScript(ts *testSetup, name, source string) error {
	_, err := ts.rt.RunScript(name, source)
	return err
}

func executeTestScripts(ts *testSetup, scripts []testScript) error {
	for _, script := range scripts {
		fullPath := filepath.Join(script.base, script.path)
		// #nosec G304 -- WPT test files are part of the repository and not user-supplied.
		//nolint:forbidigo // os.ReadFile is acceptable for locally vendored fixtures.
		contents, err := os.ReadFile(fullPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) { //nolint:forbidigo // checking for a not-exist error, not calling an os I/O function.
				return fmt.Errorf("WPT fixture %s not found; run scripts/sync-wpt.sh to fetch WPT test fixtures before running tests: %w", fullPath, err)
			}
			return err
		}

		if _, err = ts.rt.RunScript(script.path, string(contents)); err != nil {
			return err
		}
	}

	return nil
}
