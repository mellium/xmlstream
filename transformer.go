// Copyright 2016 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package xmlstream

import (
	"encoding/xml"
	"io"
)

// Discard returns a TokenWriter on which all calls succeed without doing
// anything.
func Discard() TokenWriter {
	return discard{}
}

type discard struct{}

func (discard) EncodeToken(_ xml.Token) error { return nil }

// TokenWriter is anything that can encode tokens to an XML stream, including an
// xml.Encoder.
type TokenWriter interface {
	EncodeToken(t xml.Token) error
}

// The Flusher interface is implemented by TokenWriters that can flush buffered
// data to an underlying receiver.
type Flusher interface {
	Flush() error
}

// WriterTo writes tokens to w until there are no more tokens to write or when
// an error occurs.
// The return value n is the number of tokens written.
// Any error encountered during the write is also returned.
//
// The Copy function uses WriterTo if available.
type WriterTo interface {
	WriteXML(TokenWriter) (n int, err error)
}

// ReaderFrom reads tokens from r until EOF or error.
// The return value n is the number of tokens read.
// Any error except io.EOF encountered during the read is also returned.
//
// The Copy function uses ReaderFrom if available.
type ReaderFrom interface {
	ReadXML(xml.TokenReader) (n int, err error)
}

// Marshaler is the interface implemented by objects that can marshal themselves
// into valid XML elements.
type Marshaler interface {
	TokenReader() xml.TokenReader
}

// TokenReadWriter is the interface that groups the basic Token, EncodeToken,
// and Flush methods.
type TokenReadWriter interface {
	xml.TokenReader
	TokenWriter
}

// TokenReadWriteCloser is the interface that groups the basic Token,
// EncodeToken, Flush, and Close methods.
type TokenReadWriteCloser interface {
	xml.TokenReader
	TokenWriter
	io.Closer
}

// TokenWriteCloser is the interface that groups the basic EncodeToken, and
// Close methods.
type TokenWriteCloser interface {
	TokenWriter
	io.Closer
}

// EncodeCloser is the interface that groups Encoder and io.Closer.
type EncodeCloser interface {
	Encoder
	io.Closer
}

// DecodeCloser is the interface that groups Decoder and io.Closer.
type DecodeCloser interface {
	Decoder
	io.Closer
}

// TokenWriteFlushCloser is the interface that groups the basic EncodeToken,
// Flush, and Close methods.
type TokenWriteFlushCloser interface {
	TokenWriter
	io.Closer
	Flusher
}

// TokenWriteFlusher is the interface that groups the basic EncodeToken, and
// Flush methods.
type TokenWriteFlusher interface {
	TokenWriter
	Flusher
}

// TokenReadCloser is the interface that groups the basic Token and Close
// methods.
type TokenReadCloser interface {
	xml.TokenReader
	io.Closer
}

// Encoder is the interface that groups the Encode, EncodeElement, and
// EncodeToken methods.
// Encoder is implemented by xml.Encoder.
type Encoder interface {
	TokenWriter
	Encode(v interface{}) error
	EncodeElement(v interface{}, start xml.StartElement) error
}

// Decoder is the interface that groups the Decode, DecodeElement, and Token
// methods.
// Decoder is implemented by xml.Decoder.
type Decoder interface {
	xml.TokenReader
	Decode(v interface{}) error
	DecodeElement(v interface{}, start *xml.StartElement) error
}

// DecodeEncoder is the interface that groups the Encoder and Decoder
// interfaces.
type DecodeEncoder interface {
	Decoder
	Encoder
}

// TokenReadEncoder is the interface that groups the Encode, EncodeElement,
// EncodeToken, and Token methods.
type TokenReadEncoder interface {
	xml.TokenReader
	Encoder
}

// A Transformer returns a new TokenReader that returns transformed tokens
// read from src.
type Transformer func(src xml.TokenReader) xml.TokenReader

// Inspect performs an operation for each token in the stream without
// transforming the stream in any way.
func Inspect(f func(t xml.Token)) Transformer {
	return func(src xml.TokenReader) xml.TokenReader {
		return inspector{
			d: src,
			f: f,
		}
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
	return func(src xml.TokenReader) xml.TokenReader {
		return &mapper{
			d: src,
			f: mapping,
		}
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
	return func(src xml.TokenReader) xml.TokenReader {
		return remover{
			d: src,
			f: f,
		}
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
	return func(src xml.TokenReader) xml.TokenReader {
		return &elementremover{
			d: src,
			f: f,
		}
	}
}

type elementremover struct {
	d xml.TokenReader
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
			if err = Skip(er.d); err != nil {
				return t, err
			}
			continue
		}

		return
	}
}

// BUG(ssw): Multiple uses of RemoveAttr will iterate over the attr list multiple times.

// RemoveAttr returns a Transformer that removes attributes from
// xml.StartElement's if f matches.
func RemoveAttr(f func(start xml.StartElement, attr xml.Attr) bool) Transformer {
	return func(src xml.TokenReader) xml.TokenReader {
		return &attrRemover{
			d: src,
			f: f,
		}
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

// InsertFunc calls f after writing any start element to the stream.
// The function can decide based on the passed in StartElement whether to insert
// any additional tokens into the stream by writing them to w.
func InsertFunc(f func(start xml.StartElement, level uint64, w TokenWriter) error) Transformer {
	if f == nil {
		f = func(xml.StartElement, uint64, TokenWriter) error { return nil }
	}

	var depth uint64
	var pr *PipeReader
	return func(r xml.TokenReader) xml.TokenReader {
		return ReaderFunc(func() (xml.Token, error) {
			if pr != nil {
				tok, err := pr.Token()
				if tok != nil || err != io.EOF {
					return tok, err
				}
				pr = nil
			}

			tok, err := r.Token()
			if err != nil {
				return tok, err
			}
			switch t := tok.(type) {
			case xml.StartElement:
				depth++
				var pw *PipeWriter
				pr, pw = Pipe()
				go func() {
					pw.CloseWithError(f(t, depth, pw))
				}()
			case xml.EndElement:
				if depth > 0 {
					depth--
				}
			}

			return tok, err
		})
	}
}

// Insert adds one XML stream to another just before the close token, matching
// on the token name.
// If either component of the name is empty it is considered a wildcard.
func Insert(name xml.Name, m Marshaler) Transformer {
	return func(r xml.TokenReader) xml.TokenReader {
		var inner xml.TokenReader
		return ReaderFunc(func() (xml.Token, error) {
			if inner != nil {
				tok, err := inner.Token()
				switch {
				case tok != nil && err == io.EOF:
					inner = nil
					return tok, nil
				case tok == nil && err == io.EOF:
					inner = nil
				default:
					return tok, err
				}
			}

			tok, err := r.Token()
			if err != nil {
				return tok, err
			}

			if end, ok := tok.(xml.EndElement); ok &&
				((name.Space == "" && name.Local == "") ||
					end.Name == name ||
					(end.Name.Space == name.Space && name.Local == "") ||
					(end.Name.Local == name.Local && name.Space == "")) {
				inner = MultiReader(m.TokenReader(), Token(end))
				return inner.Token()
			}

			return tok, err
		})
	}
}
