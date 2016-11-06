// Copyright 2016 Sam Whited.
// Use of this source code is governed by the BSD 2-clause license that can be
// found in the LICENSE file.

// Package xmlstream provides an API for streaming, transforming, and otherwise
// manipulating XML data.
package xmlstream // import "mellium.im/xmlstream"

import (
	"encoding/xml"
)

// A Tokenizer is anything that can decode a stream of XML tokens.
type Tokenizer interface {
	Token() (xml.Token, error)
	Skip() error
}

// A Transformer returns a new tokenizer that wraps src and optionally
// transforms any tokens read from it.
type Transformer func(src Tokenizer) Tokenizer
