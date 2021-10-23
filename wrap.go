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

func (t *token) Token() (xml.Token, error) {
	tok := t.tok
	t.tok = nil
	return tok, io.EOF
}

// Token returns a reader that returns the given token and io.EOF, then nil
// io.EOF thereafter.
func Token(t xml.Token) xml.TokenReader {
	return &token{tok: t}
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
	return innerElement(r, false)
}

// InnerElement wraps a TokenReader to return nil, io.EOF after returning the
// end element matching the most recent start element already consumed.
// It is like Inner except that it returns the end element.
func InnerElement(r xml.TokenReader) xml.TokenReader {
	return innerElement(r, true)
}

func innerElement(r xml.TokenReader, returnOuter bool) xml.TokenReader {
	var count int
	return ReaderFunc(func() (xml.Token, error) {
		if count < 0 {
			return nil, io.EOF
		}

		t, err := r.Token()
		switch t.(type) {
		case xml.StartElement:
			count++
		case xml.EndElement:
			count--
			if !returnOuter && count < 0 {
				return nil, io.EOF
			}
		}

		return t, err
	})
}

// Skip reads tokens until it has consumed the end element matching the most
// recent start element already consumed.
// It recurs if it encounters a start element, so it can be used to skip nested
// structures.
// It returns nil if it finds an end element at the same nesting level as the
// start element; otherwise it returns an error describing the problem.
// Skip does not verify that the start and end elements match.
func Skip(r xml.TokenReader) error {
	for {
		tok, err := r.Token()
		if err != nil {
			return err
		}
		switch tok.(type) {
		case xml.StartElement:
			if err := Skip(r); err != nil {
				return err
			}
		case xml.EndElement:
			return nil
		}
	}
}

// LimitReader returns a xml.TokenReader that reads from r but stops with EOF
// after n tokens (regardless of the validity of the XML at that point in the
// stream).
func LimitReader(r xml.TokenReader, n uint) xml.TokenReader {
	return ReaderFunc(func() (xml.Token, error) {
		if n <= 0 {
			return nil, io.EOF
		}
		n--
		return r.Token()
	})
}
