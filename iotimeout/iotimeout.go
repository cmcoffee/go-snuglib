/*
	Package iotimeout provides a configurable timeout for io.Reader and io.ReadCloser.
*/

package iotimeout

import (
	"errors"
	. "github.com/cmcoffee/go-snuglib/bitflag"
	"io"
	"time"
)

var ErrReadTimeout = errors.New("Timeout exceeded waiting for bytes.")

const (
	waiting = 1 << iota
	halted
)

// Timer for io tranfer
func start_timer(timeout time.Duration, flag *BitFlag, input chan []byte, expired chan struct{}) {
	timeout_seconds := int64(timeout.Round(time.Second).Seconds())

	var cnt int64

	for {
		time.Sleep(time.Second)

		if flag.Has(halted) {
			input <- nil
			break
		}

		if flag.Has(waiting) {
			cnt++
			if timeout_seconds > 0 && cnt >= timeout_seconds {
				expired <- struct{}{}
				flag.Set(halted)
			}
		} else {
			cnt = 0
			flag.Set(waiting)
		}
	}
}

type resp struct {
	n   int
	err error
}

// Timeout Reader.
type Reader struct {
	src     io.Reader
	flag    BitFlag
	input   chan []byte
	output  chan resp
	expired chan struct{}
}

// Timeout ReadCloser
type ReadCloser struct {
	*Reader
	closerFunc func() error
}

// Timeout ReadCloser: Adds a timer to io.Reader
func NewReader(reader io.Reader, timeout time.Duration) *Reader {
	t := new(Reader)
	t.src = reader
	t.input = make(chan []byte, 1)
	t.output = make(chan resp, 0)
	t.expired = make(chan struct{}, 0)

	go start_timer(timeout, &t.flag, t.input, t.expired)

	go func() {
		var (
			data resp
			p    []byte
		)
		for {
			p = <-t.input
			if p == nil {
				break
			}
			t.flag.Unset(waiting)
			data.n, data.err = reader.Read(p)
			t.output <- data
		}
	}()
	return t
}

// Time Sensitive Read function.
func (t *Reader) Read(p []byte) (n int, err error) {
	if t.flag.Has(halted) {
		return t.src.Read(p)
	}
	t.input <- p

	select {
	case data := <-t.output:
		n = data.n
		err = data.err
	case <-t.expired:
		t.flag.Set(halted)
		return -1, ErrReadTimeout
	}
	if err != nil {
		t.flag.Set(halted)
	}
	return
}

// Timeout ReadCloser: Adds a timer to io.ReadCloser
func NewReadCloser(readcloser io.ReadCloser, timeout time.Duration) *ReadCloser {
	t := NewReader(readcloser, timeout)
	return &ReadCloser{t, readcloser.Close}
}

// Close function for ReadCloser.
func (t *ReadCloser) Close() (err error) {
	t.flag.Set(halted)
	return t.closerFunc()
}
