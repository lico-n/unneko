package main

import (
	"encoding/binary"
)

var (
	unityFileHeader = []byte("UnityFS")
)

func extractUnityFiles(neko *NekoData) ([]*extractedFile, error) {
	var extracted []*extractedFile

	for hasAnotherUnityFile(neko) {
		fileSize := readUnityFileSize(neko)
		uncompressed := uncompressNeko(neko, newMaxUncompressedSizeCompleteCond(int(fileSize)))
		extracted = append(extracted, &extractedFile{data: uncompressed})

		currentNekoOffset := neko.CurrentOffset()
		remainingNekoBytes := neko.RemainingBytes()

		neko = neko.Slice(currentNekoOffset, remainingNekoBytes)
	}

	return extracted, nil
}

func readUnityFileSize(neko *NekoData) uint64 {
	headerBytes := uncompressHeader(neko, 3)
	neko.Reset()

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

func hasAnotherUnityFile(neko *NekoData) bool {
	headerBytes := uncompressHeader(neko, 1)
	neko.Reset()

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
