// Copyright 2016 Sam Whited.
// Use of this source code is governed by the BSD 2-clause license that can be
// found in the LICENSE file.

// Package xmlstream provides an API for streaming, transforming, and otherwise
// manipulating XML data.
package xmlstream // import "mellium.im/xmlstream"

import (
	"encoding/xml"
)

// A Tokenizer is anything that can decode a stream of XML tokens.
type Tokenizer interface {
	Token() (xml.Token, error)
	Skip() error
}

// A Transformer returns a new tokenizer that wraps src and optionally
// transforms any tokens read from it.
type Transformer func(src Tokenizer) Tokenizer

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
