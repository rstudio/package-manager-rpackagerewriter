// Copyright (C) 2023 by Posit Software, PBC
package archive

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html/charset"
)

// ZipArchive is very similar to Archive (see `archive.go`). However, ZIP reading requires random
// access (io.ReaderAt), so we had to create a separate util. Instead of passing a stream as `r`,
// instead save the ZIP archive to a temporary file and then pass an `*os.File` as `r`. The file
// will be scanned twice: once to calculate the original checksum and size, and once to rewrite
// to the destination.
//
// IMPORTANT: If you make improvements to `ZipArchive`, please also update `Archive` in
// `archive.go`.
type ZipArchive struct {
	bufferSize int
}

func (a *ZipArchive) RewriteBinary(r *os.File, w io.Writer) (results *Results, err error) {

	// Calculate original checksum and size
	var szOrig int64
	var shaOrig string
	hr := sha256.New()
	szOrig, err = io.Copy(hr, r)
	if err != nil {
		err = fmt.Errorf("error copying when calculating SHA in ZipArchive.RewriteBinary: %s", err)
		return
	}
	shaOrig = fmt.Sprintf("%x", hr.Sum(nil))

	// Seek back to the beginning of the file
	_, err = r.Seek(0, 0)
	if err != nil {
		err = fmt.Errorf("error seeking to beginning of file in ZipArchive.RewriteBinary: %s", err)
		return
	}

	// Zip to the destination
	//
	// `hw` calculates the SHA256 checksum for the rewritten package
	hw := sha256.New()
	// `lw` calculates the output size
	lw := &LenWriter{}
	// `hwlw` writes to both the SHA hash and the LenWriter simultaneously
	hwlw := io.MultiWriter(hw, lw)
	mw := io.MultiWriter(hwlw, w)
	// write buffer that respects `Server.PackageRewriteBufferSize`; the buffer
	// writes to the `mw` multiwriter.
	buffer := bufio.NewWriterSize(mw, a.bufferSize)
	// `zipw` compresses data before sending it to the buffered writer.
	zipw := zip.NewWriter(buffer)
	defer func() {
		_ = zipw.Close()
		_ = buffer.Flush()
		// These must be set after the buffers are flushed
		if results != nil {
			results.RewrittenChecksum = fmt.Sprintf("%x", hw.Sum(nil))
			results.RewrittenSize = lw.len
		}
	}()

	// Iterate over the files and:
	// - rewrite the Repository field in the DESCRIPTION file
	// - update the MD5 file.
	// Since we cannot guarantee the order of the ZIP files, we need to
	// capture the two sections and write them at the end.

	// Buffers information about DESCRIPTION files we find while reading
	type descriptionInfo struct {
		buffer *bytes.Buffer
		header *zip.FileHeader
	}
	descriptions := make([]descriptionInfo, 0)

	// Buffers information about MD5 files we find while reading
	type md5Info struct {
		buffer *bytes.Buffer
		header *zip.FileHeader
	}
	md5s := make([]md5Info, 0)

	// For logging
	writeTime := int64(0)

	// Create the Zip reader
	stat, err := r.Stat()
	if err != nil {
		err = fmt.Errorf("error getting file Stat() in ZipArchive.RewriteBinary: %s", err)
		return
	}
	zr, err := zip.NewReader(r, stat.Size())
	if err != nil {
		err = fmt.Errorf("error opening ZIP reader in ZipArchive.RewriteBinary: %s", err)
		return
	}

	// descPathLen is used to ensure that we are parsing the correct
	// DESCRIPTION file in the ZIP archive. Since there could be multiple,
	// we look for the one with the shortest file path. This avoids using
	// a naming convention like "[package name]/DESCRIPTION", or a regex.
	// Both of which could be brittle.
	descPathLen := 0
	descPath := ""
	// Finally, we record the shortest-path MD5 file.
	md5PathLen := 0
	md5Path := ""

	for _, f := range zr.File {
		header := &f.FileHeader

		// Ignore directories
		if header.FileInfo().IsDir() {
			continue
		}

		name := header.FileInfo().Name()

		// Only buffer the DESCRIPTION file if we have not found a file with
		// that name yet, or if we find one with a shorter path than one we
		// found earlier. This way we do not care about ZIP file ordering.
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

			var zf fs.File
			zf, err = zr.Open(header.Name)
			if err != nil {
				err = fmt.Errorf("error opening ZIP archive file '%s' in ZipArchive.RewriteBinary: %s", header.Name, err)
				return
			}
			_, err = io.Copy(descInfo.buffer, zf)
			if err != nil {
				err = fmt.Errorf("error copying DESCRIPTION data for ZIP archive file '%s' in ZipArchive.RewriteBinary: %s", header.Name, err)
				return
			}

			// Append to the list of buffered DESCRIPTION files
			descriptions = append(descriptions, descInfo)

		} else if name == "MD5" && (md5PathLen == 0 || len(header.Name) < md5PathLen) {
			// Only buffer the MD5 file if we have not found a file with
			// that name yet, or if we find one with a shorter path than one we
			// found earlier. This way we do not care about ZIP file ordering.
			md5PathLen = len(header.Name)
			md5Path = header.Name

			info := md5Info{
				buffer: bytes.NewBuffer([]byte{}),
				header: &(*header),
			}

			var zf fs.File
			zf, err = zr.Open(header.Name)
			if err != nil {
				err = fmt.Errorf("error opening ZIP archive file '%s' in ZipArchive.RewriteBinary: %s", header.Name, err)
				return
			}

			_, err = io.Copy(info.buffer, zf)
			if err != nil {
				err = fmt.Errorf("error copying MD5 file data for ZIP archive file '%s' in ZipArchive.RewriteBinary: %s", header.Name, err)
				return
			}

			// Append to the list of buffered MD5 files.
			md5s = append(md5s, info)

		} else {
			// We'll hit this block writing any data to the ZIP file where
			// we don't have a special handler above.
			start := time.Now()

			// Here, write the header and content as is.
			var zf fs.File
			zf, err = zr.Open(header.Name)
			if err != nil {
				err = fmt.Errorf("error opening ZIP archive file '%s' in ZipArchive.RewriteBinary: %s", header.Name, err)
				return
			}

			var variousW io.Writer
			variousW, err = zipw.CreateHeader(header)
			if err != nil {
				err = fmt.Errorf("error creating ZIP header for file '%s' in ZipArchive.RewriteBinary: %s", header.Name, err)
				return
			}

			if _, err = io.Copy(variousW, zf); err != nil {
				err = fmt.Errorf("error copying data for file '%s' in ZipArchive.RewriteBinary: %s", header.Name, err)
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
		// to the ZIP writer.
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

			// Get a reader that understands the DESCRIPTION encoding
			var reader io.Reader
			reader = descReadBuffer

			if toUTF {
				reader, err = charset.NewReaderLabel(useEncoding, descReadBuffer)
				if err != nil {
					err = fmt.Errorf("error getting reader for encoding '%s' in ZipArchive.RewriteBinary: %s", useEncoding, err)
					return
				}
			}

			// Copy data using UTF-8.
			descInfo.buffer.Reset()
			_, err = io.Copy(descInfo.buffer, reader)
			if err != nil {
				err = fmt.Errorf("error copying DESCRIPTION data to buffer in ZipArchive.RewriteBinary: %s", err)
				return
			}

			// Update the header's size value
			header.UncompressedSize64 = uint64(descInfo.buffer.Len())

			// Calculate the MD5
			descMd5 = fmt.Sprintf("%x", md5.Sum(descInfo.buffer.Bytes()))

			// Save the description
			descriptionText = descInfo.buffer.String()
		}

		// Write the DESCRIPTION
		var descW io.Writer
		descW, err = zipw.CreateHeader(header)
		if err != nil {
			err = fmt.Errorf("error creating ZIP header for DESCRIPTION file '%s' in ZipArchive.RewriteBinary: %s", header.Name, err)
			return
		}
		// Write the contents to the ZIP writer. For the authoritative DESCRIPTION
		// file, the buffer contains the rewritten contents. For all other
		// buffered DESCRIPTION files, the buffer contains the original contents.
		if _, err = io.Copy(descW, descInfo.buffer); err != nil {
			err = fmt.Errorf("error writing DESCRIPTION data for file '%s' in ZipArchive.RewriteBinary: %s", header.Name, err)
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
				err = fmt.Errorf("error copying MD5 data to buffer in ZipArchive.RewriteBinary: %s", err)
				return
			}
			header.UncompressedSize64 = uint64(info.buffer.Len())
		}
		var md5W io.Writer
		md5W, err = zipw.CreateHeader(header)
		if err != nil {
			err = fmt.Errorf("error creating ZIP header for MD5 file '%s' in ZipArchive.RewriteBinary: %s", header.Name, err)
			return
		}
		// Write the contents to the ZIP writer. For the authoritative MD5
		// file, the buffer contains the rewritten contents. For all other
		// buffered MD5 files, the buffer contains the original contents.
		if _, err = io.Copy(md5W, info.buffer); err != nil {
			err = fmt.Errorf("error writing MD5 data for file '%s' in ZipArchive.RewriteBinary: %s", header.Name, err)
			return
		}
	}

	// Calculate result to return. Note that there is a `defer` near the
	// top of this function that mutates the returned results further by
	// setting the RewrittenChecksum and RewrittenSize properties.
	results = &Results{
		OriginalSize:     szOrig,
		OriginalChecksum: shaOrig,
		Description:      descriptionText,
	}

	return
}

func NewArchiveZip(bufferSize int) *ZipArchive {
	return &ZipArchive{
		bufferSize: bufferSize,
	}
}
