package main

import (
	"encoding/json"
	"hash/crc32"
)

type Checksum struct {
	Adler int `json:"adler"`
	Crc32 uint32 `json:"crc32"`
	Size  int `json:"size"`
}

type ChecksumFile struct {
	Files map[string]Checksum `json:"files"`
}

func restoreFileNames(files []*extractedFile) {
	checksumFile := findChecksumFile(files)
	if checksumFile == nil {
		return
	}

	for _, v := range files {
		fileName := findFileName(checksumFile, v)
		if fileName != "" {
			v.filePath = fileName
		}

	}

}

func findChecksumFile(files []*extractedFile) *ChecksumFile {
	for _, v := range files {
		if v.fileExtension != ".json" {
			continue
		}

		var integrityFile *ChecksumFile

		if err := json.Unmarshal(v.data, &integrityFile); err != nil {
			continue
		}

		if len(integrityFile.Files) == 0 {
			continue
		}

		v.filePath = "checksum.json"

		return integrityFile
	}

	return nil
}

func findFileName(checksumFile *ChecksumFile, file *extractedFile) string {
	checksum := crc32.ChecksumIEEE(file.data)

	for fileName, checksums := range checksumFile.Files {
		if checksums.Crc32 == checksum && len(file.data) == checksums.Size {
			return fileName
		}
	}

	return ""
}