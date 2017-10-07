// Copyright 2017 Sam Whited.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package xmlstream

import (
	"encoding/xml"
	"io"
)

// Copy consumes a TokenReader and writes its tokens to a TokenWriter unti
// leither io.EOF is reached on d or an error occurs.
// It returns the number of tokens copied and the first error encountered while
// copying, if any.
// If an error is returned by the reader or writer, copy returns it immediately.
// Since Copy is defined as consuming the stream until the end, io.EOF is not
// returned.
// If no error would be returned, Copy flushes the TokenWriter when it is done.
func Copy(e TokenWriter, d TokenReader) (n int, err error) {
	defer func() {
		if err == nil || err == io.EOF {
			err = e.Flush()
		}
	}()

	var tok xml.Token
	for {
		tok, err = d.Token()
		if (err != nil && err != io.EOF) || (tok == nil && err == io.EOF) {
			return n, err
		}

		if err := e.EncodeToken(tok); err != nil {
			return n, err
		}
		n++
		if err == io.EOF {
			return n, nil
		}
	}
}
