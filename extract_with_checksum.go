package main

import (
	"hash/crc32"
)

func extractWithChecksum(neko *NekoData, completeCond CompleteCond) *extractedFile {
	uncompressed := uncompressNeko(neko, completeCond)

	return &extractedFile{
		data:          uncompressed,
	}
}

type checksumFileMap struct {
	checksums map[uint32]struct{}
	fileSizes map[int]struct{}
}

type checksumCompleteCond struct {
	maps []*checksumFileMap
}

func newChecksumCompleteCond(checksumFiles []*ChecksumFile) *checksumCompleteCond {
	maps := make([]*checksumFileMap, 0, len(checksumFiles))

	for _, v := range checksumFiles {
		checksums := make(map[uint32]struct{}, len(v.Files))
		fileSizes := make(map[int]struct{}, len(v.Files))

		for _, f := range v.Files {
			checksums[f.Crc32] = struct{}{}
			fileSizes[f.Size] = struct{}{}
		}

		maps = append(maps, &checksumFileMap{
			checksums: checksums,
			fileSizes: fileSizes,
		})
	}

	return &checksumCompleteCond{
		maps: maps,
	}
}

func (c *checksumCompleteCond) Complete(neko *NekoData, uncompressed []byte) bool {
	if neko.FullyRead() {
		return true
	}

	validFileSize := false

	for _, v := range c.maps {
		if _, ok := v.fileSizes[len(uncompressed)]; ok {
			validFileSize = true
			break
		}
	}

	if !validFileSize {
		return false
	}

	checksum := crc32.ChecksumIEEE(uncompressed)

	for _, v := range c.maps {
		if _, ok := v.checksums[checksum]; ok {
			return true
		}
	}

	return false
}
