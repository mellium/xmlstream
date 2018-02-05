// Copyright 2017 Sam Whited.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package xmlstream

import (
	"encoding/xml"
	"io"
)

type wrapReader struct {
	state int
	start xml.StartElement
	r     TokenReader
}

func (wr *wrapReader) Token() (xml.Token, error) {
	switch wr.state {
	case 0:
		wr.state++
		if wr.r == nil {
			wr.state++
		}
		return wr.start, nil
	case 1:
		t, err := wr.r.Token()
		switch {
		case t != nil && err == io.EOF:
			err = nil
			wr.state++
		case t == nil && err == io.EOF:
			wr.state += 2
			t = wr.start.End()
		}
		return t, err
	case 2:
		wr.state++
		return wr.start.End(), io.EOF
	}
	return nil, io.EOF
}

// Wrap wraps a token stream in a start element and its corresponding end
// element.
func Wrap(r TokenReader, start xml.StartElement) TokenReader {
	return &wrapReader{r: r, start: start}
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

// Inner returns a new TokenReader that returns nil, io.EOF when it consumes the
// end element matching the most recent start element already consumed.
func Inner(r TokenReader) TokenReader {
	count := 1
	return ReaderFunc(func() (xml.Token, error) {
		if count < 1 {
			return nil, io.EOF
		}

		t, err := r.Token()
		if err != nil {
			return nil, err
		}
		switch t.(type) {
		case xml.StartElement:
			count++
		case xml.EndElement:
			count--
			if count < 1 {
				return nil, io.EOF
			}
		}

		return t, err
	})
}
