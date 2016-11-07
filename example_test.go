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
