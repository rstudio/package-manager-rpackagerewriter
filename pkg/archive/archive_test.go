// Copyright (C) 2023 by Posit Software, PBC
package archive

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/rstudio/package-manager-rpackagerewriter/internal/test"
)

func TestArchiveSuite(t *testing.T) {
	suite.Run(t, &ArchiveSuite{})
}

type ArchiveSuite struct {
	suite.Suite
}

func (s *ArchiveSuite) TestNewArchive() {
	a := NewRPackageArchive(256, 6)
	s.Require().Equal(&RPackageArchive{
		bufferSize: 256,
		gzipLevel:  6,
	}, a)
}

func (s *ArchiveSuite) TestDescriptionRewrite() {
	a := NewRPackageArchive(256, 6)
	f, err := os.Open("../testdata/DT_0.4.tar.gz")
	s.Require().Nil(err)

	var b bytes.Buffer
	var bReadme bytes.Buffer
	results, err := a.RewriteWithReadme(f, &b, &bReadme)
	s.Require().Nil(err)

	// Make sure we parsed the correct DESCRIPTION file.
	s.Require().Equal("3daa96b819ca54e5fbc2c7d78cb3637982a2d44be58cea0683663b71cfc7fa19", results.OriginalChecksum)
	s.Require().Equal("f320e15acff2a42ef6db261ff71d747245d840d00adf340a24db5e92ed0cf6ca", results.RewrittenChecksum)
	s.Require().Equal(855386, b.Len())

	// Check the contents of the DESCRIPTION and MD5 files.
	md5File, err := StreamFileFromTarGz(&b, "MD5")
	s.Require().Nil(err)
	var m bytes.Buffer
	_, _ = m.ReadFrom(md5File)
	_, err = io.Copy(io.Discard, md5File)
	s.Require().Nil(err)
	test.TestifyGolden(results.Description+"\n\n"+m.String(), &s.Suite)
}

func (s *ArchiveSuite) TestDescriptionRewriteBinaryInvalid() {
	a := NewRPackageArchive(256, 6)

	tmp, err := os.CreateTemp("", "")
	s.Require().Nil(err)
	defer func(tmp *os.File) {
		_ = tmp.Close()
	}(tmp)

	_, err = tmp.WriteString("this is a test file\nwith no gzip or tar content\n.")
	s.Require().Nil(err)
	_, err = tmp.Seek(0, 0)
	s.Require().Nil(err)

	var b bytes.Buffer
	_, err = a.RewriteBinary(tmp, &b)
	s.Require().ErrorContains(err, "gzip: invalid header")
}

func (s *ArchiveSuite) TestDescriptionRewriteBinary() {
	a := NewRPackageArchive(256, 6)
	f, err := os.Open("../testdata/binaries/bindrcpp_0.2.2.tar.gz")
	s.Require().Nil(err)

	var b bytes.Buffer
	results, err := a.RewriteBinary(f, &b)
	s.Require().Nil(err)

	// Make sure we parsed the correct DESCRIPTION file.
	s.Require().Equal("269c628012cc0a4a38878550b64bacd8665554d46d966ee502c21c7e1161bf8b", results.OriginalChecksum)
	s.Require().Equal("e1fb9e6b8bb48414925a16fa88da69a88d85b16fc06667469f79fdc472db7c0b", results.RewrittenChecksum)
	s.Require().Equal(831722, b.Len())

	// Check the contents of the DESCRIPTION
	test.TestifyGolden(results.Description, &s.Suite)
}

func (s *ArchiveSuite) TestDescriptionRewriteBinary2() {
	a := NewRPackageArchive(256, 6)
	f, err := os.Open("../testdata/binaries/DT_0.23.tar.gz")
	s.Require().Nil(err)

	var b bytes.Buffer
	results, err := a.RewriteBinary(f, &b)
	s.Require().Nil(err)

	// Make sure we parsed the correct DESCRIPTION file.
	s.Require().Equal("5112b7459392834152f25d5e52073ab75f47cd1d89159dca6677fdc7294339da", results.OriginalChecksum)
	s.Require().Equal("02f5662604de07cc36aa95870a6f85376e440f7f4ab657e3435508a8cb8f2625", results.RewrittenChecksum)
	s.Require().Equal(1583810, b.Len())

	// Check the contents of the DESCRIPTION
	test.TestifyGolden(results.Description, &s.Suite)
}

// Tests rewriting a package where the DESCRIPTION uses the `latin1` character
// encoding instead of UTF-8. We expect that
// * The DESCRIPTION is needs to retain the `Encoding: latin1`
// * The DESCRIPTION is written using the `latin1` encoding instead of the default `UTF-8`
func (s *ArchiveSuite) TestDescriptionRewriteLatin1() {
	a := NewRPackageArchive(256, 6)
	f, err := os.Open("../testdata/special/Latin1SpecialChars_1.1.1.tar.gz")
	s.Require().Nil(err)

	var b bytes.Buffer
	var bReadme bytes.Buffer
	results, err := a.RewriteWithReadme(f, &b, &bReadme)
	s.Require().Nil(err)

	// Make sure we parsed the correct DESCRIPTION file.
	s.Require().Equal("d4088da9ed8d2d3979ac722528bdd13d5d0d03fbba9b6f1cc20a15b390185925", results.OriginalChecksum)
	s.Require().Equal("5cb60050250cd113192c767ca366e642648e2b308d3a581403e1a428f9258cc7", results.RewrittenChecksum)
	s.Require().Equal(679, b.Len())

	// Check the contents of the DESCRIPTION and MD5 files.
	md5File, err := StreamFileFromTarGz(&b, "MD5")
	s.Require().Nil(err)
	var m bytes.Buffer
	_, _ = m.ReadFrom(md5File)
	_, err = io.Copy(io.Discard, md5File)
	s.Require().Nil(err)
	test.TestifyGolden(results.Description+"\n\n"+m.String(), &s.Suite)

}

// Tests rewriting a package that includes a second MD5 file. Only the root
// MD5 file should be rewritten with the new DESCRIPTION checksum.
func (s *ArchiveSuite) TestDescriptionRewriteSecondMD5() {
	// SecondMD5_2.2.2 is a package that was assembled in a particular
	// order in the tarball archive. The key is that the file
	// `tests/testthat/fixtures/MD5` must be listed before the root `MD5`
	// file in the tarball header. This causes the header iteration to
	// encounter the file `tests/testthat/fixtures/MD5` first; we then
	// test to ensure that only the second MD5 file (at the package root)
	// is rewritten even though the other one was encountered first.
	a := NewRPackageArchive(256, 6)
	f, err := os.Open("../testdata/special/SecondMD5_2.2.2.tar.gz")
	s.Require().Nil(err)

	tarball, err := os.CreateTemp("", "")
	s.Require().Nil(err)
	defer func(name string) {
		_ = os.Remove(name)
	}(tarball.Name())

	var bReadme bytes.Buffer
	_, err = a.RewriteWithReadme(f, tarball, &bReadme)
	s.Require().Nil(err)

	// Check the contents of the following files:
	// * DESCRIPTION
	// * MD5
	// * tests/testthat/fixtures/MD5
	_, _ = tarball.Seek(0, 0)
	descFile, err := StreamFileFromTarGz(tarball, "SecondMD5/DESCRIPTION")
	s.Require().Nil(err)
	var m bytes.Buffer
	_, _ = m.ReadFrom(descFile)
	//
	_, _ = tarball.Seek(0, 0)
	md5File, err := StreamFileFromTarGz(tarball, "SecondMD5/MD5")
	s.Require().Nil(err)
	m.WriteString("\n")
	_, _ = m.ReadFrom(md5File)
	//
	// Read the second MD5 file, which should not have been rewritten
	_, _ = tarball.Seek(0, 0)
	testMd5File, err := StreamFileFromTarGz(tarball, "SecondMD5/tests/testthat/fixtures/MD5")
	s.Require().Nil(err)
	//
	// An explicit check for the expected contents of the file at
	// tests/testthat/fixtures/MD5. We need to ensure that this file
	// is not overwritten.
	var testMd5FileContents bytes.Buffer
	_, _ = testMd5FileContents.ReadFrom(testMd5File)
	m.WriteString("\n")
	m.WriteString(testMd5FileContents.String())
	expected := "186a2fa26490b7d2b08ff66547033617 *DESCRIPTION\n" +
		"67dae3f768de2008f9b522f0f3f94752 *R/sample.R\n"
	s.Require().Equal(expected, testMd5FileContents.String())
	//
	// Check against the golden recorded data
	test.TestifyGolden(m.String(), &s.Suite)

}

func (s *ArchiveSuite) TestReadmeResolution1() {
	a := NewRPackageArchive(256, 6)
	f, err := os.Open("../testdata/readmetest_0.2.0.tar.gz")
	s.Require().Nil(err)

	var b bytes.Buffer
	var bReadme bytes.Buffer
	results, err := a.RewriteWithReadme(f, &b, &bReadme)
	s.Require().Nil(err)

	// Make sure we parsed the correct README file.
	s.Require().Equal(true, results.ReadmeMarkdown)
	s.Require().Equal("Hi, I'm the correct readme!", bReadme.String())
}

func (s *ArchiveSuite) TestReadmeResolutionNone() {
	a := NewRPackageArchive(256, 6)
	f, err := os.Open("../testdata/hasdependencies_0.2.0.tar.gz")
	s.Require().Nil(err)

	var b bytes.Buffer
	var bReadme bytes.Buffer
	results, err := a.RewriteWithReadme(f, &b, &bReadme)
	s.Require().Nil(err)

	// Make sure we didn't find a README
	s.Require().Equal(false, results.ReadmeMarkdown)
	s.Require().Equal("", bReadme.String())
	s.Require().Equal(0, bReadme.Len())
}

func (s *ArchiveSuite) TestReadmeOnlyResolution() {
	f, err := os.Open("../testdata/readmetest_0.2.0.tar.gz")
	a := &RPackageArchive{}
	s.Require().Nil(err)

	var bReadme bytes.Buffer
	result, err := a.GetReadme(f, &bReadme)
	s.Require().Nil(err)

	// Make sure we parsed the correct README file.
	s.Require().Equal(true, result)
	s.Require().Equal("Hi, I'm the correct readme!", bReadme.String())
}

func (s *ArchiveSuite) TestReadmeOnlyResolutionNone() {
	f, err := os.Open("../testdata/hasdependencies_0.2.0.tar.gz")
	a := &RPackageArchive{}
	s.Require().Nil(err)

	var bReadme bytes.Buffer
	result, err := a.GetReadme(f, &bReadme)
	s.Require().Nil(err)

	// Make sure we didn't find a README
	s.Require().Equal(false, result)
	s.Require().Equal("", bReadme.String())
	s.Require().Equal(0, bReadme.Len())
}

func (s *ArchiveSuite) TestDescriptionFFEncoding() {
	a := NewRPackageArchive(256, 6)
	f, err := os.Open("../testdata/ff_2.2-14.tar.gz")
	s.Require().Nil(err)

	var b bytes.Buffer
	var bReadme bytes.Buffer
	results, err := a.RewriteWithReadme(f, &b, &bReadme)
	s.Require().Nil(err)

	s.Require().Equal(true, strings.Contains(results.Description, "latin1"))
	test.TestifyGolden(results.Description, &s.Suite)

}

func (s *ArchiveSuite) TestDescriptionChecksums() {
	// Calculate the checksums of all the DESCRIPTION files to
	// ensure we are not mistakenly rewriting any of them. DT
	// provides a good example as it has many files and was at
	// one point being rewritten incorrectly.
	descFiles := map[string]string{
		"DT/DESCRIPTION":                            "502dfda6fe48c0dfc2e2328028ae719c",
		"DT/inst/examples/DT-click/DESCRIPTION":     "71abd2a48dfda62caf5aefd1cfbc7010",
		"DT/inst/examples/DT-edit/DESCRIPTION":      "16204ab837655d6e26b37738864d3a44",
		"DT/inst/examples/DT-filter/DESCRIPTION":    "bc3b531b3c774fd638f172ecbdf9187f",
		"DT/inst/examples/DT-info/DESCRIPTION":      "c9710299bd7eb8704d75f550c8681399",
		"DT/inst/examples/DT-proxy/DESCRIPTION":     "3c7d263df5dd1695190daac9b0e1ea97",
		"DT/inst/examples/DT-radio/DESCRIPTION":     "cfe4df1fb43445e2abf0a1ca36a68649",
		"DT/inst/examples/DT-rows/DESCRIPTION":      "6d372ae0656a985bdf1f5b5bd9c660d3",
		"DT/inst/examples/DT-scroller/DESCRIPTION":  "eedd9f92f16586ea8a86a0d0c6c9283b",
		"DT/inst/examples/DT-selection/DESCRIPTION": "b45bee37de59858ac17cb854ae8b4b93",
	}

	var b bytes.Buffer
	var bReadme bytes.Buffer
	a := NewRPackageArchive(256, 6)
	f, err := os.Open("../testdata/DT_0.4.tar.gz")
	s.Require().Nil(err)
	results, err := a.RewriteWithReadme(f, &b, &bReadme)
	s.Require().Nil(err)
	s.Require().Equal("3daa96b819ca54e5fbc2c7d78cb3637982a2d44be58cea0683663b71cfc7fa19", results.OriginalChecksum)
	s.Require().Equal("f320e15acff2a42ef6db261ff71d747245d840d00adf340a24db5e92ed0cf6ca", results.RewrittenChecksum)

	gzf, err := gzip.NewReader(&b)
	s.Require().Nil(err)
	tarReader := tar.NewReader(gzf)

	filesValidated := 0
	for {
		var header *tar.Header
		header, err = tarReader.Next()
		if err == io.EOF {
			break
		}
		s.Require().Nil(err)
		if expectedSum, ok := descFiles[header.Name]; ok {
			var m bytes.Buffer
			_, _ = io.Copy(&m, tarReader)
			actualSum := fmt.Sprintf("%x", md5.Sum(m.Bytes()))
			s.Require().Equal(expectedSum, actualSum)
			filesValidated++
		}
	}

	s.Require().Equal(10, filesValidated)
}

func (s *ArchiveSuite) TestReadmePref() {
	s.Require().Equal(false, PreferredReadme("", ""))
	s.Require().Equal(false, PreferredReadme("", "README"))
	s.Require().Equal(false, PreferredReadme("README", ""))
	s.Require().Equal(false, PreferredReadme("bad", "README"))
	s.Require().Equal(false, PreferredReadme("README", "bad"))
	s.Require().Equal(false, PreferredReadme("README.MD", "README.txt"))
	s.Require().Equal(false, PreferredReadme("README.MD", "README"))
	s.Require().Equal(false, PreferredReadme("README.MD", "README.md"))
	s.Require().Equal(true, PreferredReadme("README.txt", "README.md"))
	s.Require().Equal(false, PreferredReadme("README.txt", "README"))
	s.Require().Equal(false, PreferredReadme("README.txt", "README.txt"))
	s.Require().Equal(true, PreferredReadme("README", "readme.txt"))
	s.Require().Equal(true, PreferredReadme("README", "readme.md"))
	s.Require().Equal(false, PreferredReadme("README", "readme"))
}

func (s *ArchiveSuite) TestBadChecksum() {
	a := NewRPackageArchive(256, 6)
	f, err := os.Open("../testdata/MortCast_2.6-1.tar.gz")
	s.Require().Nil(err)

	var b bytes.Buffer
	var bReadme bytes.Buffer
	results, err := a.RewriteWithReadme(f, &b, &bReadme)
	s.Require().Nil(err)

	// Make sure we parsed the correct DESCRIPTION file.
	s.Require().Equal("2df23b0744c0ea5f4708f06d62e46db71c43819276eb2428e9d1293e6c69ee70", results.OriginalChecksum)
	s.Require().Equal("f878588d6bd09340d9f0549568ead9a9b9fe0bcb4836173ec39a937ec46a59fa", results.RewrittenChecksum)
	s.Require().Equal(2891989, b.Len())

	// Check the contents of the DESCRIPTION and MD5 files.
	md5File, err := StreamFileFromTarGz(&b, "MD5")
	s.Require().Nil(err)
	var m bytes.Buffer
	_, _ = m.ReadFrom(md5File)
	_, err = io.Copy(io.Discard, md5File)
	s.Require().Nil(err)
	test.TestifyGolden(results.Description+"\n\n"+m.String(), &s.Suite)

}

var EntryNotFoundInTarBall = errors.New("requested entry was not found in the tarball")

// Exposes a reader for a particular file within a gzipped tar ball. The function will
// claim that a tar entry is a match if the name exactly matches the given path
// OR if the given path is nested under a single top-level directory. This is
// useful because some R packages nest their entries under a top-level directory
// (typically the package name) and others don't.
// Returns a stream associated with the requested file if it was found. Returns
// a EntryNotFoundInTarBall error if the requested file was not found.
//
// It is up to the caller to drain the stream upon completion.
func StreamFileFromTarGz(tarStream io.Reader, path string) (stream io.Reader, err error) {
	match := func(name string) bool {
		if !strings.HasSuffix(name, path) {
			return false
		}

		// We know the suffix matches, but we need to compute what's left of the path.
		remainder := strings.TrimSuffix(name, path)

		remainder = filepath.Clean(remainder)

		// Trim leading and trailing slashes
		remainder = strings.Trim(remainder, "/")

		return !strings.Contains(remainder, "/")

	}
	s, _, err := StreamFileFromTarGzCustom(tarStream, match)
	return s, err
}

func StreamFileFromTarGzCustom(tarStream io.Reader, match func(string) bool) (stream io.Reader, h *tar.Header, err error) {
	var gzf *gzip.Reader
	gzf, err = gzip.NewReader(tarStream)
	if err != nil {
		return
	}

	tarReader := tar.NewReader(gzf)

	for {
		var header *tar.Header
		header, err = tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return
		}

		// We want to match either exact matches or matches that are prefixed
		// by the package name in the tar ball.
		if header.Typeflag == tar.TypeReg && // is a plain file
			match(header.Name) {
			return tarReader, header, nil
		}
	}

	err = EntryNotFoundInTarBall
	return
}
