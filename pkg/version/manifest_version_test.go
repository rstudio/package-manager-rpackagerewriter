// Copyright (C) 2023 by Posit Software, PBC
package version

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestManifestVersionSuite(t *testing.T) {
	suite.Run(t, &ManifestVersionSuite{})
}

type ManifestVersionSuite struct {
	suite.Suite
}

type ManifestVersionToCheck struct {
	versionString   string
	expectedVersion ManifestVersion
}

// Most historical `rsconnect` versions at the time these tests were created
// (in order).
//
// CRANifestVersion `0.3.6` historically appears between `0.3.59` and `0.3.60`, but we
// feel this is a version ordering bug, so it's been moved into the expected
// position. It's very unlikely that this anomaly will ever cause a problem
// in practice.
var allManifestVersions = []ManifestVersionToCheck{
	{"v2", ManifestVersion{"v2", 2, 0, 0, []int{2}, true}},
	{"v1/1", ManifestVersion{"v1/1", 1, 1, 0, []int{1, 1}, true}},
	{"v1/2", ManifestVersion{"v1/2", 1, 2, 0, []int{1, 2}, true}},
	{"v2/3.1", ManifestVersion{"v2/3.1", 2, 3, 1, []int{2, 3, 1}, true}},
	{"v4/3/2", ManifestVersion{"v4/3/2", 4, 3, 2, []int{4, 3, 2}, true}},
}

func (s *ManifestVersionSuite) TestNewVersion() {
	v, err := ParseNewManifestVersion("v2/3/4")
	s.Require().Nil(err)
	s.Require().Equal(ManifestVersion{"v2/3/4", 2, 3, 4, []int{2, 3, 4}, true}, v)

	v, err = ParseNewManifestVersion("")
	s.Require().Nil(err)
	s.Require().Equal(ManifestVersion{"", 0, 0, 0, nil, false}, v)

	v = NewManifestVersion(2, 3, 6)
	s.Require().Equal(ManifestVersion{"v2/3/6", 2, 3, 6, []int{2, 3, 6}, true}, v)
}

func (s *ManifestVersionSuite) TestVersionParts() {
	v, err := ParseNewManifestVersion("1.2.3")
	s.Require().Nil(err)
	s.Require().Equal(1, v.Schema)
	s.Require().Equal(2, v.Major)
	s.Require().Equal(3, v.Minor)

	v, err = ParseNewManifestVersion("4.1")
	s.Require().Nil(err)
	s.Require().Equal(4, v.Schema)
	s.Require().Equal(1, v.Major)
	s.Require().Equal(0, v.Minor)

	v, err = ParseNewManifestVersion("")
	s.Require().Nil(err)
	s.Require().Equal(0, v.Schema)
	s.Require().Equal(0, v.Major)
	s.Require().Equal(0, v.Minor)
}

func (s *ManifestVersionSuite) TestCompareVersionsNotSet() {
	// second not set
	s.Require().Equal(1, CompareManifestVersions(ManifestVersion{"set", 1, 2, 3, []int{1, 2, 3}, true}, ManifestVersion{"set", 2, 2, 3, []int{2, 2, 3}, false}))
	// first not set
	s.Require().Equal(-1, CompareManifestVersions(ManifestVersion{"set", 1, 2, 3, []int{1, 2, 3}, false}, ManifestVersion{"set", 2, 2, 3, []int{2, 2, 3}, true}))
	// both not set
	s.Require().Equal(0, CompareManifestVersions(ManifestVersion{"set", 1, 2, 3, []int{1, 2, 3}, false}, ManifestVersion{"set", 2, 2, 3, []int{2, 2, 3}, false}))
}

func (s *ManifestVersionSuite) TestCompareManifestVersions() {
	// major version differs
	s.Require().Equal(-1, CompareManifestVersions(ManifestVersion{"set", 1, 2, 3, []int{1, 2, 3}, true}, ManifestVersion{"set", 2, 2, 3, []int{2, 2, 3}, true}))
	s.Require().Equal(1, CompareManifestVersions(ManifestVersion{"set", 2, 2, 3, []int{2, 2, 3}, true}, ManifestVersion{"set", 1, 2, 3, []int{1, 2, 3}, true}))

	// minor version differs
	s.Require().Equal(-1, CompareManifestVersions(ManifestVersion{"set", 1, 2, 3, []int{1, 2, 3}, true}, ManifestVersion{"set", 1, 3, 3, []int{1, 3, 3}, true}))
	s.Require().Equal(1, CompareManifestVersions(ManifestVersion{"set", 1, 3, 3, []int{1, 3, 3}, true}, ManifestVersion{"set", 1, 2, 3, []int{1, 2, 3}, true}))

	// patch version differs
	s.Require().Equal(-1, CompareManifestVersions(ManifestVersion{"set", 1, 2, 3, []int{1, 2, 3}, true}, ManifestVersion{"set", 1, 2, 4, []int{1, 2, 4}, true}))
	s.Require().Equal(1, CompareManifestVersions(ManifestVersion{"set", 1, 2, 4, []int{1, 2, 4}, true}, ManifestVersion{"set", 1, 2, 3, []int{1, 2, 3}, true}))

	// equal
	s.Require().Equal(0, CompareManifestVersions(ManifestVersion{"set", 1, 2, 3, []int{1, 2, 3}, true}, ManifestVersion{"set", 1, 2, 3, []int{1, 2, 3}, true}))
}

func (s *ManifestVersionSuite) TestCompareNumerically() {
	// major version compares numerically
	s.Require().Equal(-1, CompareManifestVersions(ManifestVersion{"set", 6, 2, 3, []int{6, 2, 3}, true}, ManifestVersion{"set", 56, 2, 3, []int{56, 2, 3}, true}))
	s.Require().Equal(1, CompareManifestVersions(ManifestVersion{"set", 56, 2, 3, []int{56, 2, 3}, true}, ManifestVersion{"set", 6, 2, 3, []int{6, 2, 3}, true}))

	// minor version compares numerically
	s.Require().Equal(-1, CompareManifestVersions(ManifestVersion{"set", 1, 6, 3, []int{1, 6, 3}, true}, ManifestVersion{"set", 1, 56, 3, []int{1, 56, 3}, true}))
	s.Require().Equal(1, CompareManifestVersions(ManifestVersion{"set", 1, 56, 3, []int{1, 56, 3}, true}, ManifestVersion{"set", 1, 6, 3, []int{1, 6, 3}, true}))

	// patch version compares numerically
	s.Require().Equal(-1, CompareManifestVersions(ManifestVersion{"set", 6, 2, 6, []int{6, 2, 6}, true}, ManifestVersion{"set", 6, 2, 56, []int{6, 2, 56}, true}))
	s.Require().Equal(1, CompareManifestVersions(ManifestVersion{"set", 6, 2, 56, []int{6, 2, 56}, true}, ManifestVersion{"set", 6, 2, 6, []int{6, 2, 6}, true}))
}

func (s *ManifestVersionSuite) TestCompareManifestVersionsSchemaMajor() {
	// schema version differs
	s.Require().Equal(-1, CompareSchemaMajor(ManifestVersion{"set", 1, 2, 3, []int{1, 2, 3}, true}, ManifestVersion{"set", 2, 2, 3, []int{2, 2, 3}, true}))
	s.Require().Equal(1, CompareSchemaMajor(ManifestVersion{"set", 2, 2, 3, []int{2, 2, 3}, true}, ManifestVersion{"set", 1, 2, 3, []int{1, 2, 3}, true}))

	// major version differs
	s.Require().Equal(-1, CompareSchemaMajor(ManifestVersion{"set", 1, 2, 3, []int{1, 2, 3}, true}, ManifestVersion{"set", 1, 3, 3, []int{1, 3, 3}, true}))
	s.Require().Equal(1, CompareSchemaMajor(ManifestVersion{"set", 1, 3, 3, []int{1, 3, 3}, true}, ManifestVersion{"set", 1, 2, 3, []int{1, 2, 3}, true}))

	// only minor version differs; should equal anyway
	s.Require().Equal(0, CompareSchemaMajor(ManifestVersion{"set", 1, 2, 3, []int{1, 2, 3}, true}, ManifestVersion{"set", 1, 2, 4, []int{1, 2, 4}, true}))
	s.Require().Equal(0, CompareSchemaMajor(ManifestVersion{"set", 1, 2, 4, []int{1, 2, 4}, true}, ManifestVersion{"set", 1, 2, 3, []int{1, 2, 3}, true}))

	// equal
	s.Require().Equal(0, CompareSchemaMajor(ManifestVersion{"set", 1, 2, 3, []int{1, 2, 3}, true}, ManifestVersion{"set", 1, 2, 3, []int{1, 2, 3}, true}))

	// LessThan test
	v1, _ := ParseNewManifestVersion("v3/4/5")
	v2, _ := ParseNewManifestVersion("v3/4/6")
	s.Require().Equal(true, v1.LessThan(v2))

	// LessThanOrEqual test
	v1, _ = ParseNewManifestVersion("v3/4/5")
	v2, _ = ParseNewManifestVersion("v3/4/6")
	s.Require().Equal(true, v1.LessThanOrEqual(v2))
	v1, _ = ParseNewManifestVersion("v3/4/5")
	v2, _ = ParseNewManifestVersion("v3/4/5")
	s.Require().Equal(true, v1.LessThanOrEqual(v2))
	v1, _ = ParseNewManifestVersion("v3/4/6")
	v2, _ = ParseNewManifestVersion("v3/4/5")
	s.Require().Equal(false, v1.LessThanOrEqual(v2))

	// LessThanMajor test
	v1, _ = ParseNewManifestVersion("v3/4/5")
	v2, _ = ParseNewManifestVersion("v3/5/6")
	s.Require().Equal(true, v1.LessThanMajor(v2))

	v1, _ = ParseNewManifestVersion("v3/4/5")
	v2, _ = ParseNewManifestVersion("v3/4/7")
	s.Require().Equal(false, v1.LessThanMajor(v2))

	// EqualsMajor test
	v1, _ = ParseNewManifestVersion("v3/4/5")
	v2, _ = ParseNewManifestVersion("v3/4/6")
	s.Require().Equal(true, v1.EqualsMajor(v2))
	v1, _ = ParseNewManifestVersion("v3/4/5")
	v2, _ = ParseNewManifestVersion("v3/4/5")
	s.Require().Equal(true, v1.EqualsMajor(v2))
	v1, _ = ParseNewManifestVersion("v3/4/6")
	v2, _ = ParseNewManifestVersion("v3/5/6")
	s.Require().Equal(false, v1.EqualsMajor(v2))
	v1, _ = ParseNewManifestVersion("v4/4/6")
	v2, _ = ParseNewManifestVersion("v3/4/6")
	s.Require().Equal(false, v1.EqualsMajor(v2))
}

func (s *ManifestVersionSuite) TestParseVersion() {
	for _, each := range allManifestVersions {
		v, err := ParseNewManifestVersion(each.versionString)
		s.Require().Nilf(err, "%s", each.versionString)
		s.Require().Equalf(each.expectedVersion, v, "%s", each.versionString)
	}
}

func (s *ManifestVersionSuite) TestParseVersionErr() {
	for _, each := range []string{
		// numeric but Atoi errors with overflow.
		"9999999999999999999.9999999999999999999.9999999999999999999.9999999999999999999",
		" ",
		"2.3.2.2 3",
	} {
		_, err := ParseNewManifestVersion(each)
		s.Require().NotNilf(err, "%s", each)
	}
}
