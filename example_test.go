// Copyright 2016 Sam Whited.
// Use of this source code is governed by the BSD 2-clause license that can be
// found in the LICENSE file.

package xmlstream_test

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"

	"mellium.im/xmlstream"
)

func ExampleIndent() {
	tokenizer := xmlstream.Indent(xml.NewDecoder(strings.NewReader(`
<quote>  <p>
    How now, my hearts! did you never see the picture
    of 'we three'?</p>
</quote>`)), xmlstream.Prefix("\n"), xmlstream.Indentation("    "))

	buf := new(bytes.Buffer)
	e := xml.NewEncoder(buf)
	for t, err := tokenizer.Token(); err == nil; t, err = tokenizer.Token() {
		e.EncodeToken(t)
	}
	e.Flush()
	fmt.Println(buf.String())
	// Output:
	// <quote>
	//     <p>
	//     How now, my hearts! did you never see the picture
	//     of &#39;we three&#39;?
	//     </p>
	// </quote>
}

func ExampleRemove() {
	removequote := xmlstream.Remove(func(t xml.Token) bool {
		switch tok := t.(type) {
		case xml.StartElement:
			if tok.Name.Local == "quote" {
				return true
			}
		case xml.EndElement:
			if tok.Name.Local == "quote" {
				return true
			}
		}
		return false
	})

	tokenizer := removequote(xml.NewDecoder(strings.NewReader(`
<quote>
  <p>Foolery, sir, does walk about the orb, like the sun; it shines everywhere.</p>
</quote>`)))

	buf := new(bytes.Buffer)
	e := xml.NewEncoder(buf)
	for t, err := tokenizer.Token(); err == nil; t, err = tokenizer.Token() {
		e.EncodeToken(t)
	}
	e.Flush()
	fmt.Println(buf.String())
	// Output:
	// <p>Foolery, sir, does walk about the orb, like the sun; it shines everywhere.</p>
}

func ExampleRemoveElement() {
	removeen := xmlstream.RemoveElement(func(start xml.StartElement) bool {
		// TODO: Probably be more specific and actually check the name.
		if len(start.Attr) > 0 && start.Attr[0].Value == "en" {
			return true
		}
		return false
	})

	tokenizer := removeen(xml.NewDecoder(strings.NewReader(`
<quote>
<p xml:lang="en">Thus the whirligig of time brings in his revenges.</p>
<p xml:lang="fr">et c’est ainsi que la roue du temps amène les occasions de revanche.</p>
</quote>
`)))

	buf := new(bytes.Buffer)
	e := xml.NewEncoder(buf)
	for t, err := tokenizer.Token(); err == nil; t, err = tokenizer.Token() {
		e.EncodeToken(t)
	}
	e.Flush()
	fmt.Println(buf.String())
	// Output:
	// <quote>
	//
	// <p xml:lang="fr">et c’est ainsi que la roue du temps amène les occasions de revanche.</p>
	// </quote>
}
