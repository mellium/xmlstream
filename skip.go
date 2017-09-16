// Copyright 2017 Sam Whited.
// Use of this source code is governed by the BSD
// 2-clause license that can be found in the LICENSE file.
//
// Code in this file copied from the Go io package:
//
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE.GO file.

package xmlstream

import (
	"encoding/xml"
)

// Skip reads tokens until it has consumed the end element matching the most
// recent start element already consumed.
// It recurs if it encounters a start element, so it can be used to skip nested
// structures.
// It returns nil if it finds an end element at the same nesting level as the
// start element; otherwise it returns an error describing the problem.
// Skip does not verify that the start and end elements match.
func Skip(r TokenReader) error {
	for {
		tok, err := r.Token()
		if err != nil {
			return err
		}
		switch tok.(type) {
		case xml.StartElement:
			if err := Skip(r); err != nil {
				return err
			}
		case xml.EndElement:
			return nil
		}
	}
}
