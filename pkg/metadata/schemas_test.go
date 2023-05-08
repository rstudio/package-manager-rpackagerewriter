// Copyright (C) 2023 by Posit Software, PBC
package metadata

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/rstudio/package-manager-rpackagerewriter/pkg/version"
)

func TestSchemaSuite(t *testing.T) {
	suite.Run(t, &SchemaSuite{})
}

type SchemaSuite struct {
	suite.Suite
}

type testCase struct {
	version  string
	expected bool
}

func (s *SchemaSuite) TestIsValidSchema() {

	for _, test := range []testCase{
		{"v1", true},
		{"v2", true},
		{"v3", true},
		{"v4", false},
		{"1", true},
		{"2", true},
		{"3", true},
		{"4", false},
		{"v1/2", true},
		{"v2/2", true},
		{"v3/2", true},
		{"v4/2", false},
		{"v1/2/0", true},
		{"v2/2/4", true},
		{"v3/2/4", true},
		{"v4/2/4", false},
	} {
		v, err := version.ParseNewManifestVersion(test.version)
		s.Require().Nil(err)
		s.Require().Equal(test.expected, IsValidSchema(v))
	}
}

func (s *SchemaSuite) TestIsValidBiocSchema() {

	for _, test := range []testCase{
		{"v1", false},
		{"v2", false},
		{"v3", true},
		{"v4", true},
		{"v13", false},
		{"1", false},
		{"2", false},
		{"3", true},
		{"4", true},
		{"13", false},
		{"v1/2", false},
		{"v2/2", false},
		{"v3/2", true},
		{"v4/2", true},
		{"v13/2", false},
		{"v1/2/0", false},
		{"v2/2/4", false},
		{"v3/2/4", true},
		{"v4/2/4", true},
		{"v13/2/4", false},
	} {
		v, err := version.ParseNewManifestVersion(test.version)
		s.Require().Nil(err)
		s.Require().Equal(test.expected, IsValidBiocSchema(v))
	}
}
