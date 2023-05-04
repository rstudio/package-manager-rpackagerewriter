// Copyright (C) 2023 by Posit Software, PBC
package version

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Version represents semantic version information with each component as an int.
// Other types of version strings are not supported.
type Version struct {
	Raw   string
	Major int
	Minor int
	Patch int
	Rev   int
	Parts []int
	Set   bool
	Err   error
}

// String generates a string representation of the version.
func (v Version) String() string {
	return v.Raw
}

// Equals tests that this version is equivalent to another.
func (v Version) Equals(other Version) bool {
	return v.Raw == other.Raw &&
		v.Set == other.Set
}

// GreaterThan tests that this version is greater than another.
func (v Version) GreaterThan(other Version) bool {
	return v.Set == other.Set &&
		CompareVersions(v, other) == 1
}

// LessThan tests that this version is greater than another.
func (v Version) LessThan(other Version) bool {
	return v.Set == other.Set &&
		CompareVersions(v, other) == -1
}

// CompareVersions compares the numerical components of two versions.
//
// returns:
//
//	-1 if a < b
//	 0 if a == b
//	 1 if a > b
func CompareVersions(a, b Version) int {
	if !a.Set {
		if !b.Set {
			return 0
		}
		return -1
	}

	if !b.Set {
		return 1
	}

	for i := range a.Parts {
		if i < len(b.Parts) {
			if a.Parts[i] > b.Parts[i] {
				return 1
			} else if a.Parts[i] < b.Parts[i] {
				return -1
			}
		} else {
			return 1
		}
	}

	if len(b.Parts) > len(a.Parts) {
		return -1
	} else {
		return 0
	}
}

// ParseNewVersion creates a Version from a string.
func ParseNewVersion(s string) (Version, error) {
	version := Version{Raw: s, Set: false}
	if s != "" {
		if err := version.Parse(); err != nil {
			return version, err
		}
		version.Set = true
	}
	return version, nil
}

// runeInSlice finds a separator in a string. Used as an argument to
// strings.FieldsFunc().
func runeInSlice(needle rune, haystack []rune) bool {
	for _, h := range haystack {
		if needle == h {
			return true
		}
	}
	return false
}

// Valid separators for versions. For example, these are all valid versions:
//
//	`1.2.3.4`
//	`1-2-3`
//	`1-2.3-4.5`
var versionSeparators = []rune{'.', '-'}

// Regex to remove non-numeric characters
var nonNumbers = regexp.MustCompile(`[^\d]+`)

var whitespace = regexp.MustCompile(`\s`)

// Parse version information from a string. Supports the following formats:
//
// - W.X
// - W.X.Y
// - W.X.Y.Z
// - W.X.Y-Z (last components may be separated by a hyphens)
//
// Each component must be integral.
func (v *Version) Parse() error {
	// Separate into components
	pieces := strings.FieldsFunc(v.Raw, func(r rune) bool {
		return runeInSlice(r, versionSeparators)
	})

	var err error
	parts := make([]int, len(pieces))

	// Convert each part to an int
	for i, piece := range pieces {
		// Err if the string contains whitespace
		if whitespace.MatchString(piece) {
			return fmt.Errorf("Unable to parse version that contains whitespace: '%s'", v.Raw)
		}

		// If the string contains non-numeric characters
		numbers := piece
		if nonNumbers.MatchString(piece) {
			// Remove numbers from component
			numbers = nonNumbers.ReplaceAllString(piece, "")
		}

		// Empty check
		if strings.TrimSpace(numbers) == "" {
			numbers = "0"
		}

		// Convert to numeric
		parts[i], err = strconv.Atoi(numbers)
		if err != nil {
			return fmt.Errorf("Unable to parse version: '%s'", v.Raw)
		}

	}
	v.Parts = parts
	if len(parts) > 0 {
		v.Major = parts[0]
	}
	if len(parts) > 1 {
		v.Minor = parts[1]
	}
	if len(parts) > 2 {
		v.Patch = parts[2]
	}
	if len(parts) > 3 {
		v.Rev = parts[3]
	}

	return nil
}

// MarshalJSON satisfies the JSON marshalling interface.
func (v Version) MarshalJSON() ([]byte, error) {
	if !v.Set {
		// marshal invalid version as null
		return []byte(`null`), nil
	} else {
		// marshal other versions normally
		return json.Marshal(v.Raw)
	}
}

// UnmarshalJSON satisfies the JSON unmarshalling interface.
func (v *Version) UnmarshalJSON(data []byte) error {
	var raw string
	if string(data) == "null" {
		v.Set = false
		v.Raw = ""
		return nil
	} else if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	v.Raw = raw
	err := v.Parse()
	if err != nil {
		v.Err = err
		return nil
	}

	v.Set = true
	return nil
}
