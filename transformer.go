// Copyright 2016 Sam Whited.
// Use of this source code is governed by the BSD 2-clause license that can be
// found in the LICENSE file.

package xmlstream

import (
	"encoding/xml"
)

// A Tokenizer is anything that can decode a stream of XML tokens, including an
// xml.Decoder.
type Tokenizer interface {
	Token() (xml.Token, error)
	Skip() error
}

// A Transformer returns a new tokenizer that returns transformed tokens read
// from src.
type Transformer func(src Tokenizer) Tokenizer

// Inspect performs an operation for each token in the stream without
// transforming the stream in any way.
func Inspect(f func(t xml.Token)) Transformer {
	return func(src Tokenizer) Tokenizer {
		return inspector{
			Tokenizer: src,
			f:         f,
		}
	}
}

type inspector struct {
	Tokenizer
	f func(t xml.Token)
}

func (t inspector) Token() (xml.Token, error) {
	tok, err := t.Tokenizer.Token()
	if err != nil {
		return nil, err
	}
	t.f(tok)
	return tok, nil
}

// Map returns a Transformer that maps the tokens in the input using the given
// mapping.
func Map(mapping func(t xml.Token) xml.Token) Transformer {
	return func(src Tokenizer) Tokenizer {
		return mapper{
			t: src,
			f: mapping,
		}
	}
}

type mapper struct {
	t Tokenizer
	f func(t xml.Token) xml.Token
}

func (m mapper) Token() (xml.Token, error) {
	tok, err := m.t.Token()
	if err != nil {
		return nil, err
	}
	return m.f(tok), nil
}

func (m mapper) Skip() error {
	return m.t.Skip()
}

// Remove returns a Transformer that removes tokens for which f matches.
func Remove(f func(t xml.Token) bool) Transformer {
	return func(src Tokenizer) Tokenizer {
		return remover{
			t: src,
			f: f,
		}
	}
}

type remover struct {
	t Tokenizer
	f func(t xml.Token) bool
}

func (r remover) Token() (t xml.Token, err error) {
	for {
		t, err = r.t.Token()
		switch {
		case err != nil:
			return nil, err
		case r.f(t):
			continue
		}
		return
	}
}

func (r remover) Skip() error {
	return r.t.Skip()
}

// RemoveElement returns a Transformer that removes entire elements (and their
// children) if f matches the elements start token.
func RemoveElement(f func(start xml.StartElement) bool) Transformer {
	return func(src Tokenizer) Tokenizer {
		return &elementremover{
			f: f,
			t: src,
		}
	}
}

type elementremover struct {
	f func(start xml.StartElement) bool
	t Tokenizer
}

func (er *elementremover) Token() (t xml.Token, err error) {
	for {
		t, err = er.t.Token()
		if err != nil {
			return nil, err
		}
		if start, ok := t.(xml.StartElement); ok && er.f(start) {
			// Skip the element and read a new token.
			if err = er.Skip(); err != nil {
				return nil, err
			}
			continue
		}

		return
	}
}

func (er *elementremover) Skip() error {
	return er.t.Skip()
}
