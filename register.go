// Package sobekencoding registers the encoding Web API with Sobek runtimes.
package sobekencoding

import (
	"github.com/grafana/sobek"

	"github.com/grafana/sobek-webapi-encoding/encoding"
)

// RegisterGlobally exposes the encoding TextDecoder/TextEncoder constructors in the provided sobek runtime.
func RegisterGlobally(rt *sobek.Runtime) error {
	return encoding.RegisterRuntime(rt)
}
