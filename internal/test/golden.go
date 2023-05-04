// Copyright (C) 2023 by Posit Software, PBC
package test

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/stretchr/testify/suite"
)

var (
	UpdateGolden = flag.Bool("update", false, "update .golden files")
)

func LoadOrUpdateGolden(actual string, path string, file string) (string, error) {
	fileWithPath := filepath.Join(path, file)
	if *UpdateGolden {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("error creating directory: %s", err)
		}
		err = os.WriteFile(fileWithPath, []byte(actual), 0644)
		if err != nil {
			return "", fmt.Errorf("error writing file: %s", err)
		}
	}
	expected, err := os.ReadFile(fileWithPath)
	if err != nil {
		return "", fmt.Errorf("error reading file: %s", err)
	}
	return string(expected), nil
}

// TestifyGolden validates incoming output against .golden files using Testify
func TestifyGolden(actual string, s *suite.Suite) {
	expected, err := LoadOrUpdateGolden(actual, "testdata", strings.ReplaceAll(s.T().Name(), "/", ".")+".golden")
	s.Assert().Nil(err)
	s.Assert().Equal(expected, actual)
}
