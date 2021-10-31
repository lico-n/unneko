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
	count int
}

func newBracketCounterCompleteCond(openBracket byte, closeBracket byte) *bracketCounterCompleteCond {
	return &bracketCounterCompleteCond{
		openBracket: openBracket,
		closeBracket: closeBracket,
	}
}

func (c *bracketCounterCompleteCond) Complete(neko *NekoData, uncompressed []byte) bool {
	c.count = c.getBracketDelta(uncompressed)

	return neko.FullyRead() || (len(uncompressed) > 0 && c.count == 0)
}

func (c *bracketCounterCompleteCond) InterruptBlock(neko *NekoData, uncompressedBlock []byte) bool {
	currentlyOpenBrackets := c.count + c.getBracketDelta(uncompressedBlock)
	return neko.FullyRead() || currentlyOpenBrackets == 0
}

func (c *bracketCounterCompleteCond) UntilError() bool {
	return false
}

func (c *bracketCounterCompleteCond) RecordError()  {
	// do nothing
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

func (c *bracketCounterCompleteCond) IsValidUncompress(_ []byte) bool {
	return true
}