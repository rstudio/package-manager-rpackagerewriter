// Copyright (C) 2023 by Posit Software, PBC
package utils

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/rstudio/package-manager-rpackagerewriter/pkg/archive"
	v1 "github.com/rstudio/package-manager-rpackagerewriter/pkg/utils/v1"
	v2 "github.com/rstudio/package-manager-rpackagerewriter/pkg/utils/v2"
)

func TestFilePathGetterSuite(t *testing.T) {
	suite.Run(t, &FilePathGetterSuite{})
}

type FilePathGetterSuite struct {
	suite.Suite
}

func (s *FilePathGetterSuite) TestGet() {
	fpf := NewFilePathGetterFactory()
	s.Require().Equal(&defaultFilePathGetterFactory{}, fpf)

	_, err := fpf.GetFilePathGetter(0)
	s.Require().ErrorContains(err, "Invalid version 0 for GetFilePathGetter")

	fpg, err := fpf.GetFilePathGetter(1)
	s.Require().Nil(err)
	s.Require().IsType(&v1.V1FilePathGetter{}, fpg)

	fpg, err = fpf.GetFilePathGetter(2)
	s.Require().Nil(err)
	s.Require().IsType(&v2.V2V3FilePathGetter{}, fpg)
}

func (s *FilePathGetterSuite) TestGetLocal() {
	fpg := &LocalSourceFilePathGetter{}
	arch := &archive.Results{
		OriginalChecksum:  "012",
		RewrittenChecksum: "123",
	}
	path := fpg.GetFilePath("test/test", arch)
	s.Require().Equal("test/test/012.tar.gz", path)
	readme := fpg.GetReadmePath("test/test", arch)
	s.Require().Equal("test/test/012.readme", readme)
	arch.ReadmeMarkdown = true
	readme = fpg.GetReadmePath("test/test", arch)
	s.Require().Equal("test/test/012.readme.md", readme)
}

func (s *FilePathGetterSuite) TestGetBioc() {
	fpf := NewBiocFilePathGetterFactory()
	s.Require().Equal(&biocFilePathGetterFactory{}, fpf)

	_, err := fpf.GetFilePathGetter(0)
	s.Require().ErrorContains(err, "Invalid version 0 for GetFilePathGetter")

	fpg, err := fpf.GetFilePathGetter(3)
	s.Require().Nil(err)
	s.Require().IsType(&v2.V2V3FilePathGetter{}, fpg)

	fpg, err = fpf.GetFilePathGetter(4)
	s.Require().Nil(err)
	s.Require().IsType(&v2.V2V3FilePathGetter{}, fpg)

	fpg, err = fpf.GetFilePathGetter(5)
	s.Require().Nil(err)
	s.Require().IsType(&v2.V2V3FilePathGetter{}, fpg)

	_, err = fpf.GetFilePathGetter(13)
	s.Require().ErrorContains(err, "Invalid version 13 for GetFilePathGetter")
}
