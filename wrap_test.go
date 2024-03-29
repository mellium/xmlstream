// Copyright 2017 The Mellium Contributors.
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

func TestToken(t *testing.T) {
	chars := xml.CharData(`a comparable token`)
	tr := xmlstream.Token(chars)

	tok, err := tr.Token()
	if string(tok.(xml.CharData)) != string(chars) {
		t.Errorf("First read got wrong token: want=%q, got=%q", chars, tok)
	}
	if err != io.EOF {
		t.Errorf("Wrong error: want=%q, got=%q", io.EOF, err)
	}

	tok, err = tr.Token()
	if err != io.EOF {
		t.Errorf("Wrong error on second read: want=%q, got=%q", io.EOF, err)
	}
	if tok != nil {
		t.Errorf("Got unexpected token on second read %T %[1]v", tok)
	}
}

func TestWrap(t *testing.T) {
	for i, tc := range [...]struct {
		I   xml.TokenReader
		O   string
		Err error
	}{
		0: {O: `<test></test>`},
		1: {I: xml.NewDecoder(strings.NewReader(`<a/>`)), O: `<test><a></a></test>`},
		2: {I: xmlstream.ReaderFunc(func() (xml.Token, error) {
			return xml.CharData("inner"), io.EOF
		}), O: `<test>inner</test>`},
		3: {I: func() xml.TokenReader {
			state := 0
			return xmlstream.ReaderFunc(func() (xml.Token, error) {
				if state > 0 {
					return nil, io.EOF
				}
				state++
				return xml.CharData("inner"), nil
			})
		}(), O: `<test>inner</test>`},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			b := new(bytes.Buffer)
			e := xml.NewEncoder(b)

			r := xmlstream.Wrap(tc.I, xml.StartElement{Name: xml.Name{Local: "test"}})

			if _, err := xmlstream.Copy(e, r); err != tc.Err {
				t.Errorf("Got unexpected error, want=`%v`, got=`%v`", tc.Err, err)
			}
			if err := e.Flush(); err != nil {
				t.Errorf("Error flushing: %q", err)
			}

			if s := b.String(); s != tc.O {
				t.Errorf("Invalid output, want=`%s`, got=`%s`", tc.O, s)
			}
		})
	}
}

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
		5: {`<msg>Foo</msg><remain/>`, `Foo<remain></remain>`, xml.StartElement{Name: xml.Name{Local: "msg"}, Attr: []xml.Attr{}}, nil},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			b := new(bytes.Buffer)
			d := xml.NewDecoder(strings.NewReader(tc.I))
			e := xml.NewEncoder(b)

			r, tok, err := xmlstream.Unwrap(d)
			if err != tc.Err {
				t.Errorf("Got unexpected error, want=`%v`, got=`%v`", tc.Err, err)
			}

			if _, err := xmlstream.Copy(e, r); err != nil {
				t.Fatal(err)
			}
			if err := e.Flush(); err != nil {
				t.Errorf("Error flushing: %q", err)
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

			if _, err := xmlstream.Copy(e, r); err != nil {
				t.Fatal(err)
			}
			if err := e.Flush(); err != nil {
				t.Errorf("Error flushing: %q", err)
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

func TestInnerElement(t *testing.T) {
	for i, tc := range [...]struct {
		N   int
		I   string
		O   string
		Err error
	}{
		0: {Err: io.EOF},
		1: {I: `Test<test/>Test`, O: `Test<test></test>Test`},
		2: {I: `<msg>Test</msg><a></a>`},
		3: {N: 1, I: `<msg><msg>Test</msg></msg><no/>`, O: `<msg><msg>Test</msg></msg>`},
		4: {I: `<msg>A<msg>Test</msg>B</msg>`},
		5: {N: 2, I: `<msg>A<!proc>B</msg><no/>`, O: `<msg>A<!proc>B</msg>`},
		6: {N: 1, I: `<msg></msg><no></no>`, O: `<msg></msg>`},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			b := new(bytes.Buffer)
			d := xml.NewDecoder(strings.NewReader(tc.I))
			e := xml.NewEncoder(b)

			// Encode the first N tokens outside of the inner reader.
			for i := 0; i < tc.N; i++ {
				tok, err := d.Token()
				if err != nil {
					t.Fatal(err)
				}
				err = e.EncodeToken(tok)
				if err != nil {
					t.Fatal(err)
				}
			}

			r := xmlstream.InnerElement(d)

			if _, err := xmlstream.Copy(e, r); err != nil {
				t.Fatal(err)
			}
			if err := e.Flush(); err != nil {
				t.Errorf("Error flushing: %q", err)
			}

			if tc.O == "" {
				tc.O = tc.I
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

func TestInnerElementEarlyEOF(t *testing.T) {
	r := xmlstream.Wrap(nil, xml.StartElement{Name: xml.Name{Local: "foo"}})
	tok, err := r.Token()
	if err != nil {
		t.Fatalf("error popping token: %v", err)
	}
	inner := xmlstream.InnerElement(r)
	b := new(bytes.Buffer)
	e := xml.NewEncoder(b)
	err = e.EncodeToken(tok)
	if err != nil {
		t.Fatalf("error encoding start token: %v", err)
	}
	_, err = xmlstream.Copy(e, inner)
	if err != nil {
		t.Fatalf("error copying data: %v", err)
	}
	if err := e.Flush(); err != nil {
		t.Fatalf("Error flushing: %q", err)
	}
	const expected = "<foo></foo>"
	if out := b.String(); out != expected {
		t.Fatalf("wrong output: want=%v, got=%v", expected, out)
	}
}

func TestWrapMallocs(t *testing.T) {
	s := xml.StartElement{
		Name: xml.Name{Local: "test"},
	}
	allocs := testing.AllocsPerRun(1000, func() {
		_ = xmlstream.Wrap(nil, s)
	})

	const expected = 0
	if allocs != expected {
		t.Fatalf("Too many allocations want=%d, got=%f", expected, allocs)
	}
}
