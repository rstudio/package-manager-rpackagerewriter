// Copyright (C) 2023 by Posit Software, PBC
package utils

import (
	"fmt"
	"path/filepath"

	"github.com/rstudio/package-manager-rpackagerewriter/pkg/archive"
	v1 "github.com/rstudio/package-manager-rpackagerewriter/pkg/utils/v1"
	v2 "github.com/rstudio/package-manager-rpackagerewriter/pkg/utils/v2"
)

type FilePathGetter interface {
	GetFilePath(outputDir string, arch *archive.Results) string
	GetReadmePath(outputDir string, arch *archive.Results) string
}

type FilePathGetterFactory interface {
	GetFilePathGetter(schemaVersion int) (FilePathGetter, error)
}

type defaultFilePathGetterFactory struct{}
type biocFilePathGetterFactory struct{}

func NewFilePathGetterFactory() FilePathGetterFactory {
	return &defaultFilePathGetterFactory{}
}

func NewBiocFilePathGetterFactory() FilePathGetterFactory {
	return &biocFilePathGetterFactory{}
}

func (f *defaultFilePathGetterFactory) GetFilePathGetter(schemaVersion int) (FilePathGetter, error) {
	switch schemaVersion {
	case 1:
		return &v1.V1FilePathGetter{}, nil
	case 2, 3, 4:
		return &v2.V2V3FilePathGetter{}, nil
	}
	return nil, fmt.Errorf("Invalid version %d for GetFilePathGetter", schemaVersion)
}

func (f *biocFilePathGetterFactory) GetFilePathGetter(schemaVersion int) (FilePathGetter, error) {
	switch schemaVersion {
	case 3, 4, 5:
		// Bioc was originally created at parity with CRAN at version 3. Version 4 of bioc was a simple copy
		//    of version 3 that was used to force a fresh sync so that a fix in the sync code would be picked
		//    up. For both of these reasons, the Bioc transformer here is the same as the CRAN V2V3FilePathGetter.
		//    See https://github.com/rstudio/package-manager/issues/5901 for history.
		return &v2.V2V3FilePathGetter{}, nil
	}
	return nil, fmt.Errorf("Invalid version %d for GetFilePathGetter", schemaVersion)
}

// LocalSourceFilePathGetter is used by local sources, which don't distinguish on schema versions
type LocalSourceFilePathGetter struct{}

func (g *LocalSourceFilePathGetter) GetFilePath(outputDir string, arch *archive.Results) string {
	return filepath.Join(outputDir, arch.OriginalChecksum+".tar.gz")
}

func (g *LocalSourceFilePathGetter) GetReadmePath(outputDir string, arch *archive.Results) string {
	ext := ".readme"
	if arch.ReadmeMarkdown {
		ext += ".md"
	}
	return filepath.Join(outputDir, arch.OriginalChecksum+ext)
}
