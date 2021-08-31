package main

import (
	"fmt"
)

var (
	luacFileHeader             = []byte{0x1B, 0x4C, 0x75, 0x61, 0x53}
	luacFileFooter             = []byte{0x5F, 0x45, 0x4E, 0x56}
	decompilableLuacFileHeader = []byte{
		0x1B, 0x4C, 0x75, 0x61, 0x53, 0x00, 0x19, 0x93, 0x0D, 0x0A, 0x1A, 0x0A, 0x04, 0x08, 0x04, 0x08,
		0x08,
	}
)

func extractLuacFiles(bigNeko *NekoData, keepOriginalLuacHeader bool) ([]*extractedFile, error) {
	var extracted []*extractedFile

	nekos, checksumFile, err := splitLuaFiles(bigNeko)
	if err != nil {
		return nil, fmt.Errorf("splitting lua files: %v", err)
	}

	if checksumFile != nil {
		extracted = append(extracted, checksumFile)
	}

	for _, neko := range nekos {
		uncompressed := uncompressNeko(neko, newNekoEndCompleteCond())

		if !keepOriginalLuacHeader {
			uncompressed = fixUncompressedLuacFileHeader(uncompressed)
		}

		extracted = append(extracted, &extractedFile{
			data:     uncompressed,
			filePath: getOriginalLuaFilePath(uncompressed),
		})
	}



	return extracted, nil
}


func splitLuaFiles(neko *NekoData) ([]*NekoData, *extractedFile, error) {
	var files []*NekoData
	var checksumFile *extractedFile

	headerIndices := neko.AllPatternIndices(luacFileHeader)
	possibleFooterIndices := neko.AllPatternIndices(luacFileFooter)

	previousEnd := 0

	nextFooterArrIndex := 0

	for i := 0; i < len(headerIndices); i++ {
		if i+1 == len(headerIndices) {
			lastFooterIndex := possibleFooterIndices[len(possibleFooterIndices)-1]
			subNeko := neko.Slice(previousEnd, lastFooterIndex+len(luacFileFooter))
			files = append(files, subNeko)
			break
		}

		nextHeaderIndex := headerIndices[i+1]

		footerIndexCandidate := possibleFooterIndices[nextFooterArrIndex]
		nextFooterArrIndex++

		if footerIndexCandidate > nextHeaderIndex {
			return nil, nil, fmt.Errorf("missing footer for luac file starting at index %08X", previousEnd)
		}

		for {
			if possibleFooterIndices[nextFooterArrIndex] > nextHeaderIndex {
				break
			}

			footerIndexCandidate = possibleFooterIndices[nextFooterArrIndex]
			nextFooterArrIndex++
		}

		neko.Seek(previousEnd)
		if nextFileIsJSONObject(neko) {
			uncompressed := uncompressNeko(neko, newBracketCounterCompleteCond('{', '}'))
			checksumFile =  &extractedFile{
				data: uncompressed,
				fileExtension: ".json",
			}
			previousEnd = neko.CurrentOffset()
		}

		nextEnd := footerIndexCandidate + len(luacFileFooter)
		subNeko := neko.Slice(previousEnd, nextEnd)

		files = append(files, subNeko)

		previousEnd = nextEnd
	}

	return files, checksumFile, nil
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

type nekoEndCompleteCond struct{}

func newNekoEndCompleteCond() *nekoEndCompleteCond {
	return &nekoEndCompleteCond{}
}

func (c *nekoEndCompleteCond) Complete(neko *NekoData, uncompressed []byte) bool {
	return neko.FullyRead()
}

func (c *nekoEndCompleteCond) InterruptBlock(neko *NekoData, uncompressedBlock []byte) bool {
	return neko.FullyRead()
}
