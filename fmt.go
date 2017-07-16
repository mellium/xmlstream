// Copyright 2016 Sam Whited.
// Use of this source code is governed by the BSD 2-clause license that can be
// found in the LICENSE file.

package xmlstream

import (
	"bytes"
	"encoding/xml"
)

// Fmt returns a transformer that indents the given XML stream.  The default
// indentation style is to remove non-significant whitespace, start elements on
// a new line and indent two spaces per level.
func Fmt(t TokenReader, opts ...FmtOption) TokenReader {
	f := &fmter{t: whitespaceRemover(t)}
	f.getOpts(opts)
	return f
}

// FmtOption is used to configure a formatters behavior.
type FmtOption func(*fmter)

// Prefix is inserted at the start of every XML element in the stream.
// The default prefix if this option is not specified is '\n'.
func Prefix(s string) FmtOption {
	return func(f *fmter) {
		f.prefix = []byte(s)
	}
}

// Indent is inserted before XML elements zero or more times according to
// their nesting depth in the stream.
// The default indentation is "  " (two ASCII spaces).
func Indent(s string) FmtOption {
	return func(f *fmter) {
		f.indent = []byte(s)
	}
}

type fmter struct {
	nesting int
	indent  []byte
	prefix  []byte
	queue   []xml.Token
	t       TokenReader
}

func (f *fmter) Token() (t xml.Token, err error) {
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

	if _, ok := t.(xml.CharData); ok {
		return t, nil
	}

	toks := []xml.Token{}

	// Add prefix
	if len(f.prefix) > 0 {
		toks = append(toks, xml.CharData(f.prefix))
	}

	// Add indentation
	switch t.(type) {
	case xml.EndElement:
		// Decrease the indentation level.
		f.nesting--

		if len(f.indent) > 0 && f.nesting > 0 {
			indent := xml.CharData(bytes.Repeat(f.indent, f.nesting))
			toks = append(toks, indent)
		}
	case xml.StartElement:
		if len(f.indent) > 0 && f.nesting > 0 {
			indent := xml.CharData(bytes.Repeat(f.indent, f.nesting))
			toks = append(toks, indent)
		}

		// Increase the indentation level.
		f.nesting++
	default:
		if len(f.indent) > 0 && f.nesting > 0 {
			indent := xml.CharData(bytes.Repeat(f.indent, f.nesting))
			toks = append(toks, indent)
		}
	}

	// Add original token
	toks = append(toks, t)

	// Queue up all tokens but the first one
	if len(toks) > 1 {
		f.queue = append(f.queue, toks[1:]...)
	}

	// Return the token that would be the next one on the queue
	return toks[0], nil
}

func (f *fmter) Skip() error {
	return f.t.Skip()
}

func (f *fmter) getOpts(opts []FmtOption) {
	f.indent = []byte{' ', ' '}
	f.prefix = []byte{'\n'}
	for _, opt := range opts {
		opt(f)
	}
}
