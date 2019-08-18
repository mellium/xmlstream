// Copyright 2016 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package xmlstream

import (
	"bytes"
	"encoding/xml"
)

// Fmt returns a transformer that indents the given XML stream.
// The default indentation style is to remove non-significant whitespace, start
// elements on a new line and indent two spaces per level.
func Fmt(d xml.TokenReader, opts ...FmtOption) xml.TokenReader {
	f := &fmter{d: whitespaceRemover(d)}
	f.getOpts(opts)
	return f
}

// FmtOption is used to configure a formatters behavior.
type FmtOption func(*fmter)

// Prefix is inserted at the start of every XML element in the stream.
func Prefix(s string) FmtOption {
	return func(f *fmter) {
		f.prefix = []byte(s)
	}
}

// Suffix is inserted at the start of every XML element in the stream.
// If no option is specified the default suffix is '\n'.
func Suffix(s string) FmtOption {
	return func(f *fmter) {
		f.suffix = []byte(s)
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
	d       xml.TokenReader
	nesting int
	indent  []byte
	prefix  []byte
	suffix  []byte
	queue   []xml.Token
}

func (f *fmter) addIndent(toks []xml.Token) []xml.Token {
	if len(f.indent) > 0 && f.nesting > 0 {
		indent := xml.CharData(bytes.Repeat(f.indent, f.nesting))
		toks = append(toks, indent)
	}
	return toks
}

func (f *fmter) Token() (t xml.Token, err error) {
	// If we've queued up a token to write next, go ahead and pop the next token
	// off the queue.
	if len(f.queue) > 0 {
		t, f.queue = f.queue[0], f.queue[1:]
		return
	}

	t, err = f.d.Token()
	if err != nil {
		return t, err
	}

	toks := []xml.Token{}

	// Add prefix
	if len(f.prefix) > 0 {
		toks = append(toks, xml.CharData(f.prefix))
	}

	// Add indentation
	switch t.(type) {
	case xml.CharData:
		// Don't indent chardata
	case xml.EndElement:
		// Decrease the indentation level.
		f.nesting--

		toks = f.addIndent(toks)
	case xml.StartElement:
		toks = f.addIndent(toks)

		// Increase the indentation level.
		f.nesting++
	default:
		toks = f.addIndent(toks)
	}

	// Add original token
	toks = append(toks, t)

	// Add suffix
	if len(f.suffix) > 0 {
		toks = append(toks, xml.CharData(f.suffix))
	}

	// Queue up all tokens but the first one
	if len(toks) > 1 {
		f.queue = append(f.queue, toks[1:]...)
	}

	// Return the token that would be the next one on the queue
	return toks[0], nil
}

func (f *fmter) getOpts(opts []FmtOption) {
	f.indent = []byte{' ', ' '}
	f.suffix = []byte{'\n'}
	for _, opt := range opts {
		opt(f)
	}
}
