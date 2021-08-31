package main

type CompleteCond interface {
	Complete(neko *NekoData, uncompressed []byte) bool
	InterruptBlock(neko *NekoData, uncompressedBlock []byte) bool
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

type jsonObjectCompleteCond struct {
	openBrackets int
}

func newJSONObjectCompleteCond() *jsonObjectCompleteCond {
	return &jsonObjectCompleteCond{}
}

func (c *jsonObjectCompleteCond) Complete(neko *NekoData, uncompressed []byte) bool {
	c.openBrackets = c.getBracketDelta(uncompressed)

	return neko.FullyRead() || (len(uncompressed) > 0 && c.openBrackets == 0)
}

func (c *jsonObjectCompleteCond) InterruptBlock(neko *NekoData, uncompressedBlock []byte) bool {
	currentlyOpenBrackets := c.openBrackets + c.getBracketDelta(uncompressedBlock)
	return neko.FullyRead() || currentlyOpenBrackets == 0
}

func (c *jsonObjectCompleteCond) getBracketDelta(data []byte) int {
	bracketDelta := 0

	for i := 0; i < len(data); i++ {
		if data[i] == '{' {
			bracketDelta++
		}
		if data[i] == '}' {
			bracketDelta--
		}
	}

	return bracketDelta
}
