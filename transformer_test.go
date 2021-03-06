// Copyright 2016 The Mellium Contributors.
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

var _ xmlstream.Encoder = (*xml.Encoder)(nil)
var _ xmlstream.Decoder = (*xml.Decoder)(nil)

func TestDiscard(t *testing.T) {
	d := xmlstream.Discard()
	if err := d.EncodeToken(xml.StartElement{}); err != nil {
		t.Errorf("Unexpected error while discarding token: %v", err)
	}
}

type tokenizerTest struct {
	Transform xmlstream.Transformer
	Input     string
	Output    string
	Err       bool
}

func runTests(t *testing.T, tcs []tokenizerTest) {
	for i, tc := range tcs {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			d := tc.Transform(xml.NewDecoder(strings.NewReader(tc.Input)))
			buf := new(bytes.Buffer)
			e := xml.NewEncoder(buf)
			var tok xml.Token
			var err error
		decodeloop:
			for err == nil {
				tok, err = d.Token()
				switch {
				case err != nil && err != io.EOF && !tc.Err:
					t.Fatalf("Unexpected error decoding token: %s", err)
				case err == io.EOF:
					break decodeloop
				}
				err = e.EncodeToken(tok)
			}
			e.Flush()
			switch {
			case err != nil && err != io.EOF && tc.Err:
			case (err == io.EOF || err == nil) && tc.Err:
				t.Fatal("Expected error, but did not get one")
			case buf.String() != tc.Output:
				t.Fatalf("got=`%s`, want=`%s`", buf.String(), tc.Output)
			}
		})
	}
}

func TestInspect(t *testing.T) {
	tokens := 0
	inspector := xmlstream.Inspect(func(t xml.Token) {
		tokens++
	})
	d := inspector(xml.NewDecoder(strings.NewReader(`<quote>Now Jove,<br/>in his next commodity of hair, send thee a beard!</quote>`)))
	for tok, err := d.Token(); err != io.EOF; tok, err = d.Token() {
		if start, ok := tok.(xml.StartElement); ok && start.Name.Local == "br" {
			if err = xmlstream.Skip(d); err == io.EOF {
				break
			}
		}
	}
	if tokens != 6 {
		t.Fatalf("Got %d tokens but expected 6", tokens)
	}
}

func TestRemove(t *testing.T) {
	runTests(t, []tokenizerTest{
		{
			Transform: xmlstream.Remove(func(t xml.Token) bool { return true }),
			Input:     `<quote>Foolery, sir, does walk about the orb like the sun,<br/>it shines every where.</quote>`,
			Output:    ``,
			Err:       false,
		},
		{
			Transform: xmlstream.Remove(func(t xml.Token) bool {
				if _, ok := t.(xml.CharData); ok {
					return false
				}
				return true
			}),
			Input:  `<quote>Now Jove, in his next commodity of hair, send thee a beard!</quote>`,
			Output: `Now Jove, in his next commodity of hair, send thee a beard!`,
			Err:    false,
		},
		{
			Transform: xmlstream.Remove(func(t xml.Token) bool {
				if _, ok := t.(xml.StartElement); ok {
					return true
				}
				return false
			}),
			Input:  `<quote>Now Jove, in his next commodity of hair, send thee a beard!</quote>`,
			Output: ``,
			Err:    true,
		},
	})
}

func TestMap(t *testing.T) {
	runTests(t, []tokenizerTest{
		{
			Transform: xmlstream.Map(func(t xml.Token) xml.Token {
				switch tok := t.(type) {
				case xml.StartElement:
					if tok.Name.Local == "quote" {
						tok.Name.Local = "blocking"
						return tok
					}
				case xml.EndElement:
					if tok.Name.Local == "quote" {
						tok.Name.Local = "blocking"
						return tok
					}
				}
				return t
			}),
			Input:  `<quote>[Re-enter Clown with a letter, and FABIAN]</quote>`,
			Output: `<blocking>[Re-enter Clown with a letter, and FABIAN]</blocking>`,
			Err:    false,
		},
	})
}

func TestRemoveAttr(t *testing.T) {
	runTests(t, []tokenizerTest{
		{
			Transform: xmlstream.RemoveAttr(func(_ xml.StartElement, _ xml.Attr) bool {
				return false
			}),
			Input:  `<quote act="1" scene="5" character="Feste">Let her hang me: he that is well hanged in this world needs to fear no colours.</quote>`,
			Output: `<quote act="1" scene="5" character="Feste">Let her hang me: he that is well hanged in this world needs to fear no colours.</quote>`,
			Err:    false,
		},
		{
			Transform: xmlstream.RemoveAttr(func(_ xml.StartElement, _ xml.Attr) bool {
				return true
			}),
			Input:  `<!-- Quote --><quote act="1" scene="5" cite="Feste">Let her hang me: he that is well hanged in this world needs to fear no colours.</quote>`,
			Output: `<!-- Quote --><quote>Let her hang me: he that is well hanged in this world needs to fear no colours.</quote>`,
			Err:    false,
		},
		{
			Transform: xmlstream.RemoveAttr(func(_ xml.StartElement, a xml.Attr) bool {
				return a.Name.Local == "cite"
			}),
			Input:  `<quote act="1" cite="Feste" scene="5">Let her hang me: he that is well hanged in this world needs to fear no colours.</quote>`,
			Output: `<quote act="1" scene="5">Let her hang me: he that is well hanged in this world needs to fear no colours.</quote>`,
			Err:    false,
		},
		{
			Transform: xmlstream.RemoveAttr(func(start xml.StartElement, a xml.Attr) bool {
				return start.Name.Local == "quote" && a.Name.Local == "cite"
			}),
			Input:  `<text cite="Feste"></text><quote act="1" cite="Feste" scene="5">Let her hang me: he that is well hanged in this world needs to fear no colours.</quote>`,
			Output: `<text cite="Feste"></text><quote act="1" scene="5">Let her hang me: he that is well hanged in this world needs to fear no colours.</quote>`,
			Err:    false,
		},
	})
}

var insertTestCases = [...]struct {
	name xml.Name
	in   string
	out  string
}{
	0: {},
	1: {
		name: xml.Name{Space: "jabber:client", Local: "message"},
		in:   `<message xmlns="jabber:client"/>`,
		out:  `<message xmlns="jabber:client"><test></test></message>`,
	},
	2: {
		name: xml.Name{Space: "jabber:server", Local: "message"},
		in:   `<message xmlns="jabber:server"/><message xmlns="jabber:server"><body>test</body></message><message></message>`,
		out:  `<message xmlns="jabber:server"><test></test></message><message xmlns="jabber:server"><body xmlns="jabber:server">test</body><test></test></message><message></message>`,
	},
	3: {
		name: xml.Name{Space: "jabber:server", Local: "message"},
		in:   `<message xmlns="jabber:badns"/>`,
		out:  `<message xmlns="jabber:badns"></message>`,
	},
	4: {
		name: xml.Name{Space: "", Local: "message"},
		in:   `<message xmlns="urn:example"/><message/>`,
		out:  `<message xmlns="urn:example"><test></test></message><message><test></test></message>`,
	},
	5: {
		name: xml.Name{Space: "urn:example", Local: ""},
		in:   `<message xmlns="urn:example"/><presence/><iq xmlns="urn:example"/>`,
		out:  `<message xmlns="urn:example"><test></test></message><presence></presence><iq xmlns="urn:example"><test></test></iq>`,
	},
	6: {
		name: xml.Name{Space: "", Local: ""},
		in:   `<message xmlns="urn:example"/><presence/><iq xmlns="urn:example"/>`,
		out:  `<message xmlns="urn:example"><test></test></message><presence><test></test></presence><iq xmlns="urn:example"><test></test></iq>`,
	},
}

type testPayload struct{}

func (testPayload) TokenReader() xml.TokenReader {
	return xmlstream.Wrap(
		nil,
		xml.StartElement{Name: xml.Name{Local: "test"}},
	)
}

func TestInsert(t *testing.T) {
	for i, tc := range insertTestCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			inserter := xmlstream.Insert(tc.name, testPayload{})
			r := inserter(xml.NewDecoder(strings.NewReader(tc.in)))
			// Prevent duplicate xmlns attributes. See https://mellium.im/issue/75
			r = xmlstream.RemoveAttr(func(start xml.StartElement, attr xml.Attr) bool {
				return (start.Name.Local == "message" || start.Name.Local == "iq") && attr.Name.Local == "xmlns"
			})(r)
			var buf strings.Builder
			e := xml.NewEncoder(&buf)
			_, err := xmlstream.Copy(e, r)
			if err != nil {
				t.Fatalf("error encoding: %v", err)
			}
			if err = e.Flush(); err != nil {
				t.Fatalf("error flushing: %v", err)
			}

			if out := buf.String(); tc.out != out {
				t.Errorf("wrong output:\nwant=%s,\n got=%s", tc.out, out)
			}
		})
	}
}

var insertFTestCases = [...]struct {
	f   func(xml.StartElement, uint64, xmlstream.TokenWriter) error
	in  string
	out string
}{
	0: {},
	1: {
		f: func(start xml.StartElement, depth uint64, w xmlstream.TokenWriter) error {
			return nil
		},
		in:  `<message foo="bar"><foo/></message>`,
		out: `<message foo="bar"><foo></foo></message>`,
	},
	2: {
		in:  `<message foo="bar"><foo/></message>`,
		out: `<message foo="bar"><foo></foo></message>`,
	},
	3: {
		f: func(start xml.StartElement, depth uint64, w xmlstream.TokenWriter) error {
			start = xml.StartElement{Name: xml.Name{Local: "test"}}
			w.EncodeToken(start)
			w.EncodeToken(start.End())
			return nil
		},
		in:  `<message foo="bar"><foo/></message>`,
		out: `<message foo="bar"><test></test><foo><test></test></foo></message>`,
	},
	4: {
		f: func(start xml.StartElement, depth uint64, w xmlstream.TokenWriter) error {
			if depth == 2 {
				start = xml.StartElement{Name: xml.Name{Local: "test"}}
				w.EncodeToken(start)
				w.EncodeToken(start.End())
			}
			return nil
		},
		in:  `<message foo="bar"><foo/></message>`,
		out: `<message foo="bar"><foo><test></test></foo></message>`,
	},
}

func TestInsertFunc(t *testing.T) {
	for i, tc := range insertFTestCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			inserter := xmlstream.InsertFunc(tc.f)
			r := inserter(xml.NewDecoder(strings.NewReader(tc.in)))
			var buf strings.Builder
			e := xml.NewEncoder(&buf)
			_, err := xmlstream.Copy(e, r)
			if err != nil {
				t.Fatalf("error encoding: %v", err)
			}
			if err = e.Flush(); err != nil {
				t.Fatalf("error flushing: %v", err)
			}

			if out := buf.String(); tc.out != out {
				t.Errorf("wrong output:\nwant=%s,\n got=%s", tc.out, out)
			}
		})
	}
}
