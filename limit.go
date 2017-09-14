// Copyright 2017 Sam Whited.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package xmlstream

import (
	"encoding/xml"
	"io"
)

// LimitReader returns a TokenReader that reads from r but stops with EOF after
// n tokens (regardless of the validity of the XML at that point in the stream).
func LimitReader(r TokenReader, n uint) TokenReader {
	return ReaderFunc(func() (xml.Token, error) {
		if n <= 0 {
			return nil, io.EOF
		}
		n--
		return r.Token()
	})
}
