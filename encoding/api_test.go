package encoding

import "testing"

// TestEncodingAPI runs the Web Platform Tests covering the shared Encoding
// API surface (TextEncoder and TextDecoder together), rather than either
// interface in isolation.
// See https://github.com/web-platform-tests/wpt/tree/master/encoding
func TestEncodingAPI(t *testing.T) {
	t.Parallel()
	base := wptPath("encoding")
	scripts := []testScript{
		{base: base, path: "api-basics.any.js"},
		{base: base, path: "api-surrogates-utf8.any.js"},
	}

	ts := newTestSetup(t)
	err := executeTestScripts(ts, scripts)
	mustNoError(t, err)
}

// TestEncodingAPIInvalidLabel runs the Web Platform Test that verifies
// TextDecoder rejects invalid and whitespace-polluted encoding labels.
func TestEncodingAPIInvalidLabel(t *testing.T) {
	t.Parallel()

	ts := newTestSetup(t)
	err := executeTestScripts(ts, []testScript{
		{base: wptPath("common"), path: "subset-tests.js"},
		{base: wptPath("encoding"), path: "api-invalid-label.any.js"},
	})
	mustNoError(t, err)
}
