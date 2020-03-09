// Copyright 2019 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause license that can be
// found in the LICENSE file.

package xmlstream

import (
	"encoding/xml"
	"io"
)

// ReadAll reads from r until an error or io.EOF and returns the data it reads.
// A successful call returns err == nil, not err == io.EOF.
// Because ReadAll is defined to read from src until io.EOF, it does not treat
// an io.EOF from Read as an error to be reported.
func ReadAll(r xml.TokenReader) ([]xml.Token, error) {
	var toks []xml.Token
	for {
		t, err := r.Token()
		switch err {
		case io.EOF:
			if t != nil {
				toks = append(toks, xml.CopyToken(t))
			}
			return toks, nil
		case nil:
			if t == nil {
				return toks, nil
			}
			toks = append(toks, xml.CopyToken(t))
		default:
			return toks, err
		}
	}
}
