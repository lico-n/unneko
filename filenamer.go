package unneko

import (
	"hash/crc32"
	"strconv"
)

type fileNamer struct {
	csFiles            []*checksumFile
	unknownFileCounter int
}

func newFileNamer(csumFiles []*checksumFile) *fileNamer {
	copies := make([]*checksumFile, 0, len(csumFiles))
	for _, v := range csumFiles {
		copies = append(copies, v.Copy())
	}

	return &fileNamer{
		csFiles: copies,
	}
}

func (n *fileNamer) GetName(fileData []byte) string {
	cs := crc32.ChecksumIEEE(fileData)

	for _, csFiles := range n.csFiles {
		for fileName, checksums := range csFiles.Files {
			if checksums.Crc32 == cs && len(fileData) == checksums.Size {
				delete(csFiles.Files, fileName)
				return fileName
			}
		}
	}

	n.unknownFileCounter++
	return strconv.Itoa(n.unknownFileCounter)
}
