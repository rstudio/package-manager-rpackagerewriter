// Copyright (C) 2023 by Posit Software, PBC
package rewriter

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/rstudio/package-manager-rewriting/internal/utils"
	"github.com/rstudio/package-manager-rewriting/pkg/archive"
	fpg "github.com/rstudio/package-manager-rewriting/pkg/utils"
)

func NewRewriteError(err error) RewriteError {
	return RewriteError{error: err}
}

type RewriteError struct {
	error
}

func (r RewriteError) Error() string {
	return r.error.Error()
}

func (r RewriteError) Is(err error) bool {
	_, ok := err.(RewriteError)
	return ok
}

type ArchiveRewriter interface {
	Rewrite(fullPath string) (*archive.RewriteResults, error)
	RewriteStream(r io.Reader, w io.Writer) (*archive.RewriteResults, error)
	RewriteBinary(r *os.File, w io.Writer, zip bool) (*archive.RewriteResults, error)
	GetReadme(stream io.Reader) (*archive.RewriteResults, error)
}

type archiveRewriter struct {
	OutputDir       string
	ReadmeOutputDir string
	tempDir         string
	fpg             fpg.FilePathGetter
	bufferSize      int
	gzipLevel       int
}

func NewArchiveRewriter(outputDir, readmeOutputDir, tempDir string, fpg fpg.FilePathGetter, bufferSize, gzipLevel int) *archiveRewriter {
	return &archiveRewriter{
		OutputDir:       outputDir,
		ReadmeOutputDir: readmeOutputDir,
		tempDir:         tempDir,
		fpg:             fpg,
		bufferSize:      bufferSize,
		gzipLevel:       gzipLevel,
	}
}

func (r *archiveRewriter) Rewrite(fullPath string) (*archive.RewriteResults, error) {
	f, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("error: could not open %s: %s", fullPath, err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	w, err := ioutil.TempFile(r.OutputDir, "")
	if err != nil {
		return nil, fmt.Errorf("error: could not create temp file for %s. %s", fullPath, err)
	}
	tempFileName := w.Name()
	defer func(err *error) {
		if *err != nil {
			_ = os.Remove(tempFileName)
		}
	}(&err)

	wReadme, err := ioutil.TempFile(r.ReadmeOutputDir, "")
	if err != nil {
		return nil, fmt.Errorf("error: could not create readme temp file for %s. %s", fullPath, err)
	}
	tempFileNameReadme := wReadme.Name()
	defer func(err *error) {
		if *err != nil {
			_ = os.Remove(tempFileNameReadme)
		}
	}(&err)

	// Rewrite the file and save using the checksum as the filename.
	arch := archive.NewArchive(r.bufferSize, r.gzipLevel)
	var aResults *archive.ArchiveResults
	if aResults, err = arch.RewriteWithReadme(f, w, wReadme); err != nil {
		return nil, fmt.Errorf("error rewriting %s: %s", w.Name(), err)
	}

	readmeStat, err := wReadme.Stat()
	if err != nil {
		return nil, fmt.Errorf("error getting readme stats on %s: %s", wReadme.Name(), err)
	}
	_ = wReadme.Close()

	// Special case where we recreate a manifest by recording original checksums in a special file.
	if here, _ := utils.FileExists(fullPath + ".original.checksum"); here {
		var bts []byte
		var origFile *os.File
		origFile, err = os.Open(fullPath + ".original.checksum")
		if err != nil {
			return nil, err
		}
		defer origFile.Close()
		bts, err = io.ReadAll(origFile)
		if err != nil {
			return nil, err
		}
		aResults.OriginalChecksum = strings.TrimSpace(string(bts))
	}

	// Move the temp file
	checksumFilePath := r.fpg.GetFilePath(r.OutputDir, aResults)
	err = os.Rename(tempFileName, checksumFilePath)
	if err != nil {
		return nil, fmt.Errorf("error moving file %s to %s: %s", tempFileName, checksumFilePath, err)
	}

	// Move the temp README file
	checksumFilePathReadme := ""
	if readmeStat.Size() > 0 {
		checksumFilePathReadme = r.fpg.GetReadmePath(r.ReadmeOutputDir, aResults)
		err = os.Rename(tempFileNameReadme, checksumFilePathReadme)
		if err != nil {
			return nil, fmt.Errorf("error moving readme file %s to %s: %s", tempFileNameReadme, checksumFilePathReadme, err)
		}
	} else {
		err = os.Remove(tempFileNameReadme)
		if err != nil {
			return nil, fmt.Errorf("error removing empty temp readme file %s: %s", tempFileNameReadme, err)
		}
	}

	return &archive.RewriteResults{
		ArchiveResults:      *aResults,
		RewrittenPath:       checksumFilePath,
		ExtractedReadmePath: checksumFilePathReadme,
	}, nil
}

func (r *archiveRewriter) RewriteStream(reader io.Reader, w io.Writer) (*archive.RewriteResults, error) {
	wReadme, err := ioutil.TempFile(r.ReadmeOutputDir, "")
	if err != nil {
		return nil, fmt.Errorf("error: could not create readme temp file for stream. %s", err)
	}
	tempFileNameReadme := wReadme.Name()
	defer func(err *error) {
		if *err != nil {
			_ = os.Remove(tempFileNameReadme)
		}
	}(&err)

	// Rewrite the file and save using the checksum as the filename.
	arch := archive.NewArchive(r.bufferSize, r.gzipLevel)
	var aResults *archive.ArchiveResults
	if aResults, err = arch.RewriteWithReadme(reader, w, wReadme); err != nil {
		return nil, fmt.Errorf("error rewriting stream: %s", err)
	}

	readmeStat, err := wReadme.Stat()
	if err != nil {
		return nil, fmt.Errorf("error getting readme stats on %s: %s", wReadme.Name(), err)
	}
	_ = wReadme.Close()

	// Move the temp README file
	checksumFilePathReadme := ""
	if readmeStat.Size() > 0 {
		checksumFilePathReadme = r.fpg.GetReadmePath(r.ReadmeOutputDir, aResults)
		err = os.Rename(tempFileNameReadme, checksumFilePathReadme)
		if err != nil {
			return nil, fmt.Errorf("error moving readme file %s to %s: %s", tempFileNameReadme, checksumFilePathReadme, err)
		}
	} else {
		err = os.Remove(tempFileNameReadme)
		if err != nil {
			return nil, fmt.Errorf("error removing empty temp readme file %s: %s", tempFileNameReadme, err)
		}
	}

	return &archive.RewriteResults{
		ArchiveResults:      *aResults,
		ExtractedReadmePath: checksumFilePathReadme,
	}, nil
}

func (r *archiveRewriter) RewriteBinary(file *os.File, w io.Writer, zip bool) (*archive.RewriteResults, error) {
	var err error
	var aResults *archive.ArchiveResults
	if zip {
		arch := archive.NewArchiveZip(r.bufferSize)
		aResults, err = arch.RewriteBinary(file, w)
	} else {
		arch := archive.NewArchive(r.bufferSize, r.gzipLevel)
		aResults, err = arch.RewriteBinary(file, w)
	}

	if err != nil {
		return nil, fmt.Errorf("error rewriting stream: %w", RewriteError{error: err})
	}

	// Require a DESCRIPTION
	if aResults.Description == "" {
		err = fmt.Errorf("no DESCRIPTION file found in archive")
		return nil, fmt.Errorf("error rewriting stream: %w", RewriteError{error: err})
	}

	return &archive.RewriteResults{
		ArchiveResults: *aResults,
	}, nil
}

func (r *archiveRewriter) GetReadme(stream io.Reader) (*archive.RewriteResults, error) {
	arc := &archive.Archive{}
	results := &archive.RewriteResults{}

	wReadme, err := ioutil.TempFile(r.ReadmeOutputDir, "")
	if err != nil {
		return nil, fmt.Errorf("error: could not create readme temp file: %s", err)
	}
	tempFileNameReadme := wReadme.Name()
	defer func(err *error) {
		if *err != nil {
			_ = os.Remove(tempFileNameReadme)
		}
	}(&err)

	// Rewrite the file and save using the checksum as the filename.
	var markdown bool
	if markdown, err = arc.GetReadme(stream, wReadme); err != nil {
		return nil, fmt.Errorf("error getting readme %s: %s", wReadme.Name(), err)
	}

	readmeStat, err := wReadme.Stat()
	if err != nil {
		return nil, fmt.Errorf("error getting readme stats on %s: %s", wReadme.Name(), err)
	}
	_ = wReadme.Close()

	// Move the temp README file
	checksumFilePathReadme := ""
	if readmeStat.Size() > 0 {
		checksumFilePathReadme = r.fpg.GetReadmePath(r.ReadmeOutputDir, &archive.ArchiveResults{ReadmeMarkdown: markdown})
		err = os.Rename(tempFileNameReadme, checksumFilePathReadme)
		if err != nil {
			return nil, fmt.Errorf("error moving readme file %s to %s: %s", tempFileNameReadme, checksumFilePathReadme, err)
		}
	} else {
		err = os.Remove(tempFileNameReadme)
		if err != nil {
			return nil, fmt.Errorf("error removing empty temp readme file %s: %s", tempFileNameReadme, err)
		}
	}

	// Track the changes in the checkpoint JSON.
	results.ExtractedReadmePath = checksumFilePathReadme

	return results, nil
}
