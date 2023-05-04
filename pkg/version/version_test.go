// Copyright (C) 2023 by Posit Software, PBC
package version

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestVersionSuite(t *testing.T) {
	suite.Run(t, &VersionSuite{})
}

type VersionSuite struct {
	suite.Suite
}

type VersionToCheck struct {
	versionString   string
	expectedVersion Version
}

// Most historical `rsconnect` versions at the time these tests were created
// (in order).
//
// Version `0.3.6` historically appears between `0.3.59` and `0.3.60`, but we
// feel this is a version ordering bug, so it's been moved into the expected
// position. It's very unlikely that this anomaly will ever cause a problem
// in practice.
var allVersions = []VersionToCheck{
	{"0.1", Version{"0.1", 0, 1, 0, 0, []int{0, 1}, true, nil}},
	{"0.2", Version{"0.2", 0, 2, 0, 0, []int{0, 2}, true, nil}},
	{"0.2.1", Version{"0.2.1", 0, 2, 1, 0, []int{0, 2, 1}, true, nil}},
	{"0.3", Version{"0.3", 0, 3, 0, 0, []int{0, 3}, true, nil}},
	{"0.3.1", Version{"0.3.1", 0, 3, 1, 0, []int{0, 3, 1}, true, nil}},
	{"0.3.2", Version{"0.3.2", 0, 3, 2, 0, []int{0, 3, 2}, true, nil}},
	{"0.3.3", Version{"0.3.3", 0, 3, 3, 0, []int{0, 3, 3}, true, nil}},
	{"0.3.4", Version{"0.3.4", 0, 3, 4, 0, []int{0, 3, 4}, true, nil}},
	{"0.3.5", Version{"0.3.5", 0, 3, 5, 0, []int{0, 3, 5}, true, nil}},
	{"0.3.6", Version{"0.3.6", 0, 3, 6, 0, []int{0, 3, 6}, true, nil}},
	{"0.3.51", Version{"0.3.51", 0, 3, 51, 0, []int{0, 3, 51}, true, nil}},
	{"0.3.52", Version{"0.3.52", 0, 3, 52, 0, []int{0, 3, 52}, true, nil}},
	{"0.3.53", Version{"0.3.53", 0, 3, 53, 0, []int{0, 3, 53}, true, nil}},
	{"0.3.54", Version{"0.3.54", 0, 3, 54, 0, []int{0, 3, 54}, true, nil}},
	{"0.3.55", Version{"0.3.55", 0, 3, 55, 0, []int{0, 3, 55}, true, nil}},
	{"0.3.56", Version{"0.3.56", 0, 3, 56, 0, []int{0, 3, 56}, true, nil}},
	{"0.3.57", Version{"0.3.57", 0, 3, 57, 0, []int{0, 3, 57}, true, nil}},
	{"0.3.58", Version{"0.3.58", 0, 3, 58, 0, []int{0, 3, 58}, true, nil}},
	{"0.3.59", Version{"0.3.59", 0, 3, 59, 0, []int{0, 3, 59}, true, nil}},
	{"0.3.60", Version{"0.3.60", 0, 3, 60, 0, []int{0, 3, 60}, true, nil}},
	{"0.3.70", Version{"0.3.70", 0, 3, 70, 0, []int{0, 3, 70}, true, nil}},
	{"0.3.71", Version{"0.3.71", 0, 3, 71, 0, []int{0, 3, 71}, true, nil}},
	{"0.3.75", Version{"0.3.75", 0, 3, 75, 0, []int{0, 3, 75}, true, nil}},
	{"0.3.80", Version{"0.3.80", 0, 3, 80, 0, []int{0, 3, 80}, true, nil}},
	{"0.3.81", Version{"0.3.81", 0, 3, 81, 0, []int{0, 3, 81}, true, nil}},
	{"0.4.0", Version{"0.4.0", 0, 4, 0, 0, []int{0, 4, 0}, true, nil}},
	{"0.4.1", Version{"0.4.1", 0, 4, 1, 0, []int{0, 4, 1}, true, nil}},
	{"0.4.1.1", Version{"0.4.1.1", 0, 4, 1, 1, []int{0, 4, 1, 1}, true, nil}},
	{"0.4.1.2", Version{"0.4.1.2", 0, 4, 1, 2, []int{0, 4, 1, 2}, true, nil}},
	{"0.4.1.3", Version{"0.4.1.3", 0, 4, 1, 3, []int{0, 4, 1, 3}, true, nil}},
	{"0.4.1.4", Version{"0.4.1.4", 0, 4, 1, 4, []int{0, 4, 1, 4}, true, nil}},
	{"0.4.1.5", Version{"0.4.1.5", 0, 4, 1, 5, []int{0, 4, 1, 5}, true, nil}},
	{"0.4.1.6", Version{"0.4.1.6", 0, 4, 1, 6, []int{0, 4, 1, 6}, true, nil}},
	{"0.4.1.7", Version{"0.4.1.7", 0, 4, 1, 7, []int{0, 4, 1, 7}, true, nil}},
	{"0.4.1.8", Version{"0.4.1.8", 0, 4, 1, 8, []int{0, 4, 1, 8}, true, nil}},
	{"0.4.1.9", Version{"0.4.1.9", 0, 4, 1, 9, []int{0, 4, 1, 9}, true, nil}},
	{"0.4.1.10", Version{"0.4.1.10", 0, 4, 1, 10, []int{0, 4, 1, 10}, true, nil}},
	{"0.4.1.11", Version{"0.4.1.11", 0, 4, 1, 11, []int{0, 4, 1, 11}, true, nil}},
	{"0.4.2", Version{"0.4.2", 0, 4, 2, 0, []int{0, 4, 2}, true, nil}},
	{"0.4.2.1", Version{"0.4.2.1", 0, 4, 2, 1, []int{0, 4, 2, 1}, true, nil}},
	{"0.4.2.2", Version{"0.4.2.2", 0, 4, 2, 2, []int{0, 4, 2, 2}, true, nil}},
	{"0.4.3", Version{"0.4.3", 0, 4, 3, 0, []int{0, 4, 3}, true, nil}},
	{"0.4.4", Version{"0.4.4", 0, 4, 4, 0, []int{0, 4, 4}, true, nil}},
	{"0.4.5", Version{"0.4.5", 0, 4, 5, 0, []int{0, 4, 5}, true, nil}},
	{"0.4.6", Version{"0.4.6", 0, 4, 6, 0, []int{0, 4, 6}, true, nil}},
	{"0.5.0-2", Version{"0.5.0-2", 0, 5, 0, 2, []int{0, 5, 0, 2}, true, nil}},
	{"0.6", Version{"0.6", 0, 6, 0, 0, []int{0, 6}, true, nil}},
	{"0.6.9000", Version{"0.6.9000", 0, 6, 9000, 0, []int{0, 6, 9000}, true, nil}},
	{"0.7", Version{"0.7", 0, 7, 0, 0, []int{0, 7}, true, nil}},
	{"0.7.0-1", Version{"0.7.0-1", 0, 7, 0, 1, []int{0, 7, 0, 1}, true, nil}},
	{"0.7.0-2", Version{"0.7.0-2", 0, 7, 0, 2, []int{0, 7, 0, 2}, true, nil}},
	{"0.8", Version{"0.8", 0, 8, 0, 0, []int{0, 8}, true, nil}},
	{"0.8.1", Version{"0.8.1", 0, 8, 1, 0, []int{0, 8, 1}, true, nil}},
	{"0.8.2", Version{"0.8.2", 0, 8, 2, 0, []int{0, 8, 2}, true, nil}},
	{"0.9.a.2", Version{"0.9.a.2", 0, 9, 0, 2, []int{0, 9, 0, 2}, true, nil}},
	{"0.9.a23.2", Version{"0.9.a23.2", 0, 9, 23, 2, []int{0, 9, 23, 2}, true, nil}},
	{"a0b1.9b2c3.a23.2b1", Version{"a0b1.9b2c3.a23.2b1", 1, 923, 23, 21, []int{1, 923, 23, 21}, true, nil}},
}

func (s *VersionSuite) TestNewVersion() {
	v, err := ParseNewVersion("1.2.3")
	s.Require().Nil(err)
	s.Require().Equal(Version{"1.2.3", 1, 2, 3, 0, []int{1, 2, 3}, true, nil}, v)

	v, err = ParseNewVersion("")
	s.Require().Nil(err)
	s.Require().Equal(Version{"", 0, 0, 0, 0, nil, false, nil}, v)
}

func (s *VersionSuite) TestVersionParts() {
	v, err := ParseNewVersion("1.2.3.4")
	s.Require().Nil(err)
	s.Require().Equal(1, v.Major)
	s.Require().Equal(2, v.Minor)
	s.Require().Equal(3, v.Patch)

	v, err = ParseNewVersion("4.1")
	s.Require().Nil(err)
	s.Require().Equal(4, v.Major)
	s.Require().Equal(1, v.Minor)
	s.Require().Equal(0, v.Patch)

	v, err = ParseNewVersion("")
	s.Require().Nil(err)
	s.Require().Equal(0, v.Major)
	s.Require().Equal(0, v.Minor)
	s.Require().Equal(0, v.Patch)
}

func (s *VersionSuite) TestCompareVersionsNotSet() {
	// second not set
	s.Require().Equal(1, CompareVersions(Version{"set", 1, 2, 3, 0, []int{1, 2, 3}, true, nil}, Version{"set", 2, 2, 3, 0, []int{2, 2, 3}, false, nil}))
	// first not set
	s.Require().Equal(-1, CompareVersions(Version{"set", 1, 2, 3, 0, []int{1, 2, 3}, false, nil}, Version{"set", 2, 2, 3, 0, []int{2, 2, 3}, true, nil}))
	// both not set
	s.Require().Equal(0, CompareVersions(Version{"set", 1, 2, 3, 0, []int{1, 2, 3}, false, nil}, Version{"set", 2, 2, 3, 0, []int{2, 2, 3}, false, nil}))
}

func (s *VersionSuite) TestCompareVersions() {
	// major version differs
	s.Require().Equal(-1, CompareVersions(Version{"set", 1, 2, 3, 0, []int{1, 2, 3}, true, nil}, Version{"set", 2, 2, 3, 0, []int{2, 2, 3}, true, nil}))
	s.Require().Equal(1, CompareVersions(Version{"set", 2, 2, 3, 0, []int{2, 2, 3}, true, nil}, Version{"set", 1, 2, 3, 0, []int{1, 2, 3}, true, nil}))

	// minor version differs
	s.Require().Equal(-1, CompareVersions(Version{"set", 1, 2, 3, 0, []int{1, 2, 3}, true, nil}, Version{"set", 1, 3, 3, 0, []int{1, 3, 3}, true, nil}))
	s.Require().Equal(1, CompareVersions(Version{"set", 1, 3, 3, 0, []int{1, 3, 3}, true, nil}, Version{"set", 1, 2, 3, 0, []int{1, 2, 3}, true, nil}))

	// patch version differs
	s.Require().Equal(-1, CompareVersions(Version{"set", 1, 2, 3, 0, []int{1, 2, 3}, true, nil}, Version{"set", 1, 2, 4, 0, []int{1, 2, 4}, true, nil}))
	s.Require().Equal(1, CompareVersions(Version{"set", 1, 2, 4, 0, []int{1, 2, 4}, true, nil}, Version{"set", 1, 2, 3, 0, []int{1, 2, 3}, true, nil}))

	// rev version differs
	s.Require().Equal(-1, CompareVersions(Version{"set", 1, 2, 3, 4, []int{1, 2, 3, 4}, true, nil}, Version{"set", 1, 2, 3, 5, []int{1, 2, 3, 5}, true, nil}))
	s.Require().Equal(1, CompareVersions(Version{"set", 1, 2, 3, 5, []int{1, 2, 3, 5}, true, nil}, Version{"set", 1, 2, 3, 4, []int{1, 2, 3, 4}, true, nil}))

	// equal
	s.Require().Equal(0, CompareVersions(Version{"set", 1, 2, 3, 4, []int{1, 2, 3, 4}, true, nil}, Version{"set", 1, 2, 3, 4, []int{1, 2, 3, 4}, true, nil}))
}

func (s *VersionSuite) TestCompareNumerically() {
	// major version compares numerically
	s.Require().Equal(-1, CompareVersions(Version{"set", 6, 2, 3, 0, []int{6, 2, 3}, true, nil}, Version{"set", 56, 2, 3, 0, []int{56, 2, 3}, true, nil}))
	s.Require().Equal(1, CompareVersions(Version{"set", 56, 2, 3, 0, []int{56, 2, 3}, true, nil}, Version{"set", 6, 2, 3, 0, []int{6, 2, 3}, true, nil}))

	// minor version compares numerically
	s.Require().Equal(-1, CompareVersions(Version{"set", 1, 6, 3, 0, []int{1, 6, 3}, true, nil}, Version{"set", 1, 56, 3, 0, []int{1, 56, 3}, true, nil}))
	s.Require().Equal(1, CompareVersions(Version{"set", 1, 56, 3, 0, []int{1, 56, 3}, true, nil}, Version{"set", 1, 6, 3, 0, []int{1, 6, 3}, true, nil}))

	// patch version compares numerically
	s.Require().Equal(-1, CompareVersions(Version{"set", 6, 2, 6, 0, []int{6, 2, 6}, true, nil}, Version{"set", 6, 2, 56, 0, []int{6, 2, 56}, true, nil}))
	s.Require().Equal(1, CompareVersions(Version{"set", 6, 2, 56, 0, []int{6, 2, 56}, true, nil}, Version{"set", 6, 2, 6, 0, []int{6, 2, 6}, true, nil}))

	// rev version compares numerically
	s.Require().Equal(-1, CompareVersions(Version{"set", 1, 2, 3, 6, []int{1, 2, 3, 6}, true, nil}, Version{"set", 1, 2, 3, 56, []int{1, 2, 3, 56}, true, nil}))
	s.Require().Equal(1, CompareVersions(Version{"set", 1, 2, 3, 56, []int{1, 2, 3, 56}, true, nil}, Version{"set", 1, 2, 3, 6, []int{1, 2, 3, 6}, true, nil}))
}

func (s *VersionSuite) TestParseVersion() {
	for _, each := range allVersions {
		v, err := ParseNewVersion(each.versionString)
		s.Require().Nilf(err, "%s", each.versionString)
		s.Require().Equalf(each.expectedVersion, v, "%s", each.versionString)
	}
}

func (s *VersionSuite) TestParseVersionErr() {
	for _, each := range []string{
		// numeric but Atoi errors with overflow.
		"9999999999999999999.9999999999999999999.9999999999999999999.9999999999999999999",
		" ",
		"2.3.2.2 3",
	} {
		_, err := ParseNewVersion(each)
		s.Require().NotNilf(err, "%s", each)
	}
}
