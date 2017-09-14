// Copyright 2016 Sam Whited.
// Use of this source code is governed by the BSD 2-clause license that can be
// found in the LICENSE file.

// Package xmlstream provides an API for streaming, transforming, and otherwise
// manipulating XML data.
//
// If you are using Go built from source, you will need a build from 2017-09-13
// or later that includes this patch: https://golang.org/cl/38791
// When Go 1.10 is released, this package will be modified to use the
// xml.TokenReader interface, the Go 1.9 shim (which only exists to let
// godoc.org generate documentation) will be removed, and a 1.0 release will be
// made.
//
// BE ADVISED: The API is unstable and subject to change.
package xmlstream // import "mellium.im/xmlstream"
