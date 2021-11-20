package main

import (
	"encoding/json"
	"errors"
	"hash/crc32"
)

type Checksum struct {
	Crc32 uint32 `json:"crc32"`
	Size  int    `json:"size"`
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

type PatchMetadata struct {
	Checksum       Checksum
	Name           string   `json:"name"`
	DownloadServer []string `json:"downloadserver"`
}

// findChecksumFiles tries to find checksum files in NekoData.
// It tries to utilize the fact that the first lz4 sequence always contains literals.
// So we search for the start of an json object and try to decompress it to check
// whether this is a checksum file. This is a pretty inefficient process
// so we try to reduce the amount of possible json file starts by including a few more literals
// At the same time we look for patch metadata which is included in .patch.metadata
// we need to identify it because it's not included in the checksum file
func findChecksumFiles(neko *NekoData) []*ChecksumFile {
	alreadyFound := make(map[int]bool)
	var fileHeaders [][]byte

	if neko.isPatch {
		fileHeaders = append(fileHeaders,
			[]byte(`{"name`), // patch metadata filestart
		)
	}

	fileHeaders = append(fileHeaders,
		[]byte(`{"f`),  // checksum file start when it's unformatted
		[]byte("{\n "), // checksum/patch metadata file start when it's formmted
	)

	var fileStartIndices []int

	// search for possible file starts
	for _, checksumFileStart := range fileHeaders {
		neko.Seek(0)
		for i := neko.Index(checksumFileStart); i != -1; i = neko.Index(checksumFileStart) {
			if alreadyFound[i] {
				neko.Seek(i + 1)
				continue
			}

			alreadyFound[i] = true
			// As the files are not that big it's very likely that the lz4 token and extended lit length
			// starts one or two bytes before the literal
			fileStartIndices = append(fileStartIndices, i-2)
			fileStartIndices = append(fileStartIndices, i-1)
			neko.Seek(i + 1)
		}
	}

	var checkSumFiles []*ChecksumFile
	var patchMetadata *PatchMetadata

	// try to decompress on each possible file start
	// we need to look for multiple checksum files because we might get the wrong one
	// from a nested nekodata
	for _, fileStart := range fileStartIndices {
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

	// for convenience reasons we will add the patch metadata to the checksum file
	// so everything else can be handled the same
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
		Crc32: crc32.ChecksumIEEE(file.data),
		Size:  len(file.data),
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
		Crc32: crc32.ChecksumIEEE(file.data),
		Size:  len(file.data),
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

		for file := range extractedChan {
			fileName, err := restoreFileName(checksumFiles, file)
			if err != nil {
				panic(err)
			}

			file.filePath = fileName
			resultCh <- file
		}
	}()

	return resultCh
}


// restoreFileName tries to find the filename by using the checksum
// It's possible that there are multiples files with the same checksum for very short luas
// f.e `return {}`, we delete a checksum entry so the next time we find the same checksum
// we will generate the next filename.
// In general this should work fine but it's possible we accidentally find a filename
// in a nested checksum. This is very unlikely though as the .patch.metadata contains bigger files
// with different file contents. Normal nekodata will only have one checksum file
func restoreFileName(checksumFiles []*ChecksumFile, file *extractedFile) (string, error) {
	checksum := crc32.ChecksumIEEE(file.data)

	for _, checksumFile := range checksumFiles {
		for fileName, checksums := range checksumFile.Files {
			if checksums.Crc32 == checksum && len(file.data) == checksums.Size {
				delete(checksumFile.Files, fileName)
				return fileName, nil
			}
		}
	}

	return "", errors.New("unable to find filename")
}
