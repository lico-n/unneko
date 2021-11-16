package main

import (
	"hash/crc32"
	"sort"
)

func extractWithChecksum(neko *NekoData, completeCond CompleteCond) *extractedFile {
	uncompressed := uncompressNeko(neko, completeCond)

	return &extractedFile{
		data: uncompressed,
	}
}

type checksumFileMap struct {
	checksums map[uint32]int
	fileSizes map[int]struct{}
}

type checksumCompleteCond struct {
	maps []*checksumFileMap
}

func newChecksumCompleteCond(checksumFiles []*ChecksumFile) *checksumCompleteCond {
	maps := make([]*checksumFileMap, 0, len(checksumFiles))

	for _, v := range checksumFiles {
		checksums := make(map[uint32]int, len(v.Files))
		fileSizes := make(map[int]struct{}, len(v.Files))

		for _, f := range v.Files {
			checksums[f.Crc32]++
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

func (c *checksumCompleteCond) MarkAsFound(data []byte) {
	checksum := crc32.ChecksumIEEE(data)

	for _, v := range c.maps {
		if v.checksums[checksum] > 0 {
			v.checksums[checksum]--
		}
	}
}

func (c *checksumCompleteCond) FoundAll() bool {
	for _, m := range c.maps {
		// if a single map is completed, we consider the extraction complete
		// all other maps are probably nested checksum files of compressed nekodata
		foundAll := true
		for _, v := range m.checksums {
			if v > 0 {
				foundAll = false
				break
			}
		}
		if foundAll {
			return true
		}
	}

	return false
}

func (c *checksumCompleteCond) PossibleFileSizes() []int {

	var result []int
	alreadySeen := make(map[int]bool)

	for _, m := range c.maps {
		for fs, _ := range m.fileSizes {
			if alreadySeen[fs] {
				continue
			}
			alreadySeen[fs] = true
			result = append(result, fs)
		}
	}

	sort.Ints(result)

	return result
}
