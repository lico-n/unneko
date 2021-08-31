package main

func extractJSONFiles(neko *NekoData) ([]*extractedFile, error) {
	var extracted []*extractedFile

	for hasAnotherJSONFile(neko) {
		uncompressed := uncompressNeko(neko, newJSONObjectCompleteCond())

		extracted = append(extracted, &extractedFile{data: uncompressed})

		currentNekoOffset := neko.CurrentOffset()
		remainingNekoBytes := neko.RemainingBytes()

		neko = neko.Slice(currentNekoOffset, remainingNekoBytes)
	}

	return extracted, nil
}

func hasAnotherJSONFile(neko *NekoData) bool {
	headerBytes := uncompressHeader(neko, 1)
	neko.Reset()

	return headerBytes[0] == '{'
}
