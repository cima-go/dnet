package rpc

import (
	"bytes"
	"compress/gzip"
	"io"
)

func Compress(input []byte) (parsed []byte, err error) {
	b := new(bytes.Buffer)
	w := gzip.NewWriter(b)

	_, _ = io.Copy(w, bytes.NewReader(input))

	_ = w.Flush()
	_ = w.Close()

	parsed = b.Bytes()
	return
}

func Decompress(input []byte) (parsed []byte, err error) {
	b := new(bytes.Buffer)
	r, err := gzip.NewReader(bytes.NewReader(input))
	if err != nil {
		return
	}

	_, _ = io.Copy(b, r)

	parsed = b.Bytes()
	return
}
