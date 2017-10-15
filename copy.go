// Copyright 2017 Sam Whited.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package xmlstream

import (
	"encoding/xml"
	"io"
)

// Copy consumes a TokenReader and writes its tokens to a TokenWriter unti
// leither io.EOF is reached on src or an error occurs.
// It returns the number of tokens copied and the first error encountered while
// copying, if any.
// If an error is returned by the reader or writer, copy returns it immediately.
// Since Copy is defined as consuming the stream until the end, io.EOF is not
// returned.
// If no error would be returned, Copy flushes the TokenWriter when it is done.
//
// If src implements the WriterTo interface, the copy is implemented by calling
// src.WriteTo(dst). Otherwise, if dst implements the ReaderFrom interface, the
// copy is implemented by calling dst.ReadFrom(src).
func Copy(dst TokenWriter, src TokenReader) (n int, err error) {
	if wt, ok := src.(WriterTo); ok {
		return wt.WriteXML(dst)
	}
	if rt, ok := dst.(ReaderFrom); ok {
		return rt.ReadXML(src)
	}

	defer func() {
		if err == nil || err == io.EOF {
			err = dst.Flush()
		}
	}()

	var tok xml.Token
	for {
		tok, err = src.Token()
		if (err != nil && err != io.EOF) || (tok == nil && err == io.EOF) {
			return n, err
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
