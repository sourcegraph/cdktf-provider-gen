package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	liboutput "github.com/sourcegraph/sourcegraph/lib/output"
)

type Format string

const (
	// FormatJSON renders plain, unformatted JSON.
	FormatJSON Format = "json"
	// FormatPretty renders pretty, human-readable content, by default indented,
	// color-coded JSON.
	FormatPretty Format = "pretty"
	// FormatText renders plain-text content, by default the '%v' directive formatting.
	FormatText Format = "text"
	// FormatNone renders nothing.
	FormatNone Format = "none"
)

var (
	// Formats is a slice of all supported default formats.
	Formats = []Format{
		FormatJSON,
		FormatPretty,
		FormatText,
		FormatNone,
	}
)

// ErrFormatUnimplemented can be used by Renderer implementations for
// unhandled format cases.
var ErrFormatUnimplemented = errors.New("custom output.Renderer does not implement this format")

// Renderer is an interface that can be implemented by types that want to
// support custom render formats.
//
// The implementation should return ErrFormatUnimplemented as the fallback
// case - this tells the top-level output.Render implementation to use its
// default behaviour if the type does not implement it.
type Renderer interface{ Render(io.Writer, Format) error }

// Render outputs v in the requested format. For custom formatting, v may
// implement output.Renderer to the default override formatting behaviour.
func Render(format Format, v any) error {
	if r, ok := (v).(Renderer); ok {
		err := r.Render(os.Stdout, format)
		// If no issue occurred, we are done.
		if err == nil {
			return nil
		}
		// If unimplemented, continue with default behaviour - otherwise
		// the custom render implementation has failed.
		if !errors.Is(err, ErrFormatUnimplemented) {
			return err
		}
	}

	switch format {
	case FormatJSON:
		return json.NewEncoder(os.Stdout).Encode(v)
	case FormatPretty:
		data, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return err
		}
		return liboutput.NewOutput(os.Stdout, liboutput.OutputOpts{}).
			WriteCode("json", string(data))
	case FormatText:
		data := strings.TrimSpace(fmt.Sprintf("%v", v))
		_, err := fmt.Println(data)
		return err
	case FormatNone:
		return nil
	default:
		return fmt.Errorf("unknown format %q", format)
	}
}
