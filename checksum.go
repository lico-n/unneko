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

type PatchMetadata struct {
	Checksum       Checksum
	Name           string   `json:"name"`
	Version        int      `json:"version"`
	FromVersion    int      `json:"fromversion"`
	DownloadServer []string `json:"downloadserver"`
	VersionServer  []string `json:"versionserver"`
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

func findChecksumFiles(neko *NekoData) []*ChecksumFile {
	alreadyFound := make(map[int]bool)
	var checksumFileStarts [][]byte

	if neko.isPatch {
		checksumFileStarts = append(checksumFileStarts,
			[]byte(`{"name`),
		)
	}

	checksumFileStarts = append(checksumFileStarts,
		[]byte(`{"f`),
		[]byte("{\n "),
	)

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

	var checkSumFiles []*ChecksumFile
	var patchMetadata *PatchMetadata

	for _, fileStart := range possibleCheckSumFileStarts {
		neko.Seek(fileStart)
		headerBytes := tryUncompressHeader(neko, 1)
		if len(headerBytes) == 0 {
			continue
		}

		neko.Seek(fileStart)
		checkSumFile := tryExtractChecksumFile(neko)
		if checkSumFile != nil {
			checkSumFiles = append(checkSumFiles, checkSumFile)
			continue
		}

		if neko.isPatch && patchMetadata == nil {
			neko.Seek(fileStart)
			patchMetadata = tryExtractPatchMetadata(neko)
		}
	}

	if patchMetadata != nil {
		for _, v := range checkSumFiles {
			v.Files["patch-meta.json"] = patchMetadata.Checksum
		}
	}

	return checkSumFiles
}

func tryExtractChecksumFile(neko *NekoData) *ChecksumFile {
	defer func() {
		if r := recover(); r != nil {
			// do nothing
		}
	}()
	file := extractJSONObjectFile(neko)

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

func tryExtractPatchMetadata(neko *NekoData) *PatchMetadata {
	defer func() {
		if r := recover(); r != nil {
			// do nothing
		}
	}()
	file := extractJSONObjectFile(neko)

	var patchMetadataFile *PatchMetadata

	if err := json.Unmarshal(file.data, &patchMetadataFile); err != nil {
		return nil
	}

	if patchMetadataFile.Name == "" || len(patchMetadataFile.DownloadServer) == 0 {
		return nil
	}

	patchMetadataFile.Checksum = Checksum{
		Adler32: adler32.Checksum(file.data),
		Crc32:   crc32.ChecksumIEEE(file.data),
		Size:    len(file.data),
	}

	return patchMetadataFile
}

func restoreFileNames(checksumFiles []*ChecksumFile, extractedChan chan *extractedFile) chan *extractedFile {

	copies := make([]*ChecksumFile, 0, len(checksumFiles))
	for _, v := range checksumFiles {
		copies = append(copies, v.Copy())
	}

	checksumFiles = copies

	resultCh := make(chan *extractedFile, 1)
	go func() {
		defer close(resultCh)
		var fileIndex = 0

		for file := range extractedChan {
			if len(checksumFiles) > 0 {
				file.filePath = restoreFileName(checksumFiles, file, fileIndex)
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

func restoreFileName(checksumFiles []*ChecksumFile, file *extractedFile, fileIndex int) string {
	checksum := crc32.ChecksumIEEE(file.data)

	for _, checksumFile := range checksumFiles {
		for fileName, checksums := range checksumFile.Files {
			if checksums.Crc32 == checksum && len(file.data) == checksums.Size {
				delete(checksumFile.Files, fileName)
				return fileName
			}
		}
	}


	if file.filePath != "" {
		return file.filePath
	}

	return fmt.Sprintf("%d%s", fileIndex, file.fileExtension)
}
