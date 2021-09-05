package main

import (
	"bytes"
)

var (
	luacFileHeader             = []byte{0x1B, 0x4C, 0x75, 0x61, 0x53}
	luacFileFooter             = []byte{0x5F, 0x45, 0x4E, 0x56}
	decompilableLuacFileHeader = []byte{
		0x1B, 0x4C, 0x75, 0x61, 0x53, 0x00, 0x19, 0x93, 0x0D, 0x0A, 0x1A, 0x0A, 0x04, 0x08, 0x04, 0x08,
		0x08,
	}
)

func extractLuacFile(neko *NekoData,  keepOriginalLuacHeader bool) *extractedFile {
	uncompressed := uncompressNeko(neko, newLuacEndCompleteCond())
	if !keepOriginalLuacHeader {
		uncompressed = fixUncompressedLuacFileHeader(uncompressed)
	}

	return &extractedFile{
		data:     uncompressed,
		filePath: getOriginalLuaFilePath(uncompressed),
		fileExtension: ".luac",
	}
}

func fixUncompressedLuacFileHeader(data []byte) []byte {
	fixedData := make([]byte, 0, len(data)+1)

	fixedData = append(fixedData, decompilableLuacFileHeader...)
	fixedData = append(fixedData, data[0x10:]...)

	return fixedData
}

func getOriginalLuaFilePath(data []byte) string {
	filePathLength := 0

	filePathIndex := 0x24

	for i := filePathIndex; i < len(data); i++ {
		if data[i] == 0x00 {
			break
		}
		filePathLength++
	}

	filePath := string(data[filePathIndex : filePathIndex+filePathLength])

	return filePath
}

type luacEndCompleteCond struct{}

func newLuacEndCompleteCond() *luacEndCompleteCond {
	return &luacEndCompleteCond{}
}

func (c *luacEndCompleteCond) Complete(neko *NekoData, uncompressed []byte) bool {
	return neko.FullyRead() || c.isEndOfFile(neko, uncompressed)
}

func (c *luacEndCompleteCond) InterruptBlock(neko *NekoData, uncompressedBlock []byte) bool {
	return neko.FullyRead() || c.isEndOfFile(neko, uncompressedBlock)
}

func (c *luacEndCompleteCond) endsInFileFooter(uncompressed []byte) bool {
	if len(uncompressed) < len(luacFileFooter) {
		return false
	}

	lastUncompressed := uncompressed[len(uncompressed)-len(luacFileFooter):]

	return bytes.Compare(lastUncompressed, luacFileFooter) == 0
}

func (c *luacEndCompleteCond) isEndOfFile(neko *NekoData, uncompressed []byte) bool {
	if !c.endsInFileFooter(uncompressed) {
		return false
	}

	startOffset := neko.CurrentOffset()
	nextHeader := tryUncompressHeader(neko, 1)
	neko.Seek(startOffset)

	// is followed by the next luac file
	if len(nextHeader) >= 5 && bytes.Compare(nextHeader[:5], luacFileHeader) == 0 {
		return true
	}

	// is followed by the checksum file
	if len(nextHeader) >= 1 && nextHeader[0] == '{' {
		return true
	}

	// check if this is the last file ending
	if neko.StillContains(luacFileFooter) {
		return false
	}

	return true
}


