// Copyright 2017 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package xmlstream_test

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"testing"

	"mellium.im/xmlstream"
)

type copyTest struct {
	r      xml.TokenReader
	n      int
	err    error
	out    string
	panics bool
}

var errTest = fmt.Errorf("test err")

var copyTests = [...]copyTest{
	0: {panics: true},
	1: {
		r:   xml.NewDecoder(strings.NewReader(`<t></t>`)),
		n:   2,
		out: `<t></t>`,
	},
	2: {
		r: xmlstream.ReaderFunc(func() (t xml.Token, err error) {
			return xml.CharData("Test"), io.EOF
		}),
		n:   1,
		out: `Test`,
		err: nil,
	},
	3: {
		r: xmlstream.ReaderFunc(func() (t xml.Token, err error) {
			return xml.CharData("Test"), errTest
		}),
		n:   0,
		out: ``,
		err: errTest,
	},
	4: {
		// Make sure that we don't try to encode nil tokens or enter an infinite
		// loop when TokenDecoders return nil, nil.
		r:   xml.NewTokenDecoder(xmlstream.Wrap(nil, xml.StartElement{Name: xml.Name{Local: "test"}})),
		n:   2,
		out: `<test></test>`,
	},
}

func TestCopy(t *testing.T) {
	for i, tc := range copyTests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			defer func() {
				r := recover()
				switch {
				case r == nil && tc.panics:
					t.Errorf("Expected panic")
				case r != nil && !tc.panics:
					t.Errorf("Got unexpected panic")
				}
			}()
			b := new(bytes.Buffer)
			e := xml.NewEncoder(b)
			n, err := xmlstream.Copy(e, tc.r)
			if e := e.Flush(); e != nil {
				t.Fatalf("Unexpected error flushing: %q", e)
			}

			if n != tc.n {
				t.Errorf("Wrong number of tokens copied: want=`%d', got=`%d'", tc.n, n)
			}
			if err != tc.err {
				t.Errorf("Unexpected error: want=`%v', got=`%v'", tc.err, err)
			}
			if o := b.String(); o != tc.out {
				t.Errorf("Unexpected output: want=`%v', got=`%v'", tc.out, o)
			}
		})
	}
}

type errTokenWriter struct{}

func (errTokenWriter) EncodeToken(t xml.Token) error {
	return errTest
}

func (errTokenWriter) Flush() error {
	return nil
}

func TestCopyBadEncode(t *testing.T) {
	n, err := xmlstream.Copy(errTokenWriter{}, xmlstream.Wrap(nil, xml.StartElement{Name: xml.Name{Local: "start"}}))
	if n != 0 {
		t.Errorf("Expected no tokens to be copied, got %d", n)
	}
	if err != errTest {
		t.Errorf("Expected testErr to be returned, got %v", err)
	}
}
