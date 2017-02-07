// Copyright 2016 Sam Whited.
// Use of this source code is governed by the BSD 2-clause license that can be
// found in the LICENSE file.

package xmlstream

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

var innerReaderTests = [...]struct {
	R    io.Reader
	Read string
	Err  error
}{
	0: {
		R:    strings.NewReader(``),
		Read: ``,
		Err:  nil,
	},
	1: {
		R:    strings.NewReader(`<test></test>`),
		Read: ``,
		Err:  nil,
	},
	2: {
		R:    strings.NewReader(`<test><inner/></test>`),
		Read: `<inner/>`,
		Err:  nil,
	},
	3: {
		R:    strings.NewReader(`<test>Inner</test>`),
		Read: `Inner`,
		Err:  nil,
	},
	4: {
		R:    strings.NewReader(`<test>Inner</oops>`),
		Read: `Inner`,
		Err:  unexpectedEndError{"oops"},
	},
	5: {
		R:    strings.NewReader(`<stream xmlns="stream">Test</stream>`),
		Read: `Test`,
		Err:  nil,
	},
	6: {
		R:    strings.NewReader(`<stream:stream><stream:features></stream:stream>`),
		Read: `<stream:features>`,
		Err:  nil,
	},
	7: {
		R:    strings.NewReader(`<stream:stream><stream:features> <stream:stream>`),
		Read: `<stream:features>`,
		Err:  notEndError,
	},
	8: {
		R:    strings.NewReader(`<stream:stream><stream:features><stream:stream>`),
		Read: `<stream:features`,
		Err:  notEndError,
	},
	9: {
		R:    strings.NewReader(`</stream:stream>`),
		Read: ``,
		Err:  notStartError,
	},
	10: {
		R:    strings.NewReader(`<!-- Test -->`),
		Read: ``,
		Err:  notStartError,
	},
	11: {
		R:    strings.NewReader(`What is dis junk?`),
		Read: ``,
		Err:  notStartError,
	},
	12: {
		R:    strings.NewReader(`<test/>`),
		Read: ``,
		Err:  nil,
	},
	13: {
		R:    strings.NewReader(`<test:test/>`),
		Read: ``,
		Err:  nil,
	},
}

func TestInnerReader(t *testing.T) {
	for i, tc := range innerReaderTests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			ir := InnerReader(tc.R)
			if ir == nil {
				t.Fatal("InnerReader returned nil reader")
			}
			b, err := ioutil.ReadAll(ir)
			if err != tc.Err {
				t.Errorf("Unxpected error: want=`%v`, got=`%v`", tc.Err, err)
			}
			if string(b) != tc.Read {
				t.Errorf("Unexpected value read: want=`%s`, got=`%s`", tc.Read, b)
			}
		})
	}
}
