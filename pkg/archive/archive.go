// Copyright (C) 2023 by Posit Software, PBC
package archive

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html/charset"

	"github.com/rstudio/package-manager-rewriting/internal/utils"
)

const (
	DescriptionRepository = "Repository: RSPM"
	DescriptionEncoding   = "Encoding: %s"

	// Additional notes about CRAN DESCRIPTION file and encodings available here:
	// https://cran.r-project.org/doc/manuals/r-release/R-exts.html#The-DESCRIPTION-file
	// https://cran.r-project.org/doc/manuals/r-release/R-exts.html#Encoding
	defaultEncodingLatin1 = "latin1"
	encodingLatin2        = "latin2"
	encodingUTF8          = "UTF-8"
)

// RPackageArchive supports rewriting .tar.gz source packages and binaries in a single data stream while
// (a) calculating the source and destination SHA256 checksums, (b) calculating the source and
// destination file sizes, (c) rewriting the `DESCRIPTION` file to include the correct `Repository`
// field, and (d) rewriting the `MD5` file with any corrections required for the updated
// DESCRIPTION.
//
// IMPORTANT: If you make improvements to `RPackageArchive`, please also update `RPackageZipArchive` in
// `archive_zip.go`. We decided to maintain the code separately since there are some
// substantial differences in reading ZIP vs. TAR archives.
type RPackageArchive struct {
	bufferSize int
	gzipLevel  int
}

type Results struct {
	OriginalSize      int64
	RewrittenSize     int64
	OriginalChecksum  string
	RewrittenChecksum string
	Description       string
	ReadmeMarkdown    bool
}

type RewriteResults struct {
	Results
	RewrittenPath       string
	ExtractedReadmePath string
}

type LenWriter struct {
	len int64
}

func (w *LenWriter) Write(p []byte) (nn int, err error) {
	nn = len(p)
	w.len += int64(nn)
	return
}

// Support having a leading path segment (as some R packages have)
var readmeRE = regexp.MustCompile(`^(?i)([^/]+/)?README(\.(txt|md))?$`)

func (a *RPackageArchive) RewriteBinary(r io.Reader, w io.Writer) (results *Results, err error) {
	return a.rewrite(r, w, nil)
}

func (a *RPackageArchive) RewriteWithReadme(r io.Reader, w, wReadme io.Writer) (results *Results, err error) {
	return a.rewrite(r, w, wReadme)
}

func (a *RPackageArchive) rewrite(r io.Reader, w, wReadme io.Writer) (results *Results, err error) {

	// Gzip and tar to the destination
	//
	// `hw` calculates the SHA256 checksum for the rewritten package
	hw := sha256.New()
	// `lw` calculates the output size
	lw := &LenWriter{}
	// `hwlw` writes to both the SHA hash and the LenWriter simultaneously
	hwlw := io.MultiWriter(hw, lw)
	// `mw` writes to both `hwlw` and `w` simultaneously.
	mw := io.MultiWriter(hwlw, w)
	// write buffer that respects `Server.PackageRewriteBufferSize`; the buffer
	// writes to the `mw` multiwriter.
	buffer := bufio.NewWriterSize(mw, a.bufferSize)
	// `gzw` compresses data before sending it to the buffered writer. The compression
	// level is set by `Server.PackageRewriteCompressionLevel`
	gzw, err := gzip.NewWriterLevel(buffer, a.gzipLevel)
	if err != nil {
		return
	}
	tw := tar.NewWriter(gzw)
	defer func() {
		_ = tw.Close()
		_ = gzw.Close()
		_ = buffer.Flush()
		// These must be set after the buffers are flushed
		if results != nil {
			results.RewrittenChecksum = fmt.Sprintf("%x", hw.Sum(nil))
			results.RewrittenSize = lw.len
		}
	}()

	// Iterate over the files and:
	// - rewrite the Repository field in the DESCRIPTION file
	// - read the best matching README file
	// - update the MD5 file.
	// Since we cannot guarantee the order of the tar files, we need to
	// capture the two sections and write them at the end.

	// Buffers information about DESCRIPTION files we find while reading
	type descriptionInfo struct {
		buffer *bytes.Buffer
		header *tar.Header
	}
	descriptions := make([]descriptionInfo, 0)

	// Buffers information about MD5 files we find while reading
	type md5Info struct {
		buffer *bytes.Buffer
		header *tar.Header
	}
	md5s := make([]md5Info, 0)

	// For logging
	writeTime := int64(0)

	// Holds the contents of the best-matching README file
	readmeBuffer := bytes.NewBuffer([]byte{})

	// Tee the reads so we can calculate the original checksum while
	// rewriting the archive.
	rFileStream, wFileStream := io.Pipe()
	defer func(rFileStream *io.PipeReader) {
		_ = rFileStream.Close()
	}(rFileStream)
	defer func(wFileStream *io.PipeWriter) {
		_ = wFileStream.Close()
	}(wFileStream)
	rHashStream := utils.NewEOFTeeReader(r, wFileStream)
	type checkResult struct {
		checksum string
		err      error
		size     int64
	}
	chanCheckResult := make(chan checkResult)
	go func() {
		// Calculate checksum
		origSize, sum, errSha := utils.ComputeSha256Stream(rHashStream)
		if errSha != nil {
			chanCheckResult <- checkResult{"", errSha, 0}
		}
		chanCheckResult <- checkResult{hex.EncodeToString(sum), nil, origSize}
	}()

	// Create the gzip and tar readers
	gr, err := gzip.NewReader(rFileStream)
	if err != nil {
		return
	}
	defer func(gr *gzip.Reader) {
		_ = gr.Close()
	}(gr)
	tr := tar.NewReader(gr)

	// descPathLen is used to ensure that we are parsing the correct
	// DESCRIPTION file in the tar archive. Since there could be multiple,
	// we look for the one with the shortest file path. This avoids using
	// a naming convention like "[package name]/DESCRIPTION", or a regex.
	// Both of which could be brittle.
	descPathLen := 0
	descPath := ""
	// We do the same thing for README files, and we also record whether
	// we found a text or markdown README file.
	readmePathLen := 0
	readmeName := ""
	readmeMarkdown := false
	// Finally, we record the shortest-path MD5 file.
	md5PathLen := 0
	md5Path := ""

	for {
		var header *tar.Header
		header, err = tr.Next()

		if err == io.EOF {
			// Reset err; don't err on EOF
			err = nil
			break
		} else if err != nil {
			return
		}

		name := header.FileInfo().Name()

		// Only buffer the DESCRIPTION file if we have not found a file with
		// that name yet, or if we find one with a shorter path than one we
		// found earlier. This way we do not care about tar file ordering.
		if name == "DESCRIPTION" && (descPathLen == 0 || len(header.Name) < descPathLen) {
			// For the read pass we want to capture only the base DESCRIPTION
			// file.
			// Save the header
			descInfo := descriptionInfo{
				buffer: bytes.NewBuffer([]byte{}),
				header: &(*header),
			}
			descPathLen = len(header.Name)
			descPath = header.Name

			_, err = io.Copy(descInfo.buffer, tr)
			if err != nil {
				err = fmt.Errorf("error copying description: %s", err)
				return
			}

			// Append to the list of buffered DESCRIPTION files
			descriptions = append(descriptions, descInfo)

		} else if wReadme != nil && readmeRE.MatchString(header.Name) {
			// Only buffer the README file if
			// (a) we have not found a file that matches the regex yet,
			// (b) if we find one with a shorter path than one we found earlier, or
			// (c) if the path length is the same but a PreferredReadme name is found.
			// This way we do not care about tar file ordering.
			_ = tw.WriteHeader(header)
			newReadmePathLen := len(filepath.Dir(name))
			if readmePathLen == 0 || (newReadmePathLen < readmePathLen) || (readmePathLen == newReadmePathLen && PreferredReadme(readmeName, name)) {
				readmeName = name
				readmeMarkdown = strings.ToLower(readmeName) == "readme.md"
				readmePathLen = newReadmePathLen

				// Reset the buffer in case we had a longer-path match first.
				readmeBuffer.Reset()

				// Write to both the readme buffer and the tar writer. This ensures that the
				// readmeBuffer always contains the best-matching README that we've found so
				// far.
				readmeMultiWriter := io.MultiWriter(readmeBuffer, tw)
				if _, err = io.Copy(readmeMultiWriter, tr); err != nil {
					err = fmt.Errorf("error writing to readme multiwriter: %s", err)
					return
				}
			} else {
				// If this isn't a new best-matching README, simply write it out to
				// the TAR writer.
				if _, err = io.Copy(tw, tr); err != nil {
					err = fmt.Errorf("error writing other readme: %s", err)
					return
				}
			}

		} else if name == "MD5" && (md5PathLen == 0 || len(header.Name) < md5PathLen) {
			// Only buffer the MD5 file if we have not found a file with
			// that name yet, or if we find one with a shorter path than one we
			// found earlier. This way we do not care about tar file ordering.
			md5PathLen = len(header.Name)
			md5Path = header.Name

			info := md5Info{
				buffer: bytes.NewBuffer([]byte{}),
				header: &(*header),
			}

			_, err = io.Copy(info.buffer, tr)
			if err != nil {
				return
			}

			// Append to the list of buffered MD5 files.
			md5s = append(md5s, info)

		} else {
			// We'll hit this block writing any data to the tarball where
			// we don't have a special handler above.
			start := time.Now()

			// Here, write the header and content as is.
			if err = tw.WriteHeader(header); err != nil {
				return
			}
			if _, err = io.Copy(tw, tr); err != nil {
				return
			}

			writeTime += time.Since(start).Milliseconds()
		}
	}

	// Write the buffered DESCRIPTION files at the end
	var (
		descMd5         string
		descriptionText string
		toUTF           bool
	)

	for _, descInfo := range descriptions {
		header := descInfo.header

		// If this is the authoritative DESCRIPTION, rewrite it as needed and save it to
		// a string for easy access later, then write the new DESCRIPTION contents
		// to the TAR writer.
		if header.Name == descPath {
			descReadBuffer := bytes.NewBuffer([]byte{})
			useEncoding := defaultEncodingLatin1

			// Rewrite the DESCRIPTION file.
			scanner := bufio.NewScanner(descInfo.buffer)
			repoFieldFound := false
			encodingFieldFound := false

			for scanner.Scan() {
				line := scanner.Bytes()
				if bytes.HasPrefix(line, []byte("Repository: ")) {
					repoFieldFound = true
					line = []byte(DescriptionRepository)
				} else if bytes.HasPrefix(line, []byte("Encoding: ")) {
					encodingFieldFound = true
					lineLower := strings.ToLower(string(line))

					// Set the `useEncoding` and DESCRIPTION encoding line correctly
					switch {
					case strings.Contains(lineLower, defaultEncodingLatin1):
						useEncoding = defaultEncodingLatin1
					case strings.Contains(lineLower, encodingLatin2):
						useEncoding = encodingLatin2
					default:
						toUTF = true
						useEncoding = encodingUTF8
					}

					line = []byte(fmt.Sprintf(DescriptionEncoding, useEncoding))
				}

				descReadBuffer.Write(line)
				descReadBuffer.Write([]byte("\n"))
			}

			// In rare cases the Repository field is not set.
			if !repoFieldFound {
				descReadBuffer.Write([]byte(DescriptionRepository + "\n"))
			}
			// Add an Encoding field if none was found
			if !encodingFieldFound {
				toUTF = true
				descReadBuffer.Write([]byte(fmt.Sprintf(DescriptionEncoding, encodingUTF8) + "\n"))
			}

			// Get a reader that understand the DESCRIPTION encoding
			var reader io.Reader
			reader = descReadBuffer

			if toUTF {
				reader, err = charset.NewReaderLabel(useEncoding, descReadBuffer)
				if err != nil {
					return
				}
			}

			// Copy data using UTF-8.
			descInfo.buffer.Reset()
			_, err = io.Copy(descInfo.buffer, reader)
			if err != nil {
				return
			}

			// Update the header's size value
			header.Size = int64(descInfo.buffer.Len())

			// Calculate the MD5
			descMd5 = fmt.Sprintf("%x", md5.Sum(descInfo.buffer.Bytes()))

			// Save the description
			descriptionText = descInfo.buffer.String()
		}

		// Write the DESCRIPTION
		if err = tw.WriteHeader(header); err != nil {
			return
		}
		// Write the contents to the TAR writer. For the authoritative DESCRIPTION
		// file, the buffer contains the rewritten contents. For all other
		// buffered DESCRIPTION files, the buffer contains the original contents.
		if _, err = io.Copy(tw, descInfo.buffer); err != nil {
			return
		}
	}

	// Write the readme file out to the writer. This extracts the README for
	// faster access later.
	if wReadme != nil && readmeBuffer.Len() > 0 {
		if _, err = io.Copy(wReadme, readmeBuffer); err != nil {
			return
		}
	}

	// MD5 handling
	// Rewrite the MD5 file.
	for _, info := range md5s {
		header := info.header

		// If this is the authoritative MD5 file, we must rewrite it.
		if header.Name == md5Path {
			md5Buffer := bytes.NewBuffer([]byte{})
			scanner := bufio.NewScanner(info.buffer)
			for scanner.Scan() {
				line := scanner.Bytes()
				if bytes.HasSuffix(line, []byte(" *DESCRIPTION")) {
					if descMd5 == "" {
						fmt.Printf("Error: No DESCRIPTION checksum was generated.")
					} else {
						line = []byte(descMd5 + " *DESCRIPTION")
					}
				}
				md5Buffer.Write(line)
				md5Buffer.Write([]byte("\n"))
			}
			info.buffer.Reset()
			_, err = io.Copy(info.buffer, md5Buffer)
			if err != nil {
				return
			}
			header.Size = int64(info.buffer.Len())
		}
		if err = tw.WriteHeader(header); err != nil {
			return
		}
		// Write the contents to the TAR writer. For the authoritative MD5
		// file, the buffer contains the rewritten contents. For all other
		// buffered MD5 files, the buffer contains the original contents.
		if _, err = io.Copy(tw, info.buffer); err != nil {
			return
		}
	}

	// At this point we're done with `rFileStream`, since `rHashStream` waits to read the same buffer at the same time,
	// we need to copy the remaining bytes to /dev/null to ensure `rHashStream` reaches the end of the stream.
	_, _ = io.Copy(io.Discard, rFileStream)

	// Wait for the original checksum calculation to complete
	originalChecksum := <-chanCheckResult
	if originalChecksum.err != nil {
		err = originalChecksum.err
		return
	}

	// Calculate result to return. Note that there is a `defer` near the
	// top of this function that mutates the returned results further by
	// setting the RewrittenChecksum and RewrittenSize properties.
	results = &Results{
		OriginalSize:     originalChecksum.size,
		OriginalChecksum: originalChecksum.checksum,
		Description:      descriptionText,
		ReadmeMarkdown:   readmeMarkdown,
	}

	return
}

func (a *RPackageArchive) GetReadme(stream io.Reader, wReadme io.Writer) (bool, error) {

	readmeBuffer := bytes.NewBuffer([]byte{})

	// Create the gzip and tar readers
	gr, err := gzip.NewReader(stream)
	if err != nil {
		return false, err
	}
	tr := tar.NewReader(gr)

	readmePathLen := 0
	readmeName := ""
	var readmeMarkdown bool

	for {
		var header *tar.Header
		header, err = tr.Next()

		if err == io.EOF {
			break
		} else if err != nil {
			return false, err
		}

		name := header.FileInfo().Name()

		if readmeRE.MatchString(header.Name) {

			// Only read the README file if
			// (a) we have not found a file that matches the regex yet,
			// (b) if we find one with a shorter path than one we found earlier, or
			// (c) if the path length is the same but a PreferredReadme name is found.
			// This way we do not care about tar file ordering.
			newReadmePathLen := len(filepath.Dir(name))
			if readmePathLen == 0 || (newReadmePathLen < readmePathLen) || (readmePathLen == newReadmePathLen && PreferredReadme(readmeName, name)) {
				readmeName = name
				readmeMarkdown = strings.ToLower(readmeName) == "readme.md"
				readmePathLen = newReadmePathLen

				// Reset the buffer in case we had a longer-path match first.
				readmeBuffer.Reset()

				if _, err = io.Copy(readmeBuffer, tr); err != nil {
					return false, err
				}
			}
		}
	}

	// Write the readme, if any, to the writer
	if _, err = io.Copy(wReadme, readmeBuffer); err != nil {
		return false, err
	}

	return readmeMarkdown, nil
}

// A map that enumerates the preference of README names, with
// a lower int value representing a "more preferred" file name
var readmeMap = map[string]int{
	"readme.md":  0,
	"readme.txt": 1,
	"readme":     2,
}

func PreferredReadme(oldName, newName string) bool {
	a, ok := readmeMap[strings.ToLower(oldName)]
	if !ok {
		return false
	}
	b, ok := readmeMap[strings.ToLower(newName)]
	if !ok {
		return false
	}
	return b < a
}

func NewRPackageArchive(bufferSize, gzipLevel int) *RPackageArchive {
	return &RPackageArchive{
		bufferSize: bufferSize,
		gzipLevel:  gzipLevel,
	}
}
