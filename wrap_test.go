// Copyright 2017 Sam Whited.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package xmlstream_test

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"mellium.im/xmlstream"
)

func TestUnwrap(t *testing.T) {
	for i, tc := range [...]struct {
		I   string
		O   string
		T   xml.Token
		Err error
	}{
		0: {Err: io.EOF},
		1: {`Test<test/>Test`, `<test></test>Test`, xml.CharData("Test"), nil},
		2: {`<msg>Test</msg>`, `Test`, xml.StartElement{Name: xml.Name{Local: "msg"}, Attr: []xml.Attr{}}, nil},
		3: {`<msg><msg>Test</msg></msg>`, `<msg>Test</msg>`, xml.StartElement{Name: xml.Name{Local: "msg"}, Attr: []xml.Attr{}}, nil},
		4: {`<msg>A<msg>Test</msg>B</msg>`, `A<msg>Test</msg>B`, xml.StartElement{Name: xml.Name{Local: "msg"}, Attr: []xml.Attr{}}, nil},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			b := new(bytes.Buffer)
			d := xml.NewDecoder(strings.NewReader(tc.I))
			e := xml.NewEncoder(b)

			r, tok, err := xmlstream.Unwrap(d)
			if err != tc.Err {
				t.Errorf("Got unexpected error, want=`%v`, got=`%v`", tc.Err, err)
			}

			if err := xmlstream.Copy(e, r); err != nil {
				t.Fatal(err)
			}

			if s := b.String(); s != tc.O {
				t.Errorf("Invalid output, want=`%s`, got=`%s`", tc.O, s)
			}
			if _, ok := tok.(xml.StartElement); !ok && r.(*xml.Decoder) != d {
				t.Errorf("Expected stream that does not return start element to return original TokenReader")
			}
			if !reflect.DeepEqual(tc.T, tok) {
				t.Errorf("Input token does not match output token: want=`%T %v`, got=`%T, %v`", tc.T, tc.T, tok, tok)
			}
		})
	}
}

func TestInner(t *testing.T) {
	for i, tc := range [...]struct {
		N   int
		I   string
		O   string
		Err error
	}{
		0: {Err: io.EOF},
		1: {0, `Test<test/>Test`, `Test<test></test>Test`, nil},
		2: {1, `<msg>Test</msg>`, `Test`, nil},
		3: {1, `<msg><msg>Test</msg></msg>`, `<msg>Test</msg>`, nil},
		4: {1, `<msg>A<msg>Test</msg>B</msg>`, `A<msg>Test</msg>B`, nil},
		5: {2, `<msg>A<!proc>B</msg>`, `<!proc>B`, nil},
		6: {1, `<msg></msg><msg></msg>`, ``, nil},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			b := new(bytes.Buffer)
			d := xml.NewDecoder(strings.NewReader(tc.I))
			e := xml.NewEncoder(b)

			// Consume the first N tokens.
			for i := 0; i < tc.N; i++ {
				_, err := d.Token()
				if err != nil {
					t.Fatal(err)
				}
			}

			r := xmlstream.Inner(d)

			if err := xmlstream.Copy(e, r); err != nil {
				t.Fatal(err)
			}

			if s := b.String(); s != tc.O {
				t.Errorf("Invalid output, want=`%s`, got=`%s`", tc.O, s)
			}

			if _, err := r.Token(); err != io.EOF {
				t.Error("Expected token stream to continue returning io.EOF")
			}
		})
	}
}
