// Copyright 2017 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package xmlstream

import (
	"encoding/xml"
	"io"
)

type token struct {
	tok xml.Token
}

func (t token) Token() (xml.Token, error) {
	return t.tok, io.EOF
}

// Token returns a token reader that always returns the given token and io.EOF.
func Token(t xml.Token) xml.TokenReader {
	return token{tok: t}
}

// Wrap wraps a token stream in a start element and its corresponding end
// element.
func Wrap(r xml.TokenReader, start xml.StartElement) xml.TokenReader {
	if r == nil {
		return &multiReader{readers: []xml.TokenReader{Token(start), Token(start.End())}}
	}
	return &multiReader{readers: []xml.TokenReader{Token(start), r, Token(start.End())}}
}

// Unwrap reads the next token from the provided TokenReader and, if it is a
// start element, returns a new TokenReader that skips the corresponding end
// element. If the token is not a start element the original TokenReader is
// returned.
func Unwrap(r xml.TokenReader) (xml.TokenReader, xml.Token, error) {
	t, err := r.Token()
	if err != nil {
		return r, t, err
	}
	_, ok := t.(xml.StartElement)
	if !ok {
		return r, t, err
	}

	return MultiReader(Inner(r), r), t, nil
}

// Inner returns a new TokenReader that returns nil, io.EOF when it consumes the
// end element matching the most recent start element already consumed.
func Inner(r xml.TokenReader) xml.TokenReader {
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
