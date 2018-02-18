// Copyright 2017 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.
//
// Code in this file copied from the Go io package:
//
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.GO file.

package xmlstream

import (
	"encoding/xml"
	"errors"
	"io"
	"sync"
)

// TODO: For now I've just copied this from io.pipe, but tokens are fundamentally
//       different than byte slices. It would be nice to add some sort of
//       buffering in the future so that multiple calls to write can happen
//       before we have to call read. This would also give a purpose to the noop
//       Flush() method. A buffer size of zero would behave exactly as this does
//       now, so I think this can be done in a backwards compatible way (the
//       only API change would be the addition of a BufferedPipe(n int) method).

// A pipe is the shared pipe structure underlying PipeReader and PipeWriter.
type pipe struct {
	rl    sync.Mutex // gates readers one at a time
	wl    sync.Mutex // gates writers one at a time
	l     sync.Mutex // protects remaining fields
	data  xml.Token  // data remaining in pending read/write
	rwait sync.Cond  // waiting reader
	wwait sync.Cond  // waiting writer
	rerr  error      // if reader closed, error to give writes
	werr  error      // if writer closed, error to give reads
}

func (p *pipe) Token() (t xml.Token, err error) {
	// One reader at a time.
	p.rl.Lock()
	defer p.rl.Unlock()

	p.l.Lock()
	defer p.l.Unlock()
	for {
		if p.rerr != nil {
			return nil, ErrClosedPipe
		}
		if p.data != nil {
			break
		}
		if p.werr != nil {
			return nil, p.werr
		}
		p.rwait.Wait()
	}
	t = p.data
	p.data = nil
	p.wwait.Signal()
	return t, err
}

func (p *pipe) EncodeToken(t xml.Token) (err error) {
	// One writer at a time.
	p.wl.Lock()
	defer p.wl.Unlock()
	p.l.Lock()
	defer p.l.Unlock()
	if p.werr != nil {
		err = ErrClosedPipe
		return err
	}
	p.data = t
	p.rwait.Signal()
	for {
		if p.data == nil {
			break
		}
		if p.rerr != nil {
			err = p.rerr
			break
		}
		if p.werr != nil {
			err = ErrClosedPipe
			break
		}
		p.wwait.Wait()
	}
	p.data = nil // in case of rerr or werr
	return err
}

func (p *pipe) rclose(err error) {
	if err == nil {
		err = ErrClosedPipe
	}
	p.l.Lock()
	defer p.l.Unlock()
	p.rerr = err
	p.rwait.Signal()
	p.wwait.Signal()
}

func (p *pipe) wclose(err error) {
	if err == nil {
		err = io.EOF
	}
	p.l.Lock()
	defer p.l.Unlock()
	p.werr = err
	p.rwait.Signal()
	p.wwait.Signal()
}

// ErrClosedPipe is the error used for read or write operations on a closed
// pipe.
var ErrClosedPipe = errors.New("xmlstream: read/write on closed pipe")

// A PipeReader is the read half of a token pipe.
type PipeReader struct {
	p *pipe
}

// Token implements the TokenReader interface.
// It reads a token from the pipe, blocking until a writer arrives or the write
// end is closed. If the write end is closed with an error, that error is
// returned as err; otherwise err is io.EOF.
func (r *PipeReader) Token() (t xml.Token, err error) {
	return r.p.Token()
}

// Close closes the PipeReader; subsequent reads from the read half of the pipe
// will return no bytes and EOF.
func (r *PipeReader) Close() error {
	r.CloseWithError(nil)
	return nil
}

// CloseWithError closes the PipeReader; subsequent reads from the read half of
// the pipe will return no tokens and the error err, or EOF if err is nil.
func (r *PipeReader) CloseWithError(err error) {
	r.p.rclose(err)
}

// A PipeWriter is the write half of a token pipe.
type PipeWriter struct {
	p *pipe
}

// EncodeToken implements the TokenWriter interface.
// It writes a token to the pipe, blocking until one or more readers have
// consumed all the data or the read end is closed.
// If the read end is closed with an error, that err is returned as err;
// otherwise err is ErrClosedPipe.
func (w *PipeWriter) EncodeToken(t xml.Token) error {
	return w.p.EncodeToken(t)
}

// Flush is currently a noop and always returns nil.
func (w *PipeWriter) Flush() error {
	return nil
}

// Close closes the PipeWriter; subsequent reads from the read half of the pipe
// will return no bytes and EOF.
func (w *PipeWriter) Close() error {
	w.CloseWithError(nil)
	return nil
}

// CloseWithError closes the PipeWriter; subsequent reads from the read half of
// the pipe will return no tokens and the error err, or EOF if err is nil.
func (w *PipeWriter) CloseWithError(err error) {
	w.p.wclose(err)
}

// Pipe creates a synchronous in-memory pipe of tokens.
// It can be used to connect code expecting an TokenReader
// with code expecting an xmlstream.TokenWriter.
//
// Reads and Writes on the pipe are matched one to one.
// That is, each Write to the PipeWriter blocks until it has satisfied a Read
// from the corresponding PipeReader.
//
// It is safe to call Read and Write in parallel with each other or with Close.
// Parallel calls to Read and parallel calls to Write are also safe:
// the individual calls will be gated sequentially.
func Pipe() (*PipeReader, *PipeWriter) {
	p := new(pipe)
	p.rwait.L = &p.l
	p.wwait.L = &p.l
	r := &PipeReader{p}
	w := &PipeWriter{p}
	return r, w
}
