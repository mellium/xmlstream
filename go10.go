// Copyright 2017 Sam Whited.
// Use of this source code is governed by the BSD 2-clause license that can be
// found in the LICENSE file.

// +build go1.10

package xmlstream

import (
	"encoding/xml"
)

// A TokenReader is anything that can decode a stream of XML tokens, including
// a Decoder.
// For more information see the documentation for xml.TokenReader.
type TokenReader = xml.TokenReader
