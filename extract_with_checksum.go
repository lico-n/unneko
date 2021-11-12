package main

import (
	"fmt"
	"hash/crc32"
)

func extractWithChecksum(neko *NekoData, completeCond CompleteCond) *extractedFile {
	uncompressed := uncompressNeko(neko, completeCond)

	return &extractedFile{
		data:          uncompressed,
		fileExtension: ".lua",
	}
}

type checksumCompleteCond struct {
	checksums map[uint32]struct{}
	fileSizes map[int]struct{}
}

func newChecksumCompleteCond(checksumFile *ChecksumFile) *checksumCompleteCond {
	checksums := make(map[uint32]struct{}, len(checksumFile.Files))
	fileSizes := make(map[int]struct{}, len(checksumFile.Files))

	for _, v := range checksumFile.Files {
		checksums[v.Crc32] = struct{}{}
		fileSizes[v.Size] = struct{}{}
	}

	return &checksumCompleteCond{
		checksums: checksums,
		fileSizes: fileSizes,
	}
}

func (c *checksumCompleteCond) Complete(neko *NekoData, uncompressed []byte) bool {
	if neko.FullyRead() {
		return true
	}

	if _, ok := c.fileSizes[len(uncompressed)]; !ok {
		return false
	}

	fmt.Println("yay")
	checksum := crc32.ChecksumIEEE(uncompressed)
	if _, ok := c.checksums[checksum]; ok {
		return true
	}

	return false
}
