// Copyright 2017 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package xmlstream

import (
	"encoding/xml"
	"io"
)

// Copy consumes a xml.TokenReader and writes its tokens to a TokenWriter until
// either io.EOF is reached on src or an error occurs.
// It returns the number of tokens copied and the first error encountered while
// copying, if any.
// If an error is returned by the reader or writer, copy returns it immediately.
// Since Copy is defined as consuming the stream until the end, io.EOF is not
// returned.
//
// If src implements the WriterTo interface, the copy is implemented by calling
// src.WriteXML(dst). Otherwise, if dst implements the ReaderFrom interface, the
// copy is implemented by calling dst.ReadXML(src).
func Copy(dst TokenWriter, src xml.TokenReader) (n int, err error) {
	if wt, ok := src.(WriterTo); ok {
		return wt.WriteXML(dst)
	}
	if rt, ok := dst.(ReaderFrom); ok {
		return rt.ReadXML(src)
	}

	var tok xml.Token
	for {
		tok, err = src.Token()
		switch {
		case err != nil && err != io.EOF:
			return n, err
		case tok == nil && err == io.EOF:
			return n, nil
		case tok == nil && err == nil:
			return n, nil
		}

		if err := dst.EncodeToken(tok); err != nil {
			return n, err
		}
		n++
		if err == io.EOF {
			return n, nil
		}
	}
}
