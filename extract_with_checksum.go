package unneko

import (
	"hash/crc32"
	"sort"
)

type checksumFileMap struct {
	checksums map[uint32]int
	fileSizes map[int]struct{}
}

type checksumCompleteCond struct {
	maps []*checksumFileMap
}

func newChecksumCompleteCond(checksumFiles []*checksumFile) *checksumCompleteCond {
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

func (c *checksumCompleteCond) Complete(neko *nekoData, uncompressed []byte) bool {
	if neko.fullyRead() {
		return true
	}

	// To optimize the decompression we will only calculate the checksum when the len(uncompressed)
	// matches a file in the checksumfile.
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

	// file successfully extracted if it's a known checksum from the checksum file
	chSum := crc32.ChecksumIEEE(uncompressed)

	for _, v := range c.maps {
		if n, ok := v.checksums[chSum]; ok {
			return n > 0
		}
	}

	return false
}

func (c *checksumCompleteCond) MarkAsFound(data []byte) {
	chSum := crc32.ChecksumIEEE(data)

	// As we don't know which checksum file we found is the correct one
	// we will decrement the checksum on each checksum file.
	// When the first one is empty it will indicate that the decompression is finished
	for _, v := range c.maps {
		if v.checksums[chSum] > 0 {
			v.checksums[chSum]--
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
	alreadySeen := make(map[int]struct{})

	for _, m := range c.maps {
		for fs, _ := range m.fileSizes {
			if _, ok := alreadySeen[fs]; ok {
				continue
			}
			alreadySeen[fs] = struct{}{}
			result = append(result, fs)
		}
	}

	sort.Ints(result)

	return result
}
