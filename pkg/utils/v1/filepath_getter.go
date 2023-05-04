// Copyright (C) 2023 by Posit Software, PBC
package v1

import (
	"path/filepath"

	"github.com/rstudio/package-manager-rewriting/pkg/archive"
)

type V1FilePathGetter struct{}

func (g *V1FilePathGetter) GetFilePath(outputDir string, arch *archive.ArchiveResults) string {
	return filepath.Join(outputDir, arch.RewrittenChecksum+".tar.gz")
}

func (g *V1FilePathGetter) GetReadmePath(outputDir string, arch *archive.ArchiveResults) string {
	ext := ".readme"
	if arch.ReadmeMarkdown {
		ext += ".md"
	}
	return filepath.Join(outputDir, arch.RewrittenChecksum+ext)
}
