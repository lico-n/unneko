package unneko

type bracketCounterCompleteCond struct {
	openBracket   byte
	closeBracket  byte
	previousDelta int
	previousPos   int
}

func newBracketCounterCompleteCond(openBracket byte, closeBracket byte) *bracketCounterCompleteCond {
	return &bracketCounterCompleteCond{
		openBracket:  openBracket,
		closeBracket: closeBracket,
	}
}

func (c *bracketCounterCompleteCond) Complete(neko *nekoData, uncompressed []byte) bool {
	currentDelta := c.getBracketDelta(uncompressed[c.previousPos:])

	delta := c.previousDelta + currentDelta

	c.previousPos = len(uncompressed)
	c.previousDelta = delta

	return neko.fullyRead() || (len(uncompressed) > 0 && delta == 0)
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
