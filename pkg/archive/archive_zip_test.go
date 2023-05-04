// Copyright (C) 2023 by Posit Software, PBC
package archive

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/rstudio/package-manager-rewriting/internal/test"
)

func TestArchiveZipSuite(t *testing.T) {
	suite.Run(t, &ArchiveZipSuite{})
}

type ArchiveZipSuite struct {
	suite.Suite
}

func (s *ArchiveZipSuite) TestNewArchive() {
	a := NewArchiveZip(256)
	s.Require().Equal(&ZipArchive{
		bufferSize: 256,
	}, a)
}

func (s *ArchiveZipSuite) TestDescriptionRewriteBinaryInvalid() {
	a := NewArchiveZip(256)

	tmp, err := os.CreateTemp("", "")
	s.Require().Nil(err)
	defer func(tmp *os.File) {
		_ = tmp.Close()
	}(tmp)

	_, err = tmp.WriteString("this is a test file\nwith no zip content\n.")
	s.Require().Nil(err)
	_, err = tmp.Seek(0, 0)
	s.Require().Nil(err)

	var b bytes.Buffer
	_, err = a.RewriteBinary(tmp, &b)
	s.Require().ErrorContains(err, "error opening ZIP reader in ZipArchive.RewriteBinary: zip: not a valid zip file")
}

func (s *ArchiveZipSuite) TestDescriptionRewriteBinary() {
	a := NewArchiveZip(256)
	f, err := os.Open("../testdata/binaries/bindrcpp_0.2.2.zip")
	s.Require().Nil(err)

	// Create a temporary output file
	out, err := os.CreateTemp("", "")
	s.Require().Nil(err)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
		}
	}(out.Name())

	// Rewrite the binary to the temporary file
	results, err := a.RewriteBinary(f, out)
	s.Require().Nil(err)
	_ = out.Close()

	// Stat the file we created
	stat, err := os.Stat(out.Name())
	s.Require().Nil(err)

	// Make sure we parsed the correct DESCRIPTION file.
	s.Require().Equal("dc4387dcd7a5ba5f778f2139121bc81dea5a44a0c2adb19a0c09dbff17e1247a", results.OriginalChecksum)
	s.Require().Equal("2b54feb73af798bda39e0c68fd13c58279147df77d870b502195d3a375c0e477", results.RewrittenChecksum)
	s.Require().Equal(int64(412918), results.OriginalSize)
	s.Require().Equal(int64(431367), results.RewrittenSize)
	s.Require().Equal(int64(431367), stat.Size())

	// Calculate rewritten checksum manually to double-check it.
	fCheck, err := os.Open(out.Name())
	s.Require().Nil(err)
	hw := sha256.New()
	_, err = io.Copy(hw, fCheck)
	s.Require().Nil(err)
	expectedSHA := fmt.Sprintf("%x", hw.Sum(nil))
	s.Require().Equal(expectedSHA, results.RewrittenChecksum)
	_ = fCheck.Close()

	// Check the contents of the DESCRIPTION and MD5 files.
	unzip, err := zip.OpenReader(out.Name())
	s.Require().Nil(err)
	defer func(unzip *zip.ReadCloser) {
		err := unzip.Close()
		if err != nil {
		}
	}(unzip)
	bufDesc := &bytes.Buffer{}
	bufMD5 := &bytes.Buffer{}
	for _, f := range unzip.File {
		if f.Name == "bindrcpp/DESCRIPTION" {
			entry, err := unzip.Open(f.Name)
			s.Require().Nil(err)
			_, err = io.Copy(bufDesc, entry)
			s.Require().Nil(err)
		} else if f.Name == "bindrcpp/MD5" {
			entry, err := unzip.Open(f.Name)
			s.Require().Nil(err)
			_, err = io.Copy(bufMD5, entry)
			s.Require().Nil(err)
		}
	}

	// Result DESCRIPTION should match what was written
	s.Require().Equal(bufDesc.String(), results.Description)
	// DESCRIPTION and MD5 check against golden file
	test.TestifyGolden(results.Description+"\n\n"+bufMD5.String(), &s.Suite)
}

// Tests rewriting a package where the DESCRIPTION uses the `latin1` character
// encoding instead of UTF-8. We expect that
// * The DESCRIPTION is needs to retain the `Encoding: latin1`
// * The DESCRIPTION is written using the `latin1` encoding instead of the default `UTF-8`
func (s *ArchiveZipSuite) TestDescriptionRewriteLatin1() {
	a := NewArchiveZip(256)
	f, err := os.Open("../testdata/special/Latin1SpecialChars_1.1.1.zip")
	s.Require().Nil(err)

	// Create a temporary output file
	out, err := os.CreateTemp("", "")
	s.Require().Nil(err)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
		}
	}(out.Name())

	results, err := a.RewriteBinary(f, out)
	s.Require().Nil(err)
	_ = out.Close()

	// Stat the file we created
	stat, err := os.Stat(out.Name())
	s.Require().Nil(err)

	// Make sure we parsed the correct DESCRIPTION file.
	s.Require().Equal("68166db1f6f8cae91d84173c38e1d78797f474c0fbc0b042370fbe14b25435bb", results.OriginalChecksum)
	s.Require().Equal("a61033dd5a8ce88dd1b30f5bbc1bcde01decfd63b3b8065ab51ff071d499a070", results.RewrittenChecksum)
	s.Require().Equal(int64(1630), stat.Size())

	// Check the contents of the DESCRIPTION and MD5 files.
	unzip, err := zip.OpenReader(out.Name())
	s.Require().Nil(err)
	defer func(unzip *zip.ReadCloser) {
		err := unzip.Close()
		if err != nil {
		}
	}(unzip)
	bufDesc := &bytes.Buffer{}
	bufMD5 := &bytes.Buffer{}
	for _, f := range unzip.File {
		if f.Name == "Latin1SpecialChars/DESCRIPTION" {
			entry, err := unzip.Open(f.Name)
			s.Require().Nil(err)
			_, err = io.Copy(bufDesc, entry)
			s.Require().Nil(err)
		} else if f.Name == "Latin1SpecialChars/MD5" {
			entry, err := unzip.Open(f.Name)
			s.Require().Nil(err)
			_, err = io.Copy(bufMD5, entry)
			s.Require().Nil(err)
		}
	}

	// Check the contents of the DESCRIPTION and MD5 files.
	// Result DESCRIPTION should match what was written
	s.Require().Equal(bufDesc.String(), results.Description)
	// DESCRIPTION and MD5 check against golden file
	test.TestifyGolden(results.Description+"\n\n"+bufMD5.String(), &s.Suite)
}

func (s *ArchiveZipSuite) TestDescriptionFFEncoding() {
	a := NewArchiveZip(256)
	f, err := os.Open("../testdata/ff_2.2-14.zip")
	s.Require().Nil(err)

	// Create a temporary output file
	out, err := os.CreateTemp("", "")
	s.Require().Nil(err)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
		}
	}(out.Name())

	results, err := a.RewriteBinary(f, out)
	s.Require().Nil(err)
	_ = out.Close()

	s.Require().Equal(true, strings.Contains(results.Description, "latin1"))
	test.TestifyGolden(results.Description, &s.Suite)
}
