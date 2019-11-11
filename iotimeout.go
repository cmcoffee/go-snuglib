/*
	Package iotimeout provides a configurable timeout for io.Reader and io.ReadCloser.
*/

package iotimeout

import (
	"errors"
	"io"
	"sync/atomic"
	"time"
)

var ErrReadTimeout = errors.New("IO timeout exceeded waiting for bytes.")

const (
	working = 1 << iota
	waiting
	halt
)

// Timer for io tranfer
func start_timer(timeout time.Duration, flag *int32, expired chan struct{}) {
	timeout_seconds := int64(timeout.Round(time.Second).Seconds())

	var cnt int64

	for {
		time.Sleep(time.Second)
		switch atomic.LoadInt32(flag) {
		case working:
			cnt = 0
			atomic.StoreInt32(flag, waiting)
		case waiting:
			cnt++
			if cnt >= timeout_seconds {
				expired <- struct{}{}
				break
			}
		case halt:
			break
		}
	}
}

type resp struct {
	n   int
	err error
}

// Timeout Reader.
type Reader struct {
	flag    int32
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
	t.input = make(chan []byte, 1)
	t.output = make(chan resp, 1)
	t.expired = make(chan struct{}, 1)

	go start_timer(timeout, &t.flag, t.expired)

	go func() {
		var data resp
		for {
			data.n, data.err = reader.Read(<-t.input)
			t.output <- data
			if data.err != nil {
				break
			}
		}
	}()
	return t
}

// Time Sensitive Read function.
func (t *Reader) Read(p []byte) (n int, err error) {
	t.input <- p

	select {
	case data := <-t.output:
		n = data.n
		err = data.err
	case <-t.expired:
		return -1, ErrReadTimeout
	}
	if err == nil {
		atomic.StoreInt32(&t.flag, working)
	} else {
		atomic.StoreInt32(&t.flag, halt)
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
	return t.closerFunc()
}
