package encoding

import (
	"testing"
)

// TestTextEncoder runs the Web Platform Tests for TextEncoder.
// See https://github.com/web-platform-tests/wpt/tree/master/encoding
func TestTextEncoder(t *testing.T) {
	t.Parallel()
	base := wptPath("encoding")
	scripts := []testScript{
		{base: base, path: "textencoder-constructor-non-utf.any.js"},
		{base: base, path: "textencoder-utf16-surrogates.any.js"},
	}

	ts := newTestSetup(t)
	err := executeTestScripts(ts, scripts)
	mustNoError(t, err)
}
