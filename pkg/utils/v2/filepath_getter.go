// Copyright (C) 2023 by Posit Software, PBC
package v2

import (
	"path/filepath"

	"github.com/rstudio/package-manager-rewriting/pkg/archive"
)

type V2V3FilePathGetter struct{}

func (g *V2V3FilePathGetter) GetFilePath(outputDir string, arch *archive.Results) string {
	return filepath.Join(outputDir, arch.OriginalChecksum+".tar.gz")
}

func (g *V2V3FilePathGetter) GetReadmePath(outputDir string, arch *archive.Results) string {
	ext := ".readme"
	if arch.ReadmeMarkdown {
		ext += ".md"
	}
	return filepath.Join(outputDir, arch.OriginalChecksum+ext)
}
