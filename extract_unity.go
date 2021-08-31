package main

import (
	"encoding/binary"
)

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

func nextFileIsUnityFile(neko *NekoData) bool {
	startOffset := neko.CurrentOffset()
	defer func() {
		neko.Seek(startOffset)
	}()

	headerBytes := tryUncompressHeader(neko, 1)

	if len(headerBytes) == 0 {
		return false
	}

	fileSignature := readNullTerminatedString(headerBytes)

	return fileSignature == "UnityFS"
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
	alreadyUncompressed int
}

func newMaxUncompressedSizeCompleteCond(maxSize int) *maxUncompressedSizeCompleteCond {
	return &maxUncompressedSizeCompleteCond{maxUncompressedSize: maxSize}
}

func (c *maxUncompressedSizeCompleteCond) Complete(neko *NekoData, uncompressed []byte) bool {
	c.alreadyUncompressed = len(uncompressed)
	return neko.FullyRead() || c.maxUncompressedSize <= c.alreadyUncompressed
}

func (c *maxUncompressedSizeCompleteCond) InterruptBlock(neko *NekoData, uncompressedBlock []byte) bool {
	return neko.FullyRead() || c.maxUncompressedSize <= c.alreadyUncompressed+len(uncompressedBlock)
}
