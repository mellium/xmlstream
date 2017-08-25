// Copyright 2016 Sam Whited.
// Use of this source code is governed by the BSD 2-clause license that can be
// found in the LICENSE file.

package xmlstream_test

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"

	"mellium.im/xmlstream"
)

func ExampleReaderFunc() {
	state := 0
	start := xml.StartElement{Name: xml.Name{Local: "quote"}}
	d := xmlstream.ReaderFunc(func() (xml.Token, error) {
		switch state {
		case 0:
			state++
			return start, nil
		case 1:
			state++
			return xml.CharData("the rain it raineth every day"), nil
		case 2:
			state++
			return start.End(), nil
		default:
			return nil, io.EOF
		}
	})

	e := xml.NewEncoder(os.Stdout)
	xmlstream.Encode(e, d)
	// Output:
	// <quote>the rain it raineth every day</quote>
}

func ExampleEncode() {
	removequote := xmlstream.Remove(func(t xml.Token) bool {
		switch tok := t.(type) {
		case xml.StartElement:
			return tok.Name.Local == "quote"
		case xml.EndElement:
			return tok.Name.Local == "quote"
		}
		return false
	})

	e := xml.NewEncoder(os.Stdout)
	xmlstream.Encode(e, removequote(xml.NewDecoder(strings.NewReader(`
<quote>
  <p>Foolery, sir, does walk about the orb, like the sun; it shines everywhere.</p>
</quote>`))))
	// Output:
	// <p>Foolery, sir, does walk about the orb, like the sun; it shines everywhere.</p>
}

func ExampleInnerReader() {
	r := xmlstream.InnerReader(strings.NewReader(`<stream:features>
<starttls xmlns='urn:ietf:params:xml:ns:xmpp-tls'>
<required/>
</starttls>
</stream:features>`))
	io.Copy(os.Stdout, r)
	// Output:
	// <starttls xmlns='urn:ietf:params:xml:ns:xmpp-tls'>
	// <required/>
	// </starttls>
}

func ExampleFmt_indentation() {
	tokenizer := xmlstream.Fmt(xml.NewDecoder(strings.NewReader(`
<quote>  <p>
                 <!-- Chardata is not indented -->
  How now, my hearts! did you never see the picture
of 'we three'?</p>
</quote>`)), xmlstream.Prefix("\n"), xmlstream.Indent("    "))

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
	//         <!-- Chardata is not indented -->
	//   How now, my hearts! did you never see the picture
	// of &#39;we three&#39;?
	//     </p>
	// </quote>
}

func ExampleRemove() {
	removequote := xmlstream.Remove(func(t xml.Token) bool {
		switch tok := t.(type) {
		case xml.StartElement:
			return tok.Name.Local == "quote"
		case xml.EndElement:
			return tok.Name.Local == "quote"
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
	removeLangEn := xmlstream.RemoveElement(func(start xml.StartElement) bool {
		// TODO: Probably be more specific and actually check the name.
		if len(start.Attr) > 0 && start.Attr[0].Value == "en" {
			return true
		}
		return false
	})

	d := removeLangEn(xml.NewDecoder(strings.NewReader(`
<quote>
<p xml:lang="en">Thus the whirligig of time brings in his revenges.</p>
<p xml:lang="fr">et c’est ainsi que la roue du temps amène les occasions de revanche.</p>
</quote>
`)))

	buf := new(bytes.Buffer)
	e := xml.NewEncoder(buf)
	for t, err := d.Token(); err == nil; t, err = d.Token() {
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
