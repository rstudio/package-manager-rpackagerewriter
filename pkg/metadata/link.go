// Copyright (C) 2023 by Posit Software, PBC
package metadata

import (
	"log"
	"regexp"
	"strings"
)

type LinkOperator int16
type LinkType int16

const (
	VersionEquals LinkOperator = 0
	VersionGT     LinkOperator = 1
	VersionGTE    LinkOperator = 2
	VersionLT     LinkOperator = 3
	VersionLTE    LinkOperator = 4
)

const (
	LinkImports   LinkType = 0
	LinkDepends   LinkType = 1
	LinkSuggests  LinkType = 2
	LinkLinkingTo LinkType = 3
	LinkEnhances  LinkType = 4
)

// Link represents the link from one package to another.
type Link struct {
	Name     string       `json:"name"`
	Raw      string       `json:"-"`
	Operator LinkOperator `json:"operator"`
	Version  string       `json:"version"`
	Type     LinkType     `json:"type"`
}

func (a Link) Equals(b Link) bool {
	return a.Name == b.Name &&
		a.Operator == b.Operator &&
		a.Type == b.Type &&
		a.Version == b.Version
}

// hasVersion matches a link version in the format: "(<operator> <version>)".
var hasVersion = regexp.MustCompile(`^([^(]+)\((.+)\)$`)

// operatorMatch locates the comparison operator.
var operatorMatch = regexp.MustCompile(`(==|<=?|>=?)?\s*(.+)`)

// ParseLinks generates a comma-separated list of links in the format:
// "<package name> (<operator> <version>)
func ParseLinks(raw string, linkType LinkType) []Link {
	links := []Link{}
	list := strings.Split(raw, ",")

	for _, raw := range list {
		raw := strings.TrimSpace(raw)
		matches := hasVersion.FindStringSubmatch(raw)
		if matches != nil {
			// Version specified.
			link := Link{
				Name: strings.TrimSpace(matches[1]),
				Raw:  strings.TrimSpace(raw),
				Type: linkType,
			}
			if len(matches) > 1 {
				// Separate the operator from the version number
				components := operatorMatch.FindStringSubmatch(matches[2])
				link.Operator = resolveOperator(components[1:])
				link.Version = resolveVersion(components[1:])
			}
			links = append(links, link)
		} else if raw != "" {
			// No version specified.
			links = append(links, Link{
				Name: strings.TrimSpace(raw),
				Raw:  strings.TrimSpace(raw),
				Type: linkType,
			})
		}
	}

	return links
}

// resolveOperator maps an operator symbol to the internal value.
func resolveOperator(components []string) LinkOperator {
	if len(components) == 1 {
		return VersionEquals
	}

	switch components[0] {
	case "":
		return VersionEquals
	case "==":
		return VersionEquals
	case ">":
		return VersionGT
	case ">=":
		return VersionGTE
	case "<":
		return VersionLT
	case "<=":
		return VersionLTE
	}

	// We shouldn't ever get here.
	log.Printf("Error: could not locate operator for %v\n", components)
	return VersionEquals
}

// resolveVersion separates the version from a potential operator string.
func resolveVersion(components []string) string {
	if len(components) == 1 {
		return components[0]
	}

	return components[1]
}
