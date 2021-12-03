// Copyright 2016 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package xmlstream

import (
	"encoding/xml"
	"unicode"
	"unicode/utf8"
)

var whitespaceRemover = Remove(func(t xml.Token) bool {
	if chars, ok := t.(xml.CharData); ok && isWhitespace(chars) {
		return true
	}
	return false
})

// TODO: Make a dummy text.Transformer instead?
func isWhitespace(b []byte) bool {
	for r, size := utf8.DecodeRune(b); r != utf8.RuneError; r, size = utf8.DecodeRune(b) {
		if !unicode.IsSpace(r) {
			return false
		}
		b = b[size:]
	}
	return true
}
