// Copyright (C) 2023 by Posit Software, PBC
package utils

import (
	"crypto/sha256"
	"io"
)

// NewEOFTeeReader - io.TeeReader is almost what we want, but it doesn't tee EOFs, which is
// important for us when calculating checksums. This is a copy of the
// io.TeeReader source with the modification that EOFs are also propagated.
func NewEOFTeeReader(r io.Reader, w *io.PipeWriter) io.Reader {
	return &EofTeeReader{r, w}
}

type EofTeeReader struct {
	r io.Reader
	w *io.PipeWriter
}

func (t *EofTeeReader) Read(p []byte) (n int, err error) {
	n, err = t.r.Read(p)
	if n > 0 {
		if n, err := t.w.Write(p[:n]); err != nil {
			_ = t.w.CloseWithError(err)
			return n, err
		}
	} else if err == io.EOF {
		_ = t.w.CloseWithError(io.EOF)
	}
	return
}

// ComputeSha256Stream computes the sha256 of a given stream
func ComputeSha256Stream(stream io.Reader) (int64, []byte, error) {
	hash := sha256.New()
	var sz int64
	var err error
	if sz, err = io.Copy(hash, stream); err != nil {
		return 0, nil, err
	}

	return sz, hash.Sum(nil), nil
}
