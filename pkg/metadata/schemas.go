// Copyright (C) 2023 by Posit Software, PBC
package metadata

import "github.com/rstudio/package-manager-rpackagerewriter/pkg/version"

var validSchemas = []int{
	1,
	2,
	3,
	4,
}

var validBiocSchemas = []int{
	3,
	4,
	5,
}

func IsValidSchema(v version.ManifestVersion) bool {
	for _, schema := range validSchemas {
		if v.Schema == schema {
			return true
		}
	}
	return false
}

func IsValidBiocSchema(v version.ManifestVersion) bool {
	for _, schema := range validBiocSchemas {
		if v.Schema == schema {
			return true
		}
	}
	return false
}
