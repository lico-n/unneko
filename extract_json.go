package main

func extractJSONObjectFile(neko *NekoData) *extractedFile {
	uncompressed := uncompressNeko(neko, newBracketCounterCompleteCond('{', '}'))
	return &extractedFile{
		data:          uncompressed,
		fileExtension: ".json",
	}
}

func extractJSONArrayFile(neko *NekoData) *extractedFile {
	uncompressed := uncompressNeko(neko, newBracketCounterCompleteCond('[', ']'))
	return &extractedFile{
		data:          uncompressed,
		fileExtension: ".json",
	}
}

type bracketCounterCompleteCond struct {
	openBracket byte
	closeBracket byte
}

func newBracketCounterCompleteCond(openBracket byte, closeBracket byte) *bracketCounterCompleteCond {
	return &bracketCounterCompleteCond{
		openBracket: openBracket,
		closeBracket: closeBracket,
	}
}

func (c *bracketCounterCompleteCond) Complete(neko *NekoData, uncompressed []byte) bool {
	delta :=  c.getBracketDelta(uncompressed)

	return neko.FullyRead() || (len(uncompressed) > 0 && delta == 0)
}

func (c *bracketCounterCompleteCond) getBracketDelta(data []byte) int {
	bracketDelta := 0

	for i := 0; i < len(data); i++ {
		if data[i] == c.openBracket {
			bracketDelta++
		}
		if data[i] == c.closeBracket {
			bracketDelta--
		}
	}

	return bracketDelta
}
