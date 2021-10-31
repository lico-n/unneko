package main

import (
	"encoding/binary"
)


func extractUnityFile(neko *NekoData) *extractedFile {
	fileSize := readUnityFileSize(neko)
	uncompressed := uncompressNeko(neko, newMaxUncompressedSizeCompleteCond(int(fileSize)))

	return &extractedFile{
		data:          uncompressed,
		fileExtension: ".assetbundle",
	}
}


func readUnityFileSize(neko *NekoData) uint64 {
	startOffset := neko.CurrentOffset()
	defer func() {
		neko.Seek(startOffset)
	}()

	headerBytes := tryUncompressHeader(neko, 3)

	fileSignature := readNullTerminatedString(headerBytes)
	currentHeaderOffset := len(fileSignature) + 1

	currentHeaderOffset += 4 // header version

	playerVersion := readNullTerminatedString(headerBytes[currentHeaderOffset:])
	currentHeaderOffset += len(playerVersion) + 1

	engineVersion := readNullTerminatedString(headerBytes[currentHeaderOffset:])
	currentHeaderOffset += len(engineVersion) + 1

	fileSizeBytes := headerBytes[currentHeaderOffset : currentHeaderOffset+8]
	fileSize := binary.BigEndian.Uint64(fileSizeBytes)

	return fileSize
}

func readNullTerminatedString(data []byte) string {
	for i := 0; i < len(data); i++ {
		if data[i] == 0x00 {
			return string(data[0:i])
		}
	}

	return ""
}


type maxUncompressedSizeCompleteCond struct {
	maxUncompressedSize int
}

func newMaxUncompressedSizeCompleteCond(maxSize int) *maxUncompressedSizeCompleteCond {
	return &maxUncompressedSizeCompleteCond{maxUncompressedSize: maxSize}
}

func (c *maxUncompressedSizeCompleteCond) Complete(neko *NekoData, uncompressed []byte) bool {
	return neko.FullyRead() || c.maxUncompressedSize <= len(uncompressed)
}

