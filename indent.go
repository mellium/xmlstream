// Copyright 2016 Sam Whited.
// Use of this source code is governed by the BSD 2-clause license that can be
// found in the LICENSE file.

package xmlstream

import (
	"bytes"
	"encoding/xml"
)

// Indent returns a transformer that indents the given XML stream.
// The default indentation style is to remove non-significant whitespace, start
// elements on a new line and indent two spaces per level.
func Indent(t Tokenizer, opts ...IndentOption) Tokenizer {
	f := &indenter{t: whitespaceRemover(t)}
	f.getOpts(opts)
	return f
}

// IndentOption is used to configure a formatters behavior.
type IndentOption func(*indenter)

// Prefix is inserted at the start of every XML element in the stream.
// The default prefix if this option is not specified is '\n'.
func Prefix(s string) IndentOption {
	return func(f *indenter) {
		f.prefix = []byte(s)
	}
}

// Indentation is inserted before XML elements zero or more times according to
// their nesting depth in the stream.
// The default indentation is "  " (two ASCII spaces).
func Indentation(s string) IndentOption {
	return func(f *indenter) {
		f.indent = []byte(s)
	}
}

type indenter struct {
	nesting int
	indent  []byte
	prefix  []byte
	queue   []xml.Token
	t       Tokenizer
}

func (f *indenter) Token() (t xml.Token, err error) {
	// If we've queued up a token to write next, go ahead and pop the next token
	// off the queue.
	if len(f.queue) > 0 {
		t, f.queue = f.queue[0], f.queue[1:]
		return
	}

	t, err = f.t.Token()
	if err != nil {
		return
	}

	switch t.(type) {
	case xml.StartElement:
		// TODO: Can this all be done more efficiently?
		toks := []xml.Token{}
		if len(f.prefix) > 0 {
			toks = append(toks, xml.CharData(f.prefix))
		}
		if len(f.indent) > 0 && f.nesting > 0 {
			indent := xml.CharData(bytes.Repeat(f.indent, f.nesting))
			toks = append(toks, indent)
		}
		toks = append(toks, t)
		if len(toks) > 1 {
			f.queue = append(f.queue, toks[1:]...)
		}

		// Increase the indentation level.
		f.nesting++

		return toks[0], nil
	case xml.EndElement:
		// Decrease the indentation level.
		f.nesting--
		toks := []xml.Token{}
		if len(f.prefix) > 0 {
			toks = append(toks, xml.CharData(f.prefix))
		}
		if len(f.indent) > 0 && f.nesting > 0 {
			indent := xml.CharData(bytes.Repeat(f.indent, f.nesting))
			toks = append(toks, indent)
		}
		toks = append(toks, t)
		if len(toks) > 1 {
			f.queue = append(f.queue, toks[1:]...)
		}

		return toks[0], nil
	}

	return
}

func (f *indenter) Skip() error {
	return f.t.Skip()
}

func (f *indenter) getOpts(opts []IndentOption) {
	f.indent = []byte{' ', ' '}
	f.prefix = []byte{'\n'}
	for _, opt := range opts {
		opt(f)
	}
}
