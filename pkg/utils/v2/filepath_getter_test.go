// Copyright (C) 2023 by Posit Software, PBC
package v2

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/rstudio/package-manager-rewriting/pkg/archive"
)

func TestFilePathGetterSuite(t *testing.T) {
	suite.Run(t, &FilePathGetterSuite{})
}

type FilePathGetterSuite struct {
	suite.Suite
}

func (s *FilePathGetterSuite) TestGet() {
	fpg := &V2V3FilePathGetter{}
	arch := &archive.ArchiveResults{
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
