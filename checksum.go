package unneko

import (
	"encoding/json"
	"hash/crc32"
)

type checksum struct {
	Crc32 uint32 `json:"crc32"`
	Size  int    `json:"size"`
}

type checksumFile struct {
	Files map[string]checksum `json:"files"`
}

func (cf *checksumFile) Add(fileName string, data []byte) {
	cf.Files[fileName] = checksum{
		Crc32: crc32.ChecksumIEEE(data),
		Size:  len(data),
	}
}

func (cf *checksumFile) Copy() *checksumFile {
	m := make(map[string]checksum, len(cf.Files))
	for k, v := range cf.Files {
		m[k] = v
	}
	return &checksumFile{
		Files: m,
	}
}

type patchMetadata struct {
	raw            []byte
	Name           string   `json:"name"`
	DownloadServer []string `json:"downloadserver"`
}

// findChecksumFiles tries to find checksum files in nekoData.
// It tries to utilize the fact that the first lz4 sequence always contains literals.
// So we search for the start of an json object and try to decompress it to check
// whether this is a checksum file. This is a pretty inefficient process
// so we try to reduce the amount of possible json file starts by including a few more literals
// At the same time we look for patch metadata which is included in .patch.metadata
// we need to identify it because it's not included in the checksum file
func findChecksumFiles(neko *nekoData) []*checksumFile {
	startOffset := neko.currentOffset()
	defer func(){
		neko.seek(startOffset)
	}()

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
		neko.seek(0)
		for i := neko.index(checksumFileStart); i != -1; i = neko.index(checksumFileStart) {
			if alreadyFound[i] {
				neko.seek(i + 1)
				continue
			}

			alreadyFound[i] = true
			// As the files are not that big it's very likely that the lz4 token and extended lit length
			// starts one or two bytes before the literal
			fileStartIndices = append(fileStartIndices, i-2)
			fileStartIndices = append(fileStartIndices, i-1)
			neko.seek(i + 1)
		}
	}

	var cSumFiles []*checksumFile
	var pMD *patchMetadata

	// try to decompress on each possible file start
	// we need to look for multiple checksum files because we might get the wrong one
	// from a nested nekodata
	for _, fileStart := range fileStartIndices {
		neko.seek(fileStart)
		headerBytes := tryUncompressHeader(neko, 1)
		if len(headerBytes) == 0 {
			continue
		}

		csFile := tryExtractChecksumFile(neko)
		if csFile != nil {
			cSumFiles = append(cSumFiles, csFile)
			continue
		}

		if neko.isPatch && pMD == nil {
			neko.seek(fileStart)
			pMD = tryExtractPatchMetadata(neko)
		}
	}

	if pMD != nil {
		for _, v := range cSumFiles {
			v.Add("patch-meta.json", pMD.raw )
		}
	}

	return cSumFiles
}

func tryExtractChecksumFile(neko *nekoData) *checksumFile {
	defer func() {
		recover() // ignore panics
	}()

	data := uncompressNeko(neko, newBracketCounterCompleteCond('{', '}'))

	var chFile *checksumFile

	if err := json.Unmarshal(data, &chFile); err != nil {
		return nil
	}

	if chFile.Files == nil {
		return nil
	}

	chFile.Add("checksum.json", data)

	return chFile
}

func tryExtractPatchMetadata(neko *nekoData) *patchMetadata {
	defer func() {
		recover() // ignore panics
	}()

	data := uncompressNeko(neko, newBracketCounterCompleteCond('{', '}'))

	var patchMetadataFile *patchMetadata

	if err := json.Unmarshal(data, &patchMetadataFile); err != nil {
		return nil
	}

	if patchMetadataFile.Name == "" || len(patchMetadataFile.DownloadServer) == 0 {
		return nil
	}

	patchMetadataFile.raw = data

	return patchMetadataFile
}
