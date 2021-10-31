package main

import (
	"hash/crc32"
)

func extractPlainFile(neko *NekoData, completeCond CompleteCond) *extractedFile {
	uncompressed := uncompressNeko(neko, completeCond)

	return &extractedFile{
		data:     uncompressed,
		fileExtension: ".lua",
	}
}

type checksumCompleteCond struct{
	checksums map[uint32]bool
}

func newChecksumCompleteCond(checksumFile *ChecksumFile) *checksumCompleteCond {
	checksums := make(map[uint32]bool, len(checksumFile.Files))

	for _, v := range checksumFile.Files {
		checksums[v.Crc32] = true
	}

	return &checksumCompleteCond{
		checksums: checksums,
	}
}

func (c *checksumCompleteCond) Complete(neko *NekoData, uncompressed []byte) bool {
	if neko.FullyRead() {
		return true
	}

	checksum := crc32.ChecksumIEEE(uncompressed)
	if c.checksums[checksum] {
		return true
	}

	return false
}
