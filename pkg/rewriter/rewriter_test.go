// Copyright (C) 2023 by Posit Software, PBC
package rewriter

import (
	"bytes"
	"errors"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/rstudio/package-manager-rpackagerewriter/internal/test"
	"github.com/rstudio/package-manager-rpackagerewriter/pkg/utils"
)

func TestRewriterSuite(t *testing.T) {
	suite.Run(t, &RewriterSuite{})
}

type RewriterSuite struct {
	suite.Suite
}

func (s *RewriterSuite) TestNewArchiveRewriter() {
	dir, err := os.MkdirTemp("", "")
	s.Require().Nil(err)
	fpg, err := utils.NewFilePathGetterFactory().GetFilePathGetter(1)
	s.Require().Nil(err)
	r := NewRPackageRewriter("outputDir", "readmeDir", dir, fpg, 256, 6)
	s.Require().Equal(&rPackageRewriter{
		OutputDir:       "outputDir",
		ReadmeOutputDir: "readmeDir",
		tempDir:         dir,
		fpg:             fpg,
		bufferSize:      256,
		gzipLevel:       6,
	}, r)
}

func (s *RewriterSuite) TestArchiveRewriterRewrite() {
	dir, _ := os.MkdirTemp("", "")
	readmeDir, err := os.MkdirTemp("", "readme")
	s.Require().Nil(err)
	fpg, err := utils.NewFilePathGetterFactory().GetFilePathGetter(1)
	s.Require().Nil(err)
	rewriter := NewRPackageRewriter(dir, readmeDir, dir, fpg, 1024*2, 6)
	results, err := rewriter.Rewrite("../testdata/adhoc_1.1.tar.gz")
	s.Require().Nil(err)
	s.Require().Equal(int64(27126), results.OriginalSize)
	s.Require().Equal(int64(26528), results.RewrittenSize)
	s.Require().Equal("fe809c8dc025d992ac0615ba3bf8414c68f4b5b41dda1becbae7332c58fada3a", results.RewrittenChecksum)
	s.Require().Equal("7b87295f1256a12b958bc2ab6d94a45e310a73b30721d0fe271033fad276a928", results.OriginalChecksum)
	s.Require().Equal(`Package: adhoc
Version: 1.1
Type: Package
Title: Calculate Ad Hoc Distance Thresholds for DNA Barcoding
        Identification
Date: 2017-03-17
Author: Gontran Sonet
Maintainer: Gontran Sonet <gosonet@gmail.com>
Description: Two functions to calculate intra- and interspecific pairwise distances, evaluate DNA barcoding identification error and calculate an ad hoc distance threshold for each particular reference library of DNA barcodes. Specimen identification at this ad hoc distance threshold (using the best close match method) will produce identifications with an estimated relative error probability that can be fixed by the user (e.g. 5%).
URL: http://jemu.myspecies.info/computer-programs
Depends: R (>= 2.15), ape, pegas, polynom
License: GPL (>= 2)
NeedsCompilation: no
Packaged: 2017-03-17 14:12:04 UTC; gsonet
Repository: RSPM
Date/Publication: 2017-03-17 18:56:42 UTC
Encoding: UTF-8
`, results.Description)

	// readme file should not exist
	_, err = os.Stat(filepath.Join(readmeDir, results.RewrittenChecksum+".readme.md"))
	s.Require().Equal(true, os.IsNotExist(err))
	_, err = os.Stat(filepath.Join(readmeDir, results.RewrittenChecksum+".readme"))
	s.Require().Equal(true, os.IsNotExist(err))
}

func (s *RewriterSuite) TestArchiveRewriterRewriteStream() {
	dir, _ := os.MkdirTemp("", "")
	readmeDir, err := os.MkdirTemp("", "readme")
	s.Require().Nil(err)
	fpg, err := utils.NewFilePathGetterFactory().GetFilePathGetter(1)
	s.Require().Nil(err)
	rewriter := NewRPackageRewriter(dir, readmeDir, dir, fpg, 1024*2, 6)
	f, err := os.Open("../testdata/adhoc_1.1.tar.gz")
	s.Require().Nil(err)
	w := bytes.NewBuffer([]byte{})
	results, err := rewriter.RewriteStream(f, w)
	s.Require().Nil(err)
	s.Require().Equal(int64(27126), results.OriginalSize)
	s.Require().Equal(int64(26528), results.RewrittenSize)
	s.Require().Equal("fe809c8dc025d992ac0615ba3bf8414c68f4b5b41dda1becbae7332c58fada3a", results.RewrittenChecksum)
	s.Require().Equal("7b87295f1256a12b958bc2ab6d94a45e310a73b30721d0fe271033fad276a928", results.OriginalChecksum)
	s.Require().Equal(`Package: adhoc
Version: 1.1
Type: Package
Title: Calculate Ad Hoc Distance Thresholds for DNA Barcoding
        Identification
Date: 2017-03-17
Author: Gontran Sonet
Maintainer: Gontran Sonet <gosonet@gmail.com>
Description: Two functions to calculate intra- and interspecific pairwise distances, evaluate DNA barcoding identification error and calculate an ad hoc distance threshold for each particular reference library of DNA barcodes. Specimen identification at this ad hoc distance threshold (using the best close match method) will produce identifications with an estimated relative error probability that can be fixed by the user (e.g. 5%).
URL: http://jemu.myspecies.info/computer-programs
Depends: R (>= 2.15), ape, pegas, polynom
License: GPL (>= 2)
NeedsCompilation: no
Packaged: 2017-03-17 14:12:04 UTC; gsonet
Repository: RSPM
Date/Publication: 2017-03-17 18:56:42 UTC
Encoding: UTF-8
`, results.Description)
	s.Require().Equal(26528, w.Len())

	// readme file should not exist
	_, err = os.Stat(filepath.Join(readmeDir, results.RewrittenChecksum+".readme.md"))
	s.Require().Equal(true, os.IsNotExist(err))
	_, err = os.Stat(filepath.Join(readmeDir, results.RewrittenChecksum+".readme"))
	s.Require().Equal(true, os.IsNotExist(err))
}

func (s *RewriterSuite) TestArchiveRewriterGetNoReadme() {
	dir, err := os.MkdirTemp("", "")
	s.Require().Nil(err)
	readmeDir, err := os.MkdirTemp("", "readme")
	s.Require().Nil(err)
	fpg, err := utils.NewFilePathGetterFactory().GetFilePathGetter(1)
	s.Require().Nil(err)
	rewriter := NewRPackageRewriter("", readmeDir, dir, fpg, 256, 6)
	f, err := os.Open("../testdata/adhoc_1.1.tar.gz")
	s.Require().Nil(err)
	archive, err := rewriter.GetReadme(f)
	s.Require().Nil(err)

	// readme file should not exist
	_, err = os.Stat(filepath.Join(readmeDir, archive.RewrittenChecksum+".readme.md"))
	s.Require().Equal(true, os.IsNotExist(err))
	_, err = os.Stat(filepath.Join(readmeDir, archive.RewrittenChecksum+".readme"))
	s.Require().Equal(true, os.IsNotExist(err))
}

func (s *RewriterSuite) TestArchiveRewriterRewriteReadme() {
	dir, err := os.MkdirTemp("", "")
	s.Require().Nil(err)
	readmeDir, err := os.MkdirTemp("", "readme")
	s.Require().Nil(err)
	fpg, err := utils.NewFilePathGetterFactory().GetFilePathGetter(2)
	s.Require().Nil(err)
	rewriter := NewRPackageRewriter(dir, readmeDir, dir, fpg, 256, 6)
	archive, err := rewriter.Rewrite("../testdata/readmetest_0.2.0.tar.gz")
	s.Require().Nil(err)
	s.Require().Equal(`Package: readmetest
Type: Package
Title: Tests that the correct README file is found
Version: 0.2.0
Date: 2015-08-15
Author: Who wrote it
Maintainer: The package maintainer <yourself@somewhere.net>
Description: More about what it does (maybe more than one line) Use
        four spaces when indenting paragraphs within the Description.
License: What license is it under?
Encoding: UTF-8
LazyData: true
NeedsCompilation: no
Packaged: 2017-09-07 17:17:15 UTC; jon
Date/Publication: 2015-08-16 23:05:52
Repository: RSPM
`, archive.Description)

	// Find readme file
	readmePath := filepath.Join(readmeDir, archive.OriginalChecksum+".readme.md")
	s.Require().Equal(readmePath, archive.ExtractedReadmePath)
	readme, err := os.ReadFile(readmePath)
	s.Require().Nil(err)
	test.TestifyGolden(string(readme), &s.Suite)
}

func (s *RewriterSuite) TestArchiveRewriterRewriteReadmeStream() {
	dir, _ := os.MkdirTemp("", "")
	readmeDir, err := os.MkdirTemp("", "readme")
	s.Require().Nil(err)
	fpg, err := utils.NewFilePathGetterFactory().GetFilePathGetter(2)
	s.Require().Nil(err)
	rewriter := NewRPackageRewriter(dir, readmeDir, dir, fpg, 256, 6)
	f, err := os.Open("../testdata/readmetest_0.2.0.tar.gz")
	s.Require().Nil(err)
	w := bytes.NewBuffer([]byte{})
	archive, err := rewriter.RewriteStream(f, w)
	s.Require().Nil(err)
	s.Require().Equal(`Package: readmetest
Type: Package
Title: Tests that the correct README file is found
Version: 0.2.0
Date: 2015-08-15
Author: Who wrote it
Maintainer: The package maintainer <yourself@somewhere.net>
Description: More about what it does (maybe more than one line) Use
        four spaces when indenting paragraphs within the Description.
License: What license is it under?
Encoding: UTF-8
LazyData: true
NeedsCompilation: no
Packaged: 2017-09-07 17:17:15 UTC; jon
Date/Publication: 2015-08-16 23:05:52
Repository: RSPM
`, archive.Description)

	// Find readme file
	readmePath := filepath.Join(readmeDir, archive.OriginalChecksum+".readme.md")
	s.Require().Equal(readmePath, archive.ExtractedReadmePath)
	readme, err := os.ReadFile(readmePath)
	s.Require().Nil(err)
	test.TestifyGolden(string(readme), &s.Suite)
}

func (s *RewriterSuite) TestArchiveRewriterGetReadme() {
	dir, err := os.MkdirTemp("", "")
	s.Require().Nil(err)
	readmeDir, err := os.MkdirTemp("", "readme")
	s.Require().Nil(err)
	fpg, err := utils.NewFilePathGetterFactory().GetFilePathGetter(2)
	s.Require().Nil(err)
	rewriter := NewRPackageRewriter("", readmeDir, dir, fpg, 256, 6)
	f, err := os.Open("../testdata/readmetest_0.2.0.tar.gz")
	s.Require().Nil(err)
	archive, err := rewriter.GetReadme(f)
	s.Require().Nil(err)

	// Find readme file
	readmePath := filepath.Join(readmeDir, archive.OriginalChecksum+".readme.md")
	s.Require().Equal(readmePath, archive.ExtractedReadmePath)
	readme, err := os.ReadFile(readmePath)
	s.Require().Nil(err)
	test.TestifyGolden(string(readme), &s.Suite)
}

func (s *RewriterSuite) TestArchiveRewriterCleanup() {
	dir, _ := os.MkdirTemp("", "")
	readmeDir, err := os.MkdirTemp("", "readme")
	s.Require().Nil(err)
	fpg, err := utils.NewFilePathGetterFactory().GetFilePathGetter(2)
	s.Require().Nil(err)
	rewriter := NewRPackageRewriter(dir, readmeDir, dir, fpg, 256, 6)
	// Attempt will fail since ff_2.2-14.zip is not an archive
	_, err = rewriter.Rewrite("../testdata/ff_2.2-14.zip")
	s.Require().ErrorContains(err, "error rewriting")

	// Ensure that the output directories are empty
	files, _ := os.ReadDir(dir)
	for _, f := range files {
		log.Printf("Found unexpected temp file %s", f.Name())
	}
	s.Require().Len(files, 0)

	readmes, _ := os.ReadDir(readmeDir)
	for _, f := range readmes {
		log.Printf("Found unexpected readme temp file %s", f.Name())
	}
	s.Require().Len(readmes, 0)
}

func (s *RewriterSuite) TestArchiveRewriterStreamCleanup() {
	dir, _ := os.MkdirTemp("", "")
	readmeDir, err := os.MkdirTemp("", "readme")
	s.Require().Nil(err)
	fpg, err := utils.NewFilePathGetterFactory().GetFilePathGetter(2)
	s.Require().Nil(err)
	rewriter := NewRPackageRewriter(dir, readmeDir, dir, fpg, 256, 6)
	f, err := os.Open("../testdata/ff_2.2-14.zip")
	s.Require().Nil(err)
	w := bytes.NewBuffer([]byte{})
	// Attempt will fail since ff_2.2-14.zip is not an archive
	_, err = rewriter.RewriteStream(f, w)
	s.Require().ErrorContains(err, "error rewriting")

	// Ensure that the output directories are empty
	files, _ := os.ReadDir(dir)
	for _, f := range files {
		log.Printf("Found unexpected temp file %s", f.Name())
	}
	s.Require().Len(files, 0)

	readmes, _ := os.ReadDir(readmeDir)
	for _, f := range readmes {
		log.Printf("Found unexpected readme temp file %s", f.Name())
	}
	s.Require().Len(readmes, 0)
}

func (s *RewriterSuite) TestArchiveRewriterRewriteBinaryTar() {
	dir, _ := os.MkdirTemp("", "")
	readmeDir, err := os.MkdirTemp("", "readme")
	s.Require().Nil(err)
	fpg, err := utils.NewFilePathGetterFactory().GetFilePathGetter(1)
	s.Require().Nil(err)
	rewriter := NewRPackageRewriter(dir, readmeDir, dir, fpg, 1024*2, 6)
	f, err := os.Open("../testdata/binaries/bindrcpp_0.2.2.tar.gz")
	s.Require().Nil(err)
	w := bytes.NewBuffer([]byte{})
	results, err := rewriter.RewriteBinary(f, w, false)
	s.Require().Nil(err)
	s.Require().Equal(int64(813035), results.OriginalSize)
	s.Require().Equal(int64(831722), results.RewrittenSize)
	s.Require().Equal("e1fb9e6b8bb48414925a16fa88da69a88d85b16fc06667469f79fdc472db7c0b", results.RewrittenChecksum)
	s.Require().Equal("269c628012cc0a4a38878550b64bacd8665554d46d966ee502c21c7e1161bf8b", results.OriginalChecksum)
	s.Require().Equal(`Package: bindrcpp
Title: An 'Rcpp' Interface to Active Bindings
Version: 0.2.2
Date: 2018-03-29
Authors@R: c(
    person("Kirill", "M\u00fcller", role = c("aut", "cre"), email = "krlmlr+r@mailbox.org", comment = c(ORCID = "0000-0002-1416-3412")),
    person("RStudio", role = "cph")
    )
Description: Provides an easy way to fill an environment with active bindings
    that call a C++ function.
License: MIT + file LICENSE
URL: https://github.com/krlmlr/bindrcpp,
        https://krlmlr.github.io/bindrcpp
BugReports: https://github.com/krlmlr/bindrcpp/issues
Imports: bindr (>= 0.1.1), Rcpp (>= 0.12.16)
Suggests: testthat
LinkingTo: plogr, Rcpp
Encoding: UTF-8
LazyData: true
RoxygenNote: 6.0.1.9000
NeedsCompilation: yes
Packaged: 2018-03-29 13:33:00 UTC; muelleki
Author: Kirill Müller [aut, cre] (<https://orcid.org/0000-0002-1416-3412>),
  RStudio [cph]
Maintainer: Kirill Müller <krlmlr+r@mailbox.org>
Repository: RSPM
Date/Publication: 2018-03-29 14:09:09 UTC
Built: R 4.2.0; x86_64-pc-linux-gnu; 2022-04-24 04:16:10 UTC; unix
`, results.Description)
	s.Require().Equal(831722, w.Len())
}

func (s *RewriterSuite) TestArchiveRewriterRewriteBinaryTarNoDescription() {
	dir, _ := os.MkdirTemp("", "")
	readmeDir, err := os.MkdirTemp("", "readme")
	s.Require().Nil(err)
	fpg, err := utils.NewFilePathGetterFactory().GetFilePathGetter(1)
	s.Require().Nil(err)
	rewriter := NewRPackageRewriter(dir, readmeDir, dir, fpg, 1024*2, 6)
	f, err := os.Open("../testdata/binaries/bindrcpp_0.2.2-no-desc.tar.gz")
	s.Require().Nil(err)
	w := bytes.NewBuffer([]byte{})
	_, err = rewriter.RewriteBinary(f, w, false)
	s.Require().ErrorContains(err, "error rewriting stream: no DESCRIPTION file found in archive")
	s.Require().Equal(true, errors.Is(err, RPackageRewriteError{}))
}

func (s *RewriterSuite) TestArchiveRewriterRewriteBinaryZip() {
	dir, _ := os.MkdirTemp("", "")
	readmeDir, err := os.MkdirTemp("", "readme")
	s.Require().Nil(err)
	fpg, err := utils.NewFilePathGetterFactory().GetFilePathGetter(1)
	s.Require().Nil(err)
	rewriter := NewRPackageRewriter(dir, readmeDir, dir, fpg, 1024*2, 6)
	f, err := os.Open("../testdata/binaries/bindrcpp_0.2.2.zip")
	s.Require().Nil(err)
	w := bytes.NewBuffer([]byte{})
	results, err := rewriter.RewriteBinary(f, w, true)
	s.Require().Nil(err)
	s.Require().Equal(int64(412918), results.OriginalSize)
	s.Require().Equal(int64(431367), results.RewrittenSize)
	s.Require().Equal("2b54feb73af798bda39e0c68fd13c58279147df77d870b502195d3a375c0e477", results.RewrittenChecksum)
	s.Require().Equal("dc4387dcd7a5ba5f778f2139121bc81dea5a44a0c2adb19a0c09dbff17e1247a", results.OriginalChecksum)
	s.Require().Equal(`Package: bindrcpp
Title: An 'Rcpp' Interface to Active Bindings
Version: 0.2.2
Date: 2018-03-29
Authors@R: c(
    person("Kirill", "M\u00fcller", role = c("aut", "cre"), email = "krlmlr+r@mailbox.org", comment = c(ORCID = "0000-0002-1416-3412")),
    person("RStudio", role = "cph")
    )
Description: Provides an easy way to fill an environment with active bindings
    that call a C++ function.
License: MIT + file LICENSE
URL: https://github.com/krlmlr/bindrcpp,
        https://krlmlr.github.io/bindrcpp
BugReports: https://github.com/krlmlr/bindrcpp/issues
Imports: bindr (>= 0.1.1), Rcpp (>= 0.12.16)
Suggests: testthat
LinkingTo: plogr, Rcpp
Encoding: UTF-8
LazyData: true
RoxygenNote: 6.0.1.9000
NeedsCompilation: yes
Packaged: 2018-03-29 13:33:00 UTC; muelleki
Author: Kirill Müller [aut, cre] (<https://orcid.org/0000-0002-1416-3412>),
  RStudio [cph]
Maintainer: Kirill Müller <krlmlr+r@mailbox.org>
Date/Publication: 2018-03-29 14:09:09 UTC
Built: R 4.2.0; x86_64-w64-mingw32; 2022-04-29 06:29:52 UTC; windows
ExperimentalWindowsRuntime: ucrt
Archs: x64
Repository: RSPM
`, results.Description)
	s.Require().Equal(431367, w.Len())
}
