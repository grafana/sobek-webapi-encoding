# Sobek Web API: Encoding

`sobek-webapi-encoding` provides the WHATWG Encoding Web API for JavaScript runtimes built with [Sobek](https://github.com/grafana/sobek). Registering the module adds the browser-compatible `TextEncoder` and `TextDecoder` globals to a runtime, including streaming decode, BOM handling, and fatal decoding errors.

## Why it exists

Sobek implements ECMAScript, but browser Web APIs are outside its scope. JavaScript written for browsers and other Web-compatible runtimes nevertheless often expects `TextEncoder` and `TextDecoder` to exist. This module supplies that missing API once, at the Sobek layer, so hosts do not need to maintain their own encoding implementation.

The module was created primarily as the shared upstream implementation for [k6](https://github.com/grafana/k6), whose JavaScript runtime is based on Sobek. It can also be used by any other Go application embedding Sobek that needs Web-compatible conversion between strings and byte arrays.

## Usage

```go
rt := sobek.New()
rt.SetFieldNameMapper(sobek.TagFieldNameMapper("js", true))

if err := sobekencoding.RegisterGlobally(rt); err != nil {
	return err
}
```

JavaScript executed by that runtime can then use the standard API:

```js
const bytes = new TextEncoder().encode("hello");
const text = new TextDecoder("utf-8").decode(bytes);
```

The field-name mapper is required for options such as `fatal`, `ignoreBOM`, and `stream` to be read from JavaScript objects. Hosts that already configure an equivalent mapper may keep their existing configuration.

`TextEncoder` produces UTF-8. `TextDecoder` supports UTF-8, UTF-16LE, and UTF-16BE labels and behavior defined by the [WHATWG Encoding Standard](https://encoding.spec.whatwg.org/).
