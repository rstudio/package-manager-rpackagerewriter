// Copyright (C) 2023 by Posit Software, PBC
package utils

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestUtilSuite(t *testing.T) {
	suite.Run(t, &UtilSuite{})
}

type UtilSuite struct {
	suite.Suite
}

func (s *UtilSuite) TestFileExists() {
	osStat = func(path string) (os.FileInfo, error) {
		return nil, nil
	}
	exists, err := FileExists("path")
	s.Require().Equal(true, exists)
	s.Require().Nil(err)

	osStat = func(path string) (os.FileInfo, error) {
		return nil, os.ErrNotExist
	}
	exists, err = FileExists("path")
	s.Require().Equal(false, exists)
	s.Require().Nil(err)

	osStat = func(path string) (os.FileInfo, error) {
		return nil, errors.New("another error")
	}
	exists, err = FileExists("path")
	s.Require().Equal(false, exists)
	s.Require().NotNil(err)
}
