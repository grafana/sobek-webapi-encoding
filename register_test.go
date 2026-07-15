package sobekencoding

import (
	"testing"

	"github.com/grafana/sobek"
)

// TestRegisterGlobally exercises the module's actual documented entry point,
// RegisterGlobally, end to end against a plain sobek runtime -- rather than
// relying solely on encoding.RegisterRuntime being covered indirectly via
// the encoding package's own test harness.
func TestRegisterGlobally(t *testing.T) {
	t.Parallel()

	rt := sobek.New()
	rt.SetFieldNameMapper(sobek.TagFieldNameMapper("json", true))

	if err := RegisterGlobally(rt); err != nil {
		t.Fatalf("RegisterGlobally returned an unexpected error: %v", err)
	}

	v, err := rt.RunString(`
		const encoder = new TextEncoder();
		const encoded = encoder.encode("Hello, World!");

		const decoder = new TextDecoder("utf-8");
		const decoded = decoder.decode(encoded);

		if (decoded !== "Hello, World!") {
			throw new Error("unexpected decoded value: " + decoded);
		}

		decoded;
	`)
	if err != nil {
		t.Fatalf("unexpected error running script: %v", err)
	}

	if got := v.String(); got != "Hello, World!" {
		t.Fatalf("expected %q, got %q", "Hello, World!", got)
	}
}

// TestRegisterGloballyFatalOption guards against RegisterGlobally leaving
// TextDecoder options such as "fatal" inert when the caller has configured a
// field name mapper as documented.
func TestRegisterGloballyFatalOption(t *testing.T) {
	t.Parallel()

	rt := sobek.New()
	rt.SetFieldNameMapper(sobek.TagFieldNameMapper("json", true))

	if err := RegisterGlobally(rt); err != nil {
		t.Fatalf("RegisterGlobally returned an unexpected error: %v", err)
	}

	_, err := rt.RunString(`
		const decoder = new TextDecoder("utf-8", { fatal: true });
		if (decoder.fatal !== true) {
			throw new Error("expected decoder.fatal to be true");
		}

		let threw = false;
		try {
			decoder.decode(new Uint8Array([0xFF]));
		} catch (e) {
			threw = e instanceof TypeError;
		}
		if (!threw) {
			throw new Error("expected decode() to throw a TypeError for invalid input");
		}
	`)
	if err != nil {
		t.Fatalf("unexpected error running script: %v", err)
	}
}
