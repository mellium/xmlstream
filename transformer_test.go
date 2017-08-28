// Copyright 2016 Sam Whited.
// Use of this source code is governed by the BSD 2-clause license that can be
// found in the LICENSE file.

package xmlstream

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"testing"
)

type tokenizerTest struct {
	Transform Transformer
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
	inspector := Inspect(func(t xml.Token) {
		tokens++
	})
	d := inspector(xml.NewDecoder(strings.NewReader(`<quote>Now Jove,<br/>in his next commodity of hair, send thee a beard!</quote>`)))
	for tok, err := d.Token(); err != io.EOF; tok, err = d.Token() {
		if start, ok := tok.(xml.StartElement); ok && start.Name.Local == "br" {
			if err = Skip(d); err == io.EOF {
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
			Transform: Remove(func(t xml.Token) bool { return true }),
			Input:     `<quote>Foolery, sir, does walk about the orb like the sun,<br/>it shines every where.</quote>`,
			Output:    ``,
			Err:       false,
		},
		{
			Transform: Remove(func(t xml.Token) bool {
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
			Transform: Remove(func(t xml.Token) bool {
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
			Transform: Map(func(t xml.Token) xml.Token {
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
			Transform: RemoveAttr(func(_ xml.StartElement, _ xml.Attr) bool {
				return false
			}),
			Input:  `<quote act="1" scene="5" character="Feste">Let her hang me: he that is well hanged in this world needs to fear no colours.</quote>`,
			Output: `<quote act="1" scene="5" character="Feste">Let her hang me: he that is well hanged in this world needs to fear no colours.</quote>`,
			Err:    false,
		},
		{
			Transform: RemoveAttr(func(_ xml.StartElement, _ xml.Attr) bool {
				return true
			}),
			Input:  `<!-- Quote --><quote act="1" scene="5" cite="Feste">Let her hang me: he that is well hanged in this world needs to fear no colours.</quote>`,
			Output: `<!-- Quote --><quote>Let her hang me: he that is well hanged in this world needs to fear no colours.</quote>`,
			Err:    false,
		},
		{
			Transform: RemoveAttr(func(_ xml.StartElement, a xml.Attr) bool {
				return a.Name.Local == "cite"
			}),
			Input:  `<quote act="1" cite="Feste" scene="5">Let her hang me: he that is well hanged in this world needs to fear no colours.</quote>`,
			Output: `<quote act="1" scene="5">Let her hang me: he that is well hanged in this world needs to fear no colours.</quote>`,
			Err:    false,
		},
		{
			Transform: RemoveAttr(func(start xml.StartElement, a xml.Attr) bool {
				return start.Name.Local == "quote" && a.Name.Local == "cite"
			}),
			Input:  `<text cite="Feste"></text><quote act="1" cite="Feste" scene="5">Let her hang me: he that is well hanged in this world needs to fear no colours.</quote>`,
			Output: `<text cite="Feste"></text><quote act="1" scene="5">Let her hang me: he that is well hanged in this world needs to fear no colours.</quote>`,
			Err:    false,
		},
	})
}
