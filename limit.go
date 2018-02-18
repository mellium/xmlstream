// Copyright 2017 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package xmlstream

import (
	"encoding/xml"
	"io"
)

// LimitReader returns a xml.TokenReader that reads from r but stops with EOF
// after n tokens (regardless of the validity of the XML at that point in the
// stream).
func LimitReader(r xml.TokenReader, n uint) xml.TokenReader {
	return ReaderFunc(func() (xml.Token, error) {
		if n <= 0 {
			return nil, io.EOF
		}
		n--
		return r.Token()
	})
}
