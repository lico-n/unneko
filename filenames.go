package main

import (
	"encoding/json"
	"fmt"
	"hash/adler32"
	"hash/crc32"
)

type Checksum struct {
	Adler32 uint32 `json:"adler32"`
	Crc32   uint32 `json:"crc32"`
	Size    int    `json:"size"`
}

type ChecksumFile struct {
	Files map[string]Checksum `json:"files"`
}

func (cf *ChecksumFile) Copy() *ChecksumFile {
	m := make(map[string]Checksum, len(cf.Files))
	for k, v := range cf.Files {
		m[k] = v
	}
	return &ChecksumFile{
		Files: m,
	}
}

func findChecksumFile(neko *NekoData) *ChecksumFile {
	alreadyFound := make(map[int]bool)
	checksumFileStarts := [][]byte{
		[]byte(`{"f`),
		[]byte("{\n "),
		[]byte(`{`),
	}

	var possibleCheckSumFileStarts []int
	for _, checksumFileStart := range checksumFileStarts {
		neko.Seek(0)
		for i := neko.Index(checksumFileStart); i != -1; i = neko.Index(checksumFileStart) {
			if alreadyFound[i] {
				neko.Seek(i + 1)
				continue
			}

			alreadyFound[i] = true
			possibleCheckSumFileStarts = append(possibleCheckSumFileStarts, i-2)
			possibleCheckSumFileStarts = append(possibleCheckSumFileStarts, i-1)
			neko.Seek(i + 1)
		}
	}

	for _, fileStart := range possibleCheckSumFileStarts {
		neko.Seek(fileStart)
		headerBytes := tryUncompressHeader(neko, 1)
		if len(headerBytes) == 0 {
			continue
		}

		neko.Seek(fileStart)
		checkSumFile := tryExtractChecksumFile(neko)
		if checkSumFile != nil {
			return checkSumFile
		}
	}

	return nil
}

func tryExtractChecksumFile(neko *NekoData) *ChecksumFile {
	defer func() {
		if r := recover(); r != nil {
			// do nothing
		}
	}()
	file := extractJSONObjectFile(neko)
	return isChecksumFile(file)
}

func restoreFileNames(checksumFile *ChecksumFile, extractedChan chan *extractedFile) chan *extractedFile {

	resultCh := make(chan *extractedFile, 1)
	go func() {
		defer close(resultCh)
		var fileIndex = 0

		for file := range extractedChan {
			if checksumFile != nil {
				file.filePath = restoreFileName(checksumFile, file, fileIndex)
				resultCh <- file
				continue
			}

			if file.filePath == "" {
				file.filePath = fmt.Sprintf("%d%s", fileIndex, file.fileExtension)
			}

			resultCh <- file

		}
	}()

	return resultCh
}

func isChecksumFile(file *extractedFile) *ChecksumFile {
	if file.fileExtension != ".json" {
		return nil
	}

	var checksumFile *ChecksumFile

	if err := json.Unmarshal(file.data, &checksumFile); err != nil {
		return nil
	}

	if checksumFile.Files == nil {
		return nil
	}

	file.filePath = "checksum.json"

	c := Checksum{
		Adler32: adler32.Checksum(file.data),
		Crc32:   crc32.ChecksumIEEE(file.data),
		Size:    len(file.data),
	}

	checksumFile.Files[file.filePath] = c

	return &ChecksumFile{
		Files: checksumFile.Files,
	}
}

func restoreFileName(checksumFile *ChecksumFile, file *extractedFile, fileIndex int) string {
	checksumFile = checksumFile.Copy()
	checksum := crc32.ChecksumIEEE(file.data)

	for fileName, checksums := range checksumFile.Files {
		if checksums.Crc32 == checksum && len(file.data) == checksums.Size {
			delete(checksumFile.Files, fileName)
			return fileName
		}
	}

	if file.filePath != "" {
		return file.filePath
	}

	return fmt.Sprintf("%d%s", fileIndex, file.fileExtension)
}
