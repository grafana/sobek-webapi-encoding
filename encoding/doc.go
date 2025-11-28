// Package encoding implements the WHATWG Encoding Standard for Sobek runtimes.
//
// This package provides TextEncoder and TextDecoder implementations that can be
// registered with a Sobek runtime to provide Web-compatible text encoding/decoding
// capabilities in JavaScript.
//
// # Supported Encodings
//
// The package supports the following encodings:
//   - UTF-8 (default)
//   - UTF-16LE (little-endian)
//   - UTF-16BE (big-endian)
//
// # Usage
//
// To register the encoding constructors with a Sobek runtime:
//
//	rt := sobek.New()
//	if err := encoding.RegisterRuntime(rt); err != nil {
//	    log.Fatal(err)
//	}
//
// After registration, TextEncoder and TextDecoder are available in JavaScript:
//
//	const encoder = new TextEncoder();
//	const encoded = encoder.encode("Hello, World!");
//
//	const decoder = new TextDecoder("utf-8");
//	const decoded = decoder.decode(encoded);
//
// # Specification
//
// This implementation follows the WHATWG Encoding Standard:
// https://encoding.spec.whatwg.org/
package encoding
