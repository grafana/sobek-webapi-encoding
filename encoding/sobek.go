package encoding

import (
	"errors"
	"fmt"

	"github.com/grafana/sobek"
)

// RegisterRuntime exports the TextDecoder and TextEncoder constructors into the provided sobek runtime.
func RegisterRuntime(rt *sobek.Runtime) error {
	if err := bindTextDecoder(rt); err != nil {
		return err
	}

	return bindTextEncoder(rt)
}

func bindTextDecoder(rt *sobek.Runtime) error {
	constructor := func(call sobek.ConstructorCall) *sobek.Object {
		label := "utf-8"
		if arg := call.Argument(0); arg != nil && !isNullish(arg) {
			// Per the WebIDL USVString conversion rules, a Symbol cannot be
			// coerced to a string and must throw a TypeError; unlike real
			// string coercion, sobek's ExportTo would otherwise silently
			// convert it to its description instead of erroring.
			if _, isSymbol := arg.(*sobek.Symbol); isSymbol {
				throwAsJSError(rt, NewError(TypeError, "the provided label value cannot be converted to a string"))
			}

			if err := rt.ExportTo(arg, &label); err != nil {
				throwAsJSError(rt, NewError(RangeError, "extracting label from the first argument: "+err.Error()))
			}
		}

		var options TextDecoderOptions
		if arg := call.Argument(1); arg != nil && !isNullish(arg) {
			if err := rt.ExportTo(arg, &options); err != nil {
				panic(rt.NewGoError(err))
			}
		}

		td, err := NewTextDecoder(label, options)
		if err != nil {
			throwAsJSError(rt, err)
		}

		return newTextDecoderObject(rt, td)
	}

	return rt.Set("TextDecoder", constructor)
}

func bindTextEncoder(rt *sobek.Runtime) error {
	constructor := func(_ sobek.ConstructorCall) *sobek.Object {
		return newTextEncoderObject(rt, NewTextEncoder())
	}

	return rt.Set("TextEncoder", constructor)
}

// newTextDecoderObject converts the given TextDecoder instance into a JS object.
//
// It is used by the TextDecoder constructor to convert the Go instance into a JS,
// and will also set the relevant properties as read-only as per the spec.
//
// In the event setting the properties on the object where to fail, the function
// will throw a JS exception.
func newTextDecoderObject(rt *sobek.Runtime, td *TextDecoder) *sobek.Object {
	obj := rt.NewObject()

	// Wrap the Go TextDecoder.Decode method in a JS function
	decodeMethod := func(call sobek.FunctionCall) sobek.Value {
		// Handle variable arguments - buffer can be undefined/missing
		var buffer sobek.Value
		if len(call.Arguments) > 0 {
			buffer = call.Arguments[0]
		}

		// Handle options parameter
		var options TextDecodeOptions
		if len(call.Arguments) > 1 && !isNullish(call.Arguments[1]) {
			if err := rt.ExportTo(call.Arguments[1], &options); err != nil {
				panic(rt.NewGoError(err))
			}
		}

		data, err := exportArrayBuffer(rt, buffer)
		if err != nil {
			panic(rt.NewGoError(err))
		}

		decoded, err := td.Decode(data, options)
		if err != nil {
			throwAsJSError(rt, err)
		}

		return rt.ToValue(decoded)
	}

	// Set the decode method to the wrapper function we just created
	if err := setReadOnlyPropertyOf(obj, "decode", rt.ToValue(decodeMethod)); err != nil {
		panic(rt.NewGoError(fmt.Errorf("defining decode read-only property on TextDecoder object: %w", err)))
	}

	// Set the encoding property
	if err := setReadOnlyPropertyOf(obj, "encoding", rt.ToValue(string(td.Encoding))); err != nil {
		panic(rt.NewGoError(fmt.Errorf("defining encoding read-only property on TextDecoder object: %w", err)))
	}

	// Set the fatal property
	if err := setReadOnlyPropertyOf(obj, "fatal", rt.ToValue(td.Fatal)); err != nil {
		panic(rt.NewGoError(fmt.Errorf("defining fatal read-only property on TextDecoder object: %w", err)))
	}

	// Set the ignoreBOM property
	if err := setReadOnlyPropertyOf(obj, "ignoreBOM", rt.ToValue(td.IgnoreBOM)); err != nil {
		panic(rt.NewGoError(fmt.Errorf("defining ignoreBOM read-only property on TextDecoder object: %w", err)))
	}

	return obj
}

func newTextEncoderObject(rt *sobek.Runtime, te *TextEncoder) *sobek.Object {
	obj := rt.NewObject()

	// Wrap the Go TextEncoder.Encode method in a JS function
	encodeMethod := func(value sobek.Value) *sobek.Object {
		// Handle undefined/null values by defaulting to empty string
		var text string
		if isNullish(value) {
			text = ""
		} else {
			text = value.String()
		}

		buffer, err := te.Encode(text)
		if err != nil {
			throwAsJSError(rt, err)
		}

		// Create a new Uint8Array from the buffer
		u, err := rt.New(rt.Get("Uint8Array"), rt.ToValue(rt.NewArrayBuffer(buffer)))
		if err != nil {
			panic(rt.NewGoError(err))
		}

		return u
	}

	// Set the encode property by wrapping the Go function in a JS function
	if err := setReadOnlyPropertyOf(obj, "encode", rt.ToValue(encodeMethod)); err != nil {
		panic(rt.NewGoError(fmt.Errorf("defining encode read-only method on TextEncoder object: %w", err)))
	}

	// Set the encoding property
	if err := setReadOnlyPropertyOf(obj, "encoding", rt.ToValue(string(te.Encoding))); err != nil {
		panic(rt.NewGoError(fmt.Errorf("defining encoding read-only property on TextEncoder object: %w", err)))
	}

	return obj
}

func throwAsJSError(rt *sobek.Runtime, err error) {
	var encErr *Error
	if errors.As(err, &encErr) {
		panic(encErr.JSError(rt))
	}

	panic(rt.NewGoError(err))
}

// setReadOnlyPropertyOf sets a read-only property on the given [sobek.Object].
func setReadOnlyPropertyOf(obj *sobek.Object, name string, value sobek.Value) error {
	err := obj.DefineDataProperty(name,
		value,
		sobek.FLAG_FALSE,
		sobek.FLAG_FALSE,
		sobek.FLAG_TRUE,
	)
	if err != nil {
		return fmt.Errorf("defining %s read-only property: %w", name, err)
	}

	return nil
}

// exportArrayBuffer interprets the given value as an ArrayBuffer, TypedArray or DataView
// and returns a copy of the underlying byte slice.
func exportArrayBuffer(rt *sobek.Runtime, v sobek.Value) ([]byte, error) {
	if isNullish(v) {
		return []byte{}, nil
	}

	asObject := v.ToObject(rt)

	var ab sobek.ArrayBuffer
	var ok bool

	switch {
	case IsTypedArray(rt, v):
		ab, ok = asObject.Get("buffer").Export().(sobek.ArrayBuffer)
		if !ok {
			return nil, errors.New("typed array buffer is not an ArrayBuffer")
		}

		// Handle TypedArray views that may have byteOffset/byteLength (e.g., subarrays)
		byteOffset := asObject.Get("byteOffset").ToInteger()
		byteLength := asObject.Get("byteLength").ToInteger()

		allBytes := ab.Bytes()
		if byteOffset < 0 || byteOffset > int64(len(allBytes)) {
			return nil, errors.New("typed array byte offset out of bounds")
		}

		end := byteOffset + byteLength
		if end > int64(len(allBytes)) {
			end = int64(len(allBytes))
		}
		return allBytes[byteOffset:end], nil
	case IsInstanceOf(rt, v, DataViewConstructor):
		// Handle DataView objects
		ab, ok = asObject.Get("buffer").Export().(sobek.ArrayBuffer)
		if !ok {
			return nil, errors.New("data view buffer is not an ArrayBuffer")
		}

		// Get the byte offset and length from the DataView
		byteOffset := asObject.Get("byteOffset").ToInteger()
		byteLength := asObject.Get("byteLength").ToInteger()

		// Extract the relevant portion of the ArrayBuffer
		allBytes := ab.Bytes()
		if byteOffset < 0 || byteOffset >= int64(len(allBytes)) {
			return nil, errors.New("data view byte offset out of bounds")
		}

		end := byteOffset + byteLength
		if end > int64(len(allBytes)) {
			end = int64(len(allBytes))
		}
		return allBytes[byteOffset:end], nil
	default:
		ab, ok = asObject.Export().(sobek.ArrayBuffer)
		if !ok {
			return nil, errors.New("data is not an ArrayBuffer, typed array, or DataView")
		}
	}

	return ab.Bytes(), nil
}

// IsInstanceOf returns true if the given value is an instance of the given constructor
// This uses the technique described in https://github.com/dop251/sobek/issues/379#issuecomment-1164441879
func IsInstanceOf(rt *sobek.Runtime, v sobek.Value, instanceOf ...JSType) bool {
	var valid bool

	for _, t := range instanceOf {
		instanceOfConstructor := rt.Get(string(t))
		if valid = v.ToObject(rt).Get("constructor").SameAs(instanceOfConstructor); valid {
			break
		}
	}

	return valid
}

// IsTypedArray returns true if the given value is an instance of a Typed Array
func IsTypedArray(rt *sobek.Runtime, v sobek.Value) bool {
	asObject := v.ToObject(rt)

	typedArrayTypes := []JSType{
		Int8ArrayConstructor,
		Uint8ArrayConstructor,
		Uint8ClampedArrayConstructor,
		Int16ArrayConstructor,
		Uint16ArrayConstructor,
		Int32ArrayConstructor,
		Uint32ArrayConstructor,
		Float32ArrayConstructor,
		Float64ArrayConstructor,
		BigInt64ArrayConstructor,
		BigUint64ArrayConstructor,
	}

	return IsInstanceOf(rt, asObject, typedArrayTypes...)
}

// JSType is a string representing a JavaScript type
type JSType string

const (
	// DataViewConstructor is the name of the DataView constructor
	DataViewConstructor = "DataView"

	// Int8ArrayConstructor is the name of the Int8ArrayConstructor constructor
	Int8ArrayConstructor = "Int8Array"

	// Uint8ArrayConstructor is the name of the Uint8ArrayConstructor constructor
	Uint8ArrayConstructor = "Uint8Array"

	// Uint8ClampedArrayConstructor is the name of the Uint8ClampedArrayConstructor constructor
	Uint8ClampedArrayConstructor = "Uint8ClampedArray"

	// Int16ArrayConstructor is the name of the Int16ArrayConstructor constructor
	Int16ArrayConstructor = "Int16Array"

	// Uint16ArrayConstructor is the name of the Uint16ArrayConstructor constructor
	Uint16ArrayConstructor = "Uint16Array"

	// Int32ArrayConstructor is the name of the Int32ArrayConstructor constructor
	Int32ArrayConstructor = "Int32Array"

	// Uint32ArrayConstructor is the name of the Uint32ArrayConstructor constructor
	Uint32ArrayConstructor = "Uint32Array"

	// Float32ArrayConstructor is the name of the Float32ArrayConstructor constructor
	Float32ArrayConstructor = "Float32Array"

	// Float64ArrayConstructor is the name of the Float64ArrayConstructor constructor
	Float64ArrayConstructor = "Float64Array"

	// BigInt64ArrayConstructor is the name of the BigInt64ArrayConstructor constructor
	BigInt64ArrayConstructor = "BigInt64Array"

	// BigUint64ArrayConstructor is the name of the BigUint64ArrayConstructor constructor
	BigUint64ArrayConstructor = "BigUint64Array"
)

func isNullish(v sobek.Value) bool {
	return v == nil || sobek.IsUndefined(v) || sobek.IsNull(v)
}
