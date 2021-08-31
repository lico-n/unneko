package main


func nextFileIsJSONObject(neko *NekoData) bool {
	headerBytes := tryUncompressHeader(neko, 1)
	neko.Reset()

	if len(headerBytes) == 0 {
		return false
	}

	return headerBytes[0] == '{'
}


func nextFileIsJSONArray(neko *NekoData) bool {
	headerBytes := tryUncompressHeader(neko, 1)
	neko.Reset()

	if len(headerBytes) == 0 {
		return false
	}

	return headerBytes[0] == '['
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
