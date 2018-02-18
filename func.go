// Copyright 2017 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package xmlstream

import (
	"encoding/xml"
)

// ReaderFunc type is an adapter to allow the use of ordinary functions as an
// TokenReader. If f is a function with the appropriate signature,
// ReaderFunc(f) is an TokenReader that calls f.
type ReaderFunc func() (xml.Token, error)

// Token calls f.
func (f ReaderFunc) Token() (xml.Token, error) {
	return f()
}
