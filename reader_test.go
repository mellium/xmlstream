// Copyright 2016 Sam Whited.
// Use of this source code is governed by the BSD 2-clause license that can be
// found in the LICENSE file.

package xmlstream

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

var _ io.ByteReader = (*numReadReader)(nil)
var _ io.Reader = (*numReadReader)(nil)

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

func TestNumRead(t *testing.T) {
	r := strings.NewReader("12345671234567")
	// Read for the tests in two chunks of 7
	mr := io.MultiReader(
		io.LimitReader(r, 7),
		io.LimitReader(r, 7),
	)
	br := bufio.NewReader(mr)

	nrr := numReadReader{
		R: br,
	}

	// One more byte so that if we break the test it will fail anyways.
	for i := 0; i < 2; i++ {
		p := make([]byte, 8)
		n, err := nrr.Read(p)
		switch {
		case err != nil:
			t.Fatalf("Unexpected error while reading:", err)
		case n != 7:
			t.Fatalf("Read an unexpected ammount, want=7, got=%d", n)
		case p[0] != '1' || p[6] != '7':
			t.Fatalf("Expected to read want=1234567, got=%d", p)
		case nrr.TotalRead != 7*(i+1):
			t.Fatalf("Wrong value for total bytes read; want=%d, got=%d", 7*(i+1), nrr.TotalRead)
		}
	}
}

func TestNumReadByte(t *testing.T) {
	b := []byte{1, 2, 3, 4, 5, 6, 7}
	r := bytes.NewReader(b)
	nrr := numReadReader{
		R: r,
	}
	for i, v := range b {
		bb, err := nrr.ReadByte()
		switch {
		case err != nil:
			t.Fatalf("Unexpected error while reading:", err)
		case bb != v:
			t.Fatalf("Unexpected byte read; want=%d, got=%d", v, bb)
		case nrr.TotalRead != i+1:
			t.Fatalf("Unexpected value for TotalRead; want=%d, got=%d", i+1, nrr.TotalRead)
		}
	}
}
