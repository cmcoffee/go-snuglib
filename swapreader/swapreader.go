package swapreader

import (
	"io"
)

// Swap Reader allows for swapping the io.Reader backed []bytes
type Reader struct {
	from_reader    bool
	reader         io.Reader
	decoder_bytes  []byte
	decoder_copied int
}

// Set []byte for reader
func (r *Reader) SetBytes(in []byte) {
	r.from_reader = false
	r.decoder_bytes = in
	r.decoder_copied = 0
}

// Set Reader to Reader
func (r *Reader) SetReader(in io.Reader) {
	r.from_reader = true
	r.reader = in
}

// swap_reader Read function.
func (r *Reader) Read(p []byte) (n int, err error) {

	if !r.from_reader {
		buffer_len := len(r.decoder_bytes) - r.decoder_copied

		if len(p) <= buffer_len {
			for i := 0; i < len(p); i++ {
				p[i] = r.decoder_bytes[r.decoder_copied]
				r.decoder_copied++
			}
		} else {
			for i := 0; i < buffer_len; i++ {
				p[i] = r.decoder_bytes[r.decoder_copied]
				r.decoder_copied++
			}
		}

		transferred := len(r.decoder_bytes) - r.decoder_copied

		if transferred == 0 {
			err = io.EOF
		}

		return buffer_len - transferred, err
	} else {
		return r.Read(p)
	}

}
