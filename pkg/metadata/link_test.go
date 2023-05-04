// Copyright (C) 2023 by Posit Software, PBC
package metadata

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestLinkSuite(t *testing.T) {
	suite.Run(t, &LinkSuite{})
}

type LinkSuite struct {
	suite.Suite
}

func (s *LinkSuite) TestResolveOperator() {
	var operator LinkOperator
	operator = resolveOperator([]string{"0.1.2"})
	s.Require().Equal(VersionEquals, operator)

	operator = resolveOperator([]string{"", "0.1.2"})
	s.Require().Equal(VersionEquals, operator)

	operator = resolveOperator([]string{"==", "0.1.2"})
	s.Require().Equal(VersionEquals, operator)

	operator = resolveOperator([]string{">", "0.1.2"})
	s.Require().Equal(VersionGT, operator)

	operator = resolveOperator([]string{">=", "0.1.2"})
	s.Require().Equal(VersionGTE, operator)

	operator = resolveOperator([]string{"<", "0.1.2"})
	s.Require().Equal(VersionLT, operator)

	operator = resolveOperator([]string{"<=", "0.1.2"})
	s.Require().Equal(VersionLTE, operator)
}

func (s *LinkSuite) TestResolveVersion() {
	version := resolveVersion([]string{"<=", "0.1.2"})
	s.Require().Equal("0.1.2", version)

	version = resolveVersion([]string{"0.1.2"})
	s.Require().Equal("0.1.2", version)
}

func (s *LinkSuite) TestParseLinks() {
	type rawCase struct {
		raw      string
		expected []Link
	}
	rawCases := []rawCase{
		{
			raw: "R (>= 2.0.1)",
			expected: []Link{
				{
					Name:     "R",
					Raw:      "R (>= 2.0.1)",
					Operator: VersionGTE,
					Version:  "2.0.1",
					Type:     LinkDepends,
				},
			},
		},
		{
			raw: "R",
			expected: []Link{
				{
					Name:     "R",
					Raw:      "R",
					Operator: VersionEquals,
					Version:  "",
					Type:     LinkDepends,
				},
			},
		},
		{
			raw: " R (>= 2.0.1) ",
			expected: []Link{
				{
					Name:     "R",
					Raw:      "R (>= 2.0.1)",
					Operator: VersionGTE,
					Version:  "2.0.1",
					Type:     LinkDepends,
				},
			},
		},
		{
			raw: " R (>=2.0.1) ",
			expected: []Link{
				{
					Name:     "R",
					Raw:      "R (>=2.0.1)",
					Operator: VersionGTE,
					Version:  "2.0.1",
					Type:     LinkDepends,
				},
			},
		},
	}
	for _, cs := range rawCases {
		links := ParseLinks(cs.raw, LinkDepends)
		s.Require().Equalf(links, cs.expected, "Failed parsing raw link %s", cs.raw)

	}
}
