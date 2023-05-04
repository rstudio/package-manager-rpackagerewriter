// Copyright (C) 2023 by Posit Software, PBC
package utils

import (
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestReaderSuite(t *testing.T) {
	suite.Run(t, &ReaderSuite{})
}

type ReaderSuite struct {
	suite.Suite
}

func (s *ReaderSuite) TestEOFTeeReader() {
	r := strings.NewReader("whatever")
	pipeR, pipeW := io.Pipe()
	tee := NewEOFTeeReader(r, pipeW)

	done := make(chan struct{})

	checkStream := func(r io.Reader) {
		b, err := ioutil.ReadAll(r)
		s.Require().Nil(err) // ReadAll doesn't return an err if == EOF
		s.Require().Equal([]byte("whatever"), b)
		done <- struct{}{}
	}

	// Read the two streams in parallel since they're not buffered.
	go checkStream(tee)
	go checkStream(pipeR)

	<-done
	<-done

	// The fact that both streams terminated their ReadAll() call means that
	// they errored or EOF'd, and since we check against error, this means that
	// we've successfully got two client streams to get an EOF out of a tee.
}
