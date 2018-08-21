// Copyright 2016 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package xmlstream_test

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"mellium.im/xmlstream"
)

func ExampleSkip() {
	e := xml.NewEncoder(os.Stdout)

	r := xml.NewDecoder(strings.NewReader(`<par>I don't like to look out of the windows even—there are so many of those creeping women, and they creep so fast.</par><par>I wonder if they all come out of that wall paper, as I did?</par>`))

	r.Token() // <par>

	if err := xmlstream.Skip(r); err != nil && err != io.EOF {
		log.Fatal("Error in skipping par:", err)
	}
	if _, err := xmlstream.Copy(e, r); err != nil {
		log.Fatal("Error in Skip example:", err)
	}
	if err := e.Flush(); err != nil {
		log.Fatal("Error flushing:", err)
	}

	// Output:
	// <par>I wonder if they all come out of that wall paper, as I did?</par>
}

func ExampleMultiReader() {
	e := xml.NewEncoder(os.Stdout)
	e.Indent("", "  ")

	r1 := xml.NewDecoder(strings.NewReader(`<title>Dover Beach</title>`))
	r2 := xml.NewDecoder(strings.NewReader(`<author>Matthew Arnold</author>`))
	r3 := xml.NewDecoder(strings.NewReader(`<incipit>The sea is calm to-night.</incipit>`))

	r := xmlstream.MultiReader(r1, r2, r3)

	if _, err := xmlstream.Copy(e, r); err != nil {
		log.Fatal("Error in MultiReader example:", err)
	}
	if err := e.Flush(); err != nil {
		log.Fatal("Error flushing:", err)
	}
	// Output:
	// <title>Dover Beach</title>
	// <author>Matthew Arnold</author>
	// <incipit>The sea is calm to-night.</incipit>
}

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
	if _, err := xmlstream.Copy(e, d); err != nil {
		log.Fatal("Error in func example:", err)
	}
	if err := e.Flush(); err != nil {
		log.Fatal("Error flushing:", err)
	}
	// Output:
	// <quote>the rain it raineth every day</quote>
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
</quote>`)), xmlstream.Indent("    "))

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
	//
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

func ExampleWrap() {
	var r xml.TokenReader = xml.NewDecoder(strings.NewReader(`<body>No mind that ever lived stands firm in evil days, but goes astray.</body>`))
	e := xml.NewEncoder(os.Stdout)
	e.Indent("", "  ")

	r = xmlstream.Wrap(r, xml.StartElement{
		Name: xml.Name{Local: "message"},
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "from"}, Value: "ismene@example.org/Fo6Eeb2e"},
		},
	})

	if _, err := xmlstream.Copy(e, r); err != nil {
		log.Fatal("Error in wrap example:", err)
	}
	if err := e.Flush(); err != nil {
		log.Fatal("Error flushing:", err)
	}
	// Output:
	// <message from="ismene@example.org/Fo6Eeb2e">
	//   <body>No mind that ever lived stands firm in evil days, but goes astray.</body>
	// </message>
}

func ExampleUnwrap() {
	var r xml.TokenReader = xml.NewDecoder(strings.NewReader(`<message from="ismene@example.org/dIoK6Wi3"><body>No mind that ever lived stands firm in evil days, but goes astray.</body></message>`))
	e := xml.NewEncoder(os.Stdout)

	r, tok, err := xmlstream.Unwrap(r)
	if err != nil {
		log.Fatal("Error unwraping:", err)
	}

	fmt.Printf("%s:\n", tok.(xml.StartElement).Name.Local)
	if _, err := xmlstream.Copy(e, r); err != nil {
		log.Fatal("Error in unwrap example:", err)
	}
	if err := e.Flush(); err != nil {
		log.Fatal("Error flushing:", err)
	}

	// Output:
	// message:
	// <body>No mind that ever lived stands firm in evil days, but goes astray.</body>
}

func ExampleLimitReader() {
	e := xml.NewEncoder(os.Stdout)
	var r xml.TokenReader = xml.NewDecoder(strings.NewReader(`<one>One hen</one><two>Two ducks</two>`))

	r = xmlstream.LimitReader(r, 3)

	if _, err := xmlstream.Copy(e, r); err != nil {
		log.Fatal("Error in LimitReader example:", err)
	}
	if err := e.Flush(); err != nil {
		log.Fatal("Error flushing:", err)
	}

	// Output:
	// <one>One hen</one>
}
