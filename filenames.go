package main

import (
	"encoding/json"
	"hash/crc32"
)

type Checksum struct {
	Adler32 uint32 `json:"adler32"`
	Crc32   uint32 `json:"crc32"`
	Size    int    `json:"size"`
}

type ChecksumFile struct {
	Files map[string]Checksum `json:"files"`
	ExtractedFile *extractedFile
}

func restoreFileNames(extractedChan chan *extractedFile) chan *extractedFile {
	var checksumFile *ChecksumFile
	var buffer []*extractedFile

	resultCh := make(chan *extractedFile,1)
	go func() {
		defer close(resultCh)

		for file := range extractedChan {
			// already found checksum file, restore filename and pass it along
			if checksumFile != nil {
				fileName := findFileName(checksumFile, file)
				if fileName != "" {
					file.filePath = fileName
				}
				resultCh <- file
				continue
			}


			checksumFile = isChecksumFile(file)

			// no checksum file write it into buffer, until we find it
			if checksumFile == nil {
				buffer = append(buffer, file)
				continue
			}

			// found checksum file, process all files in buffer and continue
			resultCh <- checksumFile.ExtractedFile
			for _, v := range buffer {
				fileName := findFileName(checksumFile, v)
				if fileName != "" {
					v.filePath = fileName
				}
				resultCh <- v
			}

			buffer = nil
		}

		for _, v := range buffer {
			resultCh <- v
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

	if len(checksumFile.Files) == 0 {
		return nil
	}

	file.filePath = "checksum.json"

	return &ChecksumFile{
		Files:         checksumFile.Files,
		ExtractedFile: file,
	}
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
