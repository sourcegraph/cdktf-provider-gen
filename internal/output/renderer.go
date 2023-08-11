package output

import (
	"io"

	liboutput "github.com/sourcegraph/sourcegraph/lib/output"
)

type DiffRenderer string

var _ Renderer = DiffRenderer("")

func (d DiffRenderer) Render(w io.Writer, format Format) error {
	switch format {
	case FormatPretty:
		return liboutput.NewOutput(w, liboutput.OutputOpts{}).
			WriteCode("diff", string(d))

	default:
		return ErrFormatUnimplemented
	}
}
