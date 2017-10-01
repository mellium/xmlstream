// Copyright 2017 Sam Whited.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package xmlstream

import (
	"encoding/xml"
	"io"
)

// Copy consumes a TokenReader and writes its tokens to a TokenWriter.
// If an error is returned by the reader or writer, copy returns it immediately.
// Since Copy is defined as consuming the stream until the end, io.EOF is not
// returned.
// If no error would be returned, Copy flushes the TokenWriter when it is done.
func Copy(e TokenWriter, d TokenReader) (err error) {
	defer func() {
		if err == nil || err == io.EOF {
			err = e.Flush()
		}
	}()

	var tok xml.Token
	for {
		tok, err = d.Token()
		switch {
		case err == io.EOF && tok != nil:
			err = nil
		case err != nil:
			return err
		}

		if err = e.EncodeToken(tok); err != nil {
			return err
		}
	}
}
