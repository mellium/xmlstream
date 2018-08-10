package xmlstream

import (
	"encoding/xml"
)

// TeeReader returns a Reader that writes to w what it reads from r.
// All reads from r performed through it are matched with
// corresponding writes to w. There is no internal buffering -
// the write must complete before the read completes.
// Any error encountered while writing is reported as a read error.
func TeeReader(r xml.TokenReader, w TokenWriter) xml.TokenReader {
	return teeReader{r, w}
}

type teeReader struct {
	r xml.TokenReader
	w TokenWriter
}

func (t teeReader) Token() (xml.Token, error) {
	tok, err := t.r.Token()
	if tok != nil {
		if err := t.w.EncodeToken(tok); err != nil {
			return tok, err
		}
	}
	return tok, err
}
