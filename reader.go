// Copyright 2016 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

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
	errNotEnd   = errors.New("expected end element, found something else")
	errNotStart = errors.New("expected start element, found something else")
)

type unexpectedEndError struct {
	name xml.Name
}

func (u unexpectedEndError) Error() string {
	if u.name.Space == "" {
		return fmt.Sprintf("Unexpected end element </%s>", u.name.Local)
	}
	return fmt.Sprintf("Unexpected end element </%s:%s>", u.name.Local, u.name.Space)
}

type nopCloser struct {
	xml.TokenReader
}

func (nopCloser) Close() error { return nil }

// NopCloser returns a TokenReadCloser with a no-op Close method wrapping
// the provided Reader r.
func NopCloser(r xml.TokenReader) TokenReadCloser {
	return nopCloser{r}
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
	after := reader.After(lr, func(n int, err error) (int, error) {
		if err != io.EOF {
			return n, err
		}

		tok, err := d.RawToken()
		if err != nil {
			return n, err
		}
		rawend, ok := tok.(xml.EndElement)
		switch {
		case !ok:
			return n, errNotEnd
		case rawend != end:
			return n, unexpectedEndError{rawend.Name}
		}
		return n, nil
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
			return errNotStart
		}
		// Don't use rawstart.End() because that apparently handles namespace
		// prefixes even though it's a raw token.
		end = xml.EndElement{Name: rawstart.Name}
		// 3 == len('</>')
		lr.N = int64(br.Buffered() - len(rawstart.Name.Local) - len(rawstart.Name.Space) - 3)
		// If there is a namespace on the rawtoken, subtract one more for the ":"
		// separator (it's a prefix).
		if rawstart.Name.Space != "" {
			lr.N--
		}

		return nil
	})

	return before
}
