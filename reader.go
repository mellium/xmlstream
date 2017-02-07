// Copyright 2016 Sam Whited.
// Use of this source code is governed by the BSD 2-clause license that can be
// found in the LICENSE file.

package xmlstream

import (
	"bufio"
	"encoding/xml"
	"errors"
	"fmt"
	"io"

	"mellium.im/reader"
)

var (
	notEndError   = errors.New("Expected end element, found something else")
	notStartError = errors.New("Expected start element, found something else")
)

type unexpectedEndError struct {
	localName string
}

func (u unexpectedEndError) Error() string {
	return fmt.Sprintf("Unexpected end element </%s>", u.localName)
}

// TODO: We almost certainly need to expose the start token somehow, but I can't
//       think of a clean API to do it.

// InnerReader is an io.Reader which attempts to decode an xml.StartElement from
// the stream on the first call to Read (returning an error if an invalid start
// token is found) and returns a new reader which only reads the inner XML
// without parsing it or checking its validity.
// After the inner XML is read, the end token is parsed and if it does not exist
// or does not match the original start token an error is returned.
func InnerReader(r io.Reader) io.Reader {
	var end xml.EndElement

	br := bufio.NewReader(r)
	d := xml.NewDecoder(br)

	lr := &io.LimitedReader{
		R: br,
		N: 0,
	}

	// After the body has been read, pop the end token and verify that it matches
	// the start token.
	after := reader.After(lr, func() error {
		tok, err := d.RawToken()
		if err != nil {
			return err
		}
		rawend, ok := tok.(xml.EndElement)
		switch {
		case !ok:
			return notEndError
		case rawend != end:
			return unexpectedEndError{rawend.Name.Local}
		}
		return nil
	})

	// Before we read the body, pop the start token and set the length on the
	// limit reader to the length of the inner XML.
	before := reader.Before(after, func() error {
		tok, err := d.RawToken()
		if err != nil {
			return err
		}
		rawstart, ok := tok.(xml.StartElement)
		if !ok {
			return notStartError
		}
		// Don't use rawstart.End() because that apparently handles namespace
		// prefixes even though it's a raw token.
		end = xml.EndElement{Name: rawstart.Name}
		lr.N = int64(br.Buffered() - len(rawstart.Name.Local) - len(rawstart.Name.Space) - 3) // 3 == len('</>')
		// If there is a namespace on the rawtoken, subtract one more for the ":"
		// separator (it's a prefix).
		if rawstart.Name.Space != "" {
			lr.N--
		}

		return nil
	})

	return before
}

// numReadReader keeps track of the number of bytes that have been read during
// the lifetime of the reader.
type numReadReader struct {
	R interface {
		io.ByteReader
		io.Reader
	}
	TotalRead int
}

func (nrr *numReadReader) Read(p []byte) (n int, err error) {
	n, err = nrr.R.Read(p)
	nrr.TotalRead += n
	return
}

func (nrr *numReadReader) ReadByte() (byte, error) {
	b, err := nrr.R.ReadByte()
	nrr.TotalRead++
	return b, err
}
