package main

import (
	"hash/crc32"
	"sort"
)

func extractUncompressed(neko *NekoData, checksumFiles []*ChecksumFile) *extractedFile {
	var possibleFileSizes []int
	alreadySeen := make(map[int]bool)

	for _, v := range checksumFiles {
		for _, f := range v.Files {
			if alreadySeen[f.Size] {
				continue
			}

			alreadySeen[f.Size] = true

			possibleFileSizes = append(possibleFileSizes, f.Size)
		}
	}

	sort.Ints(possibleFileSizes)

	var data []byte

	for _, v := range possibleFileSizes {
		toRead := v - len(data)
		newData := neko.ReadBytes(toRead)
		data = append(data, newData...)

		checksum := crc32.ChecksumIEEE(data)
		if ok := containsChecksum(checksumFiles, checksum); ok {
			return &extractedFile{
				data: data,
			}
		}
	}

	return &extractedFile{
		data: data,
	}
}

func containsChecksum(checksumFiles []*ChecksumFile, checksum uint32) bool {
	for _, v := range checksumFiles {
		for _, f := range v.Files {
			if f.Crc32 == checksum {
				return true
			}
		}
	}

	return false
}
