// Copyright 2017 Sam Whited.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package xmlstream

import (
	"encoding/xml"
	"io"
)

// Wrap wraps a token stream in a start element and its corresponding end
// element.
func Wrap(r TokenReader, start xml.StartElement) TokenReader {
	state := 0
	return ReaderFunc(func() (t xml.Token, err error) {
		switch state {
		case 0:
			state++
			return start, nil
		case 1:
			t, err = r.Token()
			if err == io.EOF {
				state++
				err = nil
			}
			if t == nil {
				state++
				t = start.End()
			}
			return t, err
		case 2:
			state++
			return start.End(), io.EOF
		default:
			return nil, io.EOF
		}
	})
}

// Unwrap reads the next token from the provided TokenReader and, if it is a
// start element, returns a new TokenReader that skips the corresponding end
// element. If the token is not a start element the original TokenReader is
// returned.
func Unwrap(r TokenReader) (TokenReader, xml.Token, error) {
	t, err := r.Token()
	if err != nil {
		return r, t, err
	}
	start, ok := t.(xml.StartElement)
	if !ok {
		return r, t, err
	}

	depth := 0
	return ReaderFunc(func() (t xml.Token, err error) {
		t, err = r.Token()
		switch tok := t.(type) {
		case xml.StartElement:
			if tok.Name == start.Name {
				depth++
			}
		case xml.EndElement:
			if tok.Name == start.Name {
				depth--
				if depth == -1 {
					t, err = r.Token()
				}
			}
		}

		return t, err
	}), t, err
}
