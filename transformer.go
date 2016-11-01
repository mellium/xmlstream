// Copyright 2016 Sam Whited.
// Use of this source code is governed by the BSD 2-clause license that can be
// found in the LICENSE file.

// Package xmlstream provides an experimental API for streaming, transforming,
// and otherwise manipulating XML data.
//
// Be advised: This API is still unstable and is subject to change.
package xmlstream // import "mellium.im/xmlstream"

import (
	"encoding/xml"
	"io"
)

// Transformer transforms tokens on an XML stream.
type Transformer interface {
	// Transform reads tokens from src, performs some transformation on them, and
	// writes the new tokens to dst.
	Transform(dst *xml.Encoder, src *xml.Decoder) error

	// Reset resets the state and allows a Transformer to be reused.
	Reset()
}

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

func (nopTransformer) Transform(dst *xml.Encoder, src *xml.Decoder) error {
	t, err := src.Token()
	if err != nil {
		return err
	}
	return dst.EncodeToken(t)
}

// NewEncoder returns a new xml.Encoder that wraps e by transforming any tokens
// encoded to it before passing them along to the e.
func NewEncoder(e *xml.Encoder, t Transformer) *xml.Encoder {
	r, w := io.Pipe()
	pipeencoder := xml.NewEncoder(w)
	d := xml.NewDecoder(r)
	go func() {
		if err := t.Transform(e, d); err != nil {
			return
		}
	}()
	return pipeencoder
}

// NewDecoder returns a new xml.Decoder that wraps d by transforming any tokens
// decoded from it before returning them.
func NewDecoder(d *xml.Decoder, t Transformer) *xml.Decoder {
	r, w := io.Pipe()
	pipedecoder := xml.NewDecoder(r)
	e := xml.NewEncoder(w)
	go func() {
		if err := t.Transform(e, d); err != nil {
			return
		}
	}()
	return pipedecoder
}
