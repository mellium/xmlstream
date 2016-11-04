// Copyright 2016 Sam Whited.
// Use of this source code is governed by the BSD 2-clause license that can be
// found in the LICENSE file.

// Package xmlstream provides an experimental API for streaming, transforming,
// and otherwise manipulating XML data.
//
// Be advised: This API is unstable and subject to change.
package xmlstream // import "mellium.im/xmlstream"

import (
	"encoding/xml"
)

// A Tokenizer is anything that can decode a stream of XML tokens.
type Tokenizer interface {
	Token() (xml.Token, error)
	RawToken() (xml.Token, error)
	Skip() error
}

// A Encoder is anything that can encode a stream of XML tokens.
type Encoder interface {
	EncodeToken(t xml.Token) error
	Encode(v interface{}) error
	EncodeElement(v interface{}, start xml.StartElement) error
	Flush() error
}

// Transformer transforms tokens on an XML stream.
type Transformer interface {
	// Transform decodes tokens from src, performs some transformation on them,
	// and encodes the new tokens to dst. It should attempt to consume tokens from
	// src until an error is returned. It may or may not return nil instead of
	// io.EOF depending on the implementation.
	Transform(dst Encoder, src Tokenizer) error

	// Reset resets the state and allows a Transformer to be reused.
	Reset()
}

// TransformerFunc is an adapter to allow the use of ordinary functions as XML
// transformers. If f is a function with the appropriate signature,
// TransformerFunc(f) is a Transformer that calls f.
type TransformerFunc func(dst Encoder, src Tokenizer) error

// Transform calls f(w, r).
func (f TransformerFunc) Transform(dst Encoder, src Tokenizer) error {
	return f(dst, src)
}

// TODO:
// Chain returns a Transformer that applies t in sequence.
// func Chain(t ...Transformer) Transformer {
// }

// NopResetter can be embedded by implementations of Transformer to add a nop
// Reset method.
type NopResetter struct{}

// Reset implements the Reset method of the Transformer interface.
func (NopResetter) Reset() {
}

// NopTransformer returns a transformer that does nothing but pass tokens
// through the pipeline from the input decoder to the output encoder.
func NopTransformer() Transformer {
	return nopTransformer{}
}

type nopTransformer struct {
	NopResetter
}

func (nopTransformer) Transform(dst Encoder, src Tokenizer) error {
	t, err := src.Token()
	if err != nil {
		return err
	}
	return dst.EncodeToken(t)
}
