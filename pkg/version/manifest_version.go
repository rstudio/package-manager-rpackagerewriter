// Copyright (C) 2023 by Posit Software, PBC
package version

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ManifestVersion represents semantic version information with each component as an int.
// Other types of version strings are not supported.
type ManifestVersion struct {
	Raw    string
	Schema int
	Major  int
	Minor  int
	Parts  []int
	Set    bool
}

// String generates a string representation of the version.
func (v ManifestVersion) String() string {
	return v.Raw
}

// Equals tests that this version is equivalent to another.
func (v ManifestVersion) Equals(other ManifestVersion) bool {
	return v.Raw == other.Raw &&
		v.Set == other.Set
}

// EqualsMajor tests that this version is equivalent to another in major version.
func (v ManifestVersion) EqualsMajor(other ManifestVersion) bool {
	return v.Set == other.Set &&
		CompareSchemaMajor(v, other) == 0
}

// GreaterThan tests that this version is greater than another.
func (v ManifestVersion) GreaterThan(other ManifestVersion) bool {
	return v.Set == other.Set &&
		CompareManifestVersions(v, other) == 1
}

// LessThan tests that this version is greater than another.
func (v ManifestVersion) LessThan(other ManifestVersion) bool {
	return v.Set == other.Set &&
		CompareManifestVersions(v, other) == -1
}

// LessThanOrEqual tests that this version is less than or equal to another.
func (v ManifestVersion) LessThanOrEqual(other ManifestVersion) bool {
	return v.Set == other.Set &&
		CompareManifestVersions(v, other) <= 0
}

// LessThanMajor tests that this version is greater in schema/major version than another
func (v ManifestVersion) LessThanMajor(other ManifestVersion) bool {
	return v.Set == other.Set &&
		CompareSchemaMajor(v, other) == -1
}

// CompareManifestVersions compares the numerical components of two versions.
//
// returns:
//
//	-1 if a < b
//	 0 if a == b
//	 1 if a > b
func CompareManifestVersions(a, b ManifestVersion) int {
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

// CompareSchemaMajor compares schema, major (NOT minor) of two versions.
//
// returns:
//
//	-1 iff a < b
//	 0 iff a == b
//	 1 iff a > b
//
// The schema and major version are sorted numerically.
func CompareSchemaMajor(a, b ManifestVersion) int {
	if a.Schema < b.Schema {
		return -1
	} else if a.Schema > b.Schema {
		return 1
	} // else same schema

	if a.Major < b.Major {
		return -1
	} else if a.Major > b.Major {
		return 1
	} // else same major

	return 0
}

// ParseNewManifestVersion creates a ManifestVersion from a string.
func ParseNewManifestVersion(s string) (ManifestVersion, error) {
	version := ManifestVersion{Raw: s, Set: false}
	if s != "" {
		if err := version.Parse(); err != nil {
			return version, err
		}
		version.Set = true
	}
	return version, nil
}

func NewManifestVersion(schema, major, minor int) ManifestVersion {
	return ManifestVersion{
		Schema: schema,
		Major:  major,
		Minor:  minor,
		Set:    true,
		Parts:  []int{schema, major, minor},
		Raw:    fmt.Sprintf("v%d/%d/%d", schema, major, minor),
	}
}

// ManifestVersionSeparators are valid separators for versions. For example, these are all valid versions:
//
//	`1.2.3.4`
//	`1-2-3`
//	`1-2.3-4.5`
var ManifestVersionSeparators = []rune{'/', '.'}

// Parse version information from a string. Supports the following formats:
//
// - W.X
// - W.X.Y
// - W.X.Y.Z
// - W.X.Y-Z (last components may be separated by a hyphens)
//
// Each component must be integral.
func (v *ManifestVersion) Parse() error {
	// Separate into components
	pieces := strings.FieldsFunc(v.Raw, func(r rune) bool {
		return runeInSlice(r, ManifestVersionSeparators)
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
		v.Schema = parts[0]
	}
	if len(parts) > 1 {
		v.Major = parts[1]
	}
	if len(parts) > 2 {
		v.Minor = parts[2]
	}

	return nil
}

// MarshalJSON satisfies the JSON marshalling interface.
func (v ManifestVersion) MarshalJSON() ([]byte, error) {
	if !v.Set {
		// marshal invalid version as null
		return []byte(`null`), nil
	} else {
		// marshal other versions normally
		return json.Marshal(v.Raw)
	}
}

// UnmarshalJSON satisfies the JSON unmarshalling interface.
func (v *ManifestVersion) UnmarshalJSON(data []byte) error {
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
		return err
	}

	v.Set = true
	return nil
}
