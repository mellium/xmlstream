// Copyright 2019 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause license that can be
// found in the LICENSE file.

package xmlstream_test

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"mellium.im/xmlstream"
)

var (
	aStart   = xml.StartElement{Name: xml.Name{Local: "a"}, Attr: []xml.Attr{}}
	fooStart = xml.StartElement{Name: xml.Name{Local: "foo"}, Attr: []xml.Attr{}}
)

var readAllTests = [...]struct {
	in  xml.TokenReader
	out []xml.Token
	err error
}{
	0: {in: xml.NewDecoder(strings.NewReader(`<a></a>`)), out: []xml.Token{aStart, aStart.End()}},
	1: {in: xml.NewDecoder(strings.NewReader(`<a>a</a>`)), out: []xml.Token{aStart, xml.CharData("a"), aStart.End()}},
	2: {in: xml.NewDecoder(strings.NewReader(`<a>a</a><foo/>`)), out: []xml.Token{
		aStart, xml.CharData("a"), aStart.End(),
		fooStart, fooStart.End(),
	}},
	3: {
		// Make sure that we don't try to encode nil tokens or enter an infinite
		// loop when TokenDecoders return nil, nil.
		in:  xml.NewTokenDecoder(xmlstream.Wrap(nil, xml.StartElement{Name: xml.Name{Local: "a"}})),
		out: []xml.Token{aStart, aStart.End()},
	},
}

func TestReadAll(t *testing.T) {
	for i, tc := range readAllTests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			toks, err := xmlstream.ReadAll(tc.in)
			if err != tc.err {
				t.Fatalf("Unexpected error: want=%q, got=%q", tc.err, err)
			}

			if len(toks) != len(tc.out) {
				t.Fatalf("Unexpected output:\nwant=%+v,\n got=%+v", tc.out, toks)
			}

			for i, tok := range toks {
				// This is terrible, but it was the quickest way I could think to
				// compare tokens.
				if fmt.Sprintf("%#v", tok) != fmt.Sprintf("%#v", tc.out[i]) {
					t.Errorf("Unexpected token %d:\nwant=%#v,\n got=%#v", i, tc.out[i], tok)
				}
			}
		})
	}
}
