// Copyright 2017 Sam Whited.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

// +build !go1.10

package xmlstream_test

import (
	"encoding/xml"
	"log"
	"os"
	"strings"

	"mellium.im/xmlstream"
)

func ExampleWrap() {
	var r xmlstream.TokenReader = xml.NewDecoder(strings.NewReader(`<body>No mind that ever lived stands firm in evil days, but goes astray.</body>`))
	e := xml.NewEncoder(os.Stdout)
	e.Indent("", "  ")

	r = xmlstream.Wrap(r, xml.StartElement{
		Name: xml.Name{Local: "message"},
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "from"}, Value: "ismene@example.org/Fo6Eeb2e"},
		},
	})

	if err := xmlstream.Copy(e, r); err != nil {
		log.Fatal("Error in wrap example:", err)
	}
	// Output:
	// <message from="ismene@example.org/Fo6Eeb2e">
	//   <body>No mind that ever lived stands firm in evil days, but goes astray.</body>
	// </message>
}

func ExampleUnwrap() {
	var r xmlstream.TokenReader = xml.NewDecoder(strings.NewReader(`<message from="ismene@example.org/dIoK6Wi3"><body>No mind that ever lived stands firm in evil days, but goes astray.</body></message>`))
	e := xml.NewEncoder(os.Stdout)

	r = xmlstream.Unwrap(r)

	if err := xmlstream.Copy(e, r); err != nil {
		log.Fatal("Error in unwrap example:", err)
	}
	// Output:
	// <body>No mind that ever lived stands firm in evil days, but goes astray.</body>
}

func ExampleLimitReader() {
	e := xml.NewEncoder(os.Stdout)
	var r xmlstream.TokenReader = xml.NewDecoder(strings.NewReader(`<one>One hen</one><two>Two ducks</two>`))

	r = xmlstream.LimitReader(r, 3)

	if err := xmlstream.Copy(e, r); err != nil {
		log.Fatal("Error in LimitReader example:", err)
	}

	// Output:
	// <one>One hen</one>
}
