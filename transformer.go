// Copyright 2016 Sam Whited.
// Use of this source code is governed by the BSD 2-clause license that can be
// found in the LICENSE file.

package xmlstream

import (
	"encoding/xml"
	"io"
)

// TokenWriter is anything that can encode tokens to an XML stream, including an
// xml.Encoder.
type TokenWriter interface {
	EncodeToken(t xml.Token) error
	Flush() error
}

// A Transformer returns a new xml.Decoder that returns transformed tokens read
// from src.
type Transformer func(src xml.TokenReader) *xml.Decoder

// Encode consumes an xml.TokenReader and encodes any tokens that it outputs.
// If an error is returned on the Decode or Encode side, it is returned
// immediately.
// Since Encode is defined as consuming the stream until the end, io.EOF is not
// returned.
// If no error would be returned, Encode flushes the TokenWriter when it is
// done.
func Encode(e TokenWriter, d xml.TokenReader) (err error) {
	defer func() {
		if err == nil || err == io.EOF {
			err = e.Flush()
		}
	}()

	var tok xml.Token
	for {
		tok, err = d.Token()
		if err != nil {
			return err
		}

		if err = e.EncodeToken(tok); err != nil {
			return err
		}
	}
}

// Inspect performs an operation for each token in the stream without
// transforming the stream in any way.
func Inspect(f func(t xml.Token)) Transformer {
	return func(src xml.TokenReader) *xml.Decoder {
		return xml.NewTokenDecoder(inspector{
			d: src,
			f: f,
		})
	}
}

type inspector struct {
	d xml.TokenReader
	f func(t xml.Token)
}

func (t inspector) Token() (xml.Token, error) {
	tok, err := t.d.Token()
	if err != nil {
		return nil, err
	}
	t.f(tok)
	return tok, nil
}

// Map returns a Transformer that maps the tokens in the input using the given
// mapping.
func Map(mapping func(t xml.Token) xml.Token) Transformer {
	return func(src xml.TokenReader) *xml.Decoder {
		return xml.NewTokenDecoder(&mapper{
			d: src,
			f: mapping,
		})
	}
}

type mapper struct {
	d xml.TokenReader
	f func(t xml.Token) xml.Token
}

func (m *mapper) Token() (xml.Token, error) {
	tok, err := m.d.Token()
	if err != nil {
		return nil, err
	}
	return m.f(tok), nil
}

// Remove returns a Transformer that removes tokens for which f matches.
func Remove(f func(t xml.Token) bool) Transformer {
	return func(src xml.TokenReader) *xml.Decoder {
		return xml.NewTokenDecoder(remover{
			d: src,
			f: f,
		})
	}
}

type remover struct {
	d xml.TokenReader
	f func(t xml.Token) bool
}

func (r remover) Token() (t xml.Token, err error) {
	for {
		t, err = r.d.Token()
		switch {
		case err != nil:
			return nil, err
		case r.f(t):
			continue
		}
		return
	}
}

// RemoveElement returns a Transformer that removes entire elements (and their
// children) if f matches the elements start token.
func RemoveElement(f func(start xml.StartElement) bool) Transformer {
	return func(src xml.TokenReader) *xml.Decoder {
		return xml.NewTokenDecoder(&elementremover{
			d: xml.NewTokenDecoder(src),
			f: f,
		})
	}
}

type elementremover struct {
	d *xml.Decoder
	f func(start xml.StartElement) bool
}

func (er *elementremover) Token() (t xml.Token, err error) {
	for {
		t, err = er.d.Token()
		if err != nil {
			return nil, err
		}
		if start, ok := t.(xml.StartElement); ok && er.f(start) {
			// Skip the element and read a new token.
			if err = er.d.Skip(); err != nil {
				return t, err
			}
			continue
		}

		return
	}
}

// BUG(ssw): Multiple uses of RemoveAttr will iterate over the attr list
//           multiple times.

// RemoveAttr returns a Transformer that removes attributes from
// xml.StartElement's if f matches.
func RemoveAttr(f func(start xml.StartElement, attr xml.Attr) bool) Transformer {
	return func(src xml.TokenReader) *xml.Decoder {
		return xml.NewTokenDecoder(&attrRemover{
			d: src,
			f: f,
		})
	}
}

type attrRemover struct {
	d xml.TokenReader
	f func(xml.StartElement, xml.Attr) bool
}

func (ar *attrRemover) Token() (xml.Token, error) {
	tok, err := ar.d.Token()
	if err != nil {
		return tok, err
	}

	start, ok := tok.(xml.StartElement)
	if !ok {
		return tok, err
	}

	b := start.Attr[:0]
	for _, attr := range start.Attr {
		if !ar.f(start, attr) {
			b = append(b, attr)
		}
	}
	start.Attr = b

	return start, nil
}
