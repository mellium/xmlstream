// Copyright 2017 Sam Whited.
// Use of this source code is governed by the BSD 2-clause license that can be
// found in the LICENSE file.

package xmlstream

import (
	"encoding/xml"
	"io"
)

// Wrap wraps a token stream in a start element and its corresponding end
// element.
func Wrap(start xml.StartElement, r xml.TokenReader) xml.TokenReader {
	state := 0
	return ReaderFunc(func() (t xml.Token, err error) {
		switch state {
		case 0:
			state++
			return start, nil
		case 1:
			t, err = r.Token()
			if err == io.EOF {
				state++
				err = nil
			}
			if t == nil {
				state++
				t = start.End()
			}
			return t, err
		case 2:
			state++
			return start.End(), io.EOF
		default:
			return nil, io.EOF
		}
	})
}
