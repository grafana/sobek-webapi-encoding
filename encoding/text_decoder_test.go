package encoding

import (
	"strings"
	"testing"

	"golang.org/x/text/encoding/unicode"
)

// TestTextDecoder runs the Web Platform Tests for TextDecoder.
// See https://github.com/web-platform-tests/wpt/tree/master/encoding
func TestTextDecoder(t *testing.T) {
	t.Parallel()
	base := wptPath("encoding")
	scripts := []testScript{
		{base: base, path: "textdecoder-arguments.js"},
		{base: base, path: "textdecoder-byte-order-marks.js"},
		{base: base, path: "textdecoder-copy.js"},
		{base: base, path: "textdecoder-eof.js"},
		{base: base, path: "textdecoder-fatal.js"},
		{base: base, path: "textdecoder-fatal-streaming.js"},
		{base: base, path: "textdecoder-ignorebom.js"},
		{base: base, path: "textdecoder-labels.js"},
		{base: base, path: "textdecoder-streaming.js"},
		{base: base, path: "textdecoder-utf16-surrogates.js"},
	}

	ts := newTestSetup(t)
	err := executeTestScripts(ts, scripts)
	mustNoError(t, err)
}

func TestTextDecoderUTF8StreamingStateMachine(t *testing.T) {
	t.Parallel()

	t.Run("IncompleteThenInvalidContinuation", func(t *testing.T) {
		t.Parallel()
		td := &TextDecoder{
			TextDecoderCommon: TextDecoderCommon{
				Encoding: UTF8EncodingFormat,
			},
			decoder: unicode.UTF8,
		}

		out, err := td.Decode([]byte{0xF0, 0x9F}, TextDecodeOptions{Stream: true})
		mustNoError(t, err)
		mustEqual(t, "", out)

		out, err = td.Decode([]byte{0x41}, TextDecodeOptions{Stream: true})
		mustNoError(t, err)
		mustEqual(t, "\uFFFDA", out)

		out, err = td.Decode(nil, TextDecodeOptions{})
		mustNoError(t, err)
		mustEqual(t, "", out)
	})

	t.Run("ImmediateInvalidStartByte", func(t *testing.T) {
		t.Parallel()
		td := &TextDecoder{
			TextDecoderCommon: TextDecoderCommon{
				Encoding: UTF8EncodingFormat,
			},
			decoder: unicode.UTF8,
		}

		out, err := td.Decode([]byte{0xC1}, TextDecodeOptions{Stream: true})
		mustNoError(t, err)
		mustEqual(t, "\uFFFD", out)

		out, err = td.Decode(nil, TextDecodeOptions{})
		mustNoError(t, err)
		mustEqual(t, "", out)
	})

	t.Run("ASCIIStreamingProducesOutput", func(t *testing.T) {
		t.Parallel()
		td := &TextDecoder{
			TextDecoderCommon: TextDecoderCommon{
				Encoding: UTF8EncodingFormat,
			},
			decoder: unicode.UTF8,
		}

		out, err := td.Decode([]byte("A"), TextDecodeOptions{Stream: true})
		mustNoError(t, err)
		mustEqual(t, "A", out)

		out, err = td.Decode([]byte("B"), TextDecodeOptions{Stream: true})
		mustNoError(t, err)
		mustEqual(t, "B", out)

		out, err = td.Decode(nil, TextDecodeOptions{})
		mustNoError(t, err)
		mustEqual(t, "", out)
	})

	t.Run("FatalFlushOnTruncatedSequence", func(t *testing.T) {
		t.Parallel()
		td := &TextDecoder{
			TextDecoderCommon: TextDecoderCommon{
				Encoding: UTF8EncodingFormat,
				Fatal:    true,
			},
			decoder: unicode.UTF8,
		}

		_, err := td.Decode([]byte{0xF0, 0x9F}, TextDecodeOptions{})
		mustError(t, err)
	})
}

func TestTextDecoderUTF16FatalStreaming(t *testing.T) {
	t.Parallel()

	newFatalDecoder := func() *TextDecoder {
		return &TextDecoder{
			TextDecoderCommon: TextDecoderCommon{
				Encoding: UTF16LEEncodingFormat,
				Fatal:    true,
			},
			decoder: unicode.UTF16(unicode.LittleEndian, unicode.UseBOM),
		}
	}

	t.Run("OddFollowedByOddCompletes", func(t *testing.T) {
		t.Parallel()
		td := newFatalDecoder()

		out, err := td.Decode([]byte{0x00}, TextDecodeOptions{Stream: true})
		mustNoError(t, err)
		mustEqual(t, "", out)

		out, err = td.Decode([]byte{0x00}, TextDecodeOptions{})
		mustNoError(t, err)
		mustEqual(t, "\u0000", out)
	})

	t.Run("EvenThenOddThrows", func(t *testing.T) {
		t.Parallel()
		td := newFatalDecoder()

		out, err := td.Decode([]byte{0x00, 0x00}, TextDecodeOptions{Stream: true})
		mustNoError(t, err)
		mustEqual(t, "\u0000", out)

		_, err = td.Decode([]byte{0x00}, TextDecodeOptions{})
		mustError(t, err)
	})

	t.Run("OddThenEvenThrows", func(t *testing.T) {
		t.Parallel()
		td := newFatalDecoder()

		out, err := td.Decode([]byte{0x00}, TextDecodeOptions{Stream: true})
		mustNoError(t, err)
		mustEqual(t, "", out)

		_, err = td.Decode([]byte{0x00, 0x00}, TextDecodeOptions{})
		mustError(t, err)
	})

	t.Run("EvenChunksStreamSuccessfully", func(t *testing.T) {
		t.Parallel()
		td := newFatalDecoder()

		out, err := td.Decode([]byte{0x00, 0x00}, TextDecodeOptions{Stream: true})
		mustNoError(t, err)
		mustEqual(t, "\u0000", out)

		out, err = td.Decode([]byte{0x00, 0x00}, TextDecodeOptions{})
		mustNoError(t, err)
		mustEqual(t, "\u0000", out)
	})
}

func TestTextDecoderUTF16LEStreamingSingleByteWindow(t *testing.T) {
	t.Parallel()

	encoded := []byte{
		0x00, 0x00, 0x31, 0x00, 0x32, 0x00, 0x33, 0x00,
		0x41, 0x00, 0x42, 0x00, 0x43, 0x00, 0x61, 0x00,
		0x62, 0x00, 0x63, 0x00, 0x80, 0x00, 0xFF, 0x00,
		0x00, 0x01, 0x00, 0x10, 0xFD, 0xFF, 0x00, 0xD8,
		0x00, 0xDC, 0xFF, 0xDB, 0xFF, 0xDF,
	}
	expected := "\x00123ABCabc\u0080\u00FF\u0100\u1000\uFFFD\U00010000\U0010FFFF"

	td := &TextDecoder{
		TextDecoderCommon: TextDecoderCommon{
			Encoding: UTF16LEEncodingFormat,
		},
		decoder: unicode.UTF16(unicode.LittleEndian, unicode.UseBOM),
	}

	var out strings.Builder
	for _, b := range encoded {
		chunk := []byte{b}
		part, err := td.Decode(chunk, TextDecodeOptions{Stream: true})
		mustNoError(t, err)
		out.WriteString(part)
	}

	part, err := td.Decode(nil, TextDecodeOptions{})
	mustNoError(t, err)
	out.WriteString(part)

	mustEqual(t, expected, out.String())
}

// TestTextDecoderDecodeZeroLengthDataViewAtBufferEnd guards against an
// off-by-one in exportArrayBuffer's DataView bounds check that rejected a
// valid zero-length DataView placed exactly at the end of its buffer.
func TestTextDecoderDecodeZeroLengthDataViewAtBufferEnd(t *testing.T) {
	t.Parallel()

	ts := newTestSetup(t)

	v, err := ts.rt.RunScript("test.js", `
		const decoder = new TextDecoder();
		decoder.decode(new DataView(new ArrayBuffer(4), 4, 0));
	`)
	mustNoError(t, err)
	mustEqual(t, "", v.String())
}

// TestTextDecoderConstructorSymbolLabelThrowsTypeError guards against
// the TextDecoder constructor throwing a RangeError for a Symbol label,
// which per WebIDL USVString conversion rules should be a TypeError since a
// Symbol cannot be coerced to a string at all.
func TestTextDecoderConstructorSymbolLabelThrowsTypeError(t *testing.T) {
	t.Parallel()

	ts := newTestSetup(t)

	_, err := ts.rt.RunScript("test.js", `
		let threw;
		try {
			new TextDecoder(Symbol("x"));
		} catch (e) {
			threw = e;
		}
		if (!(threw instanceof TypeError)) {
			throw new Error("expected a TypeError, got " + threw);
		}
	`)
	mustNoError(t, err)
}
