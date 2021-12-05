package unneko

// completeCond will indicate when the uncompressed file is finished
// Compared to normal lz4 decompression we do not know the file boundaries
// so we need to know when to stop the decompression in other ways.
type completeCond interface {
	Complete(neko *nekoData, uncompressed []byte) bool
}

func tryUncompressHeader(neko *nekoData, numberOfSeq int) []byte {
	startOffset := neko.currentOffset()
	defer func() {
		recover() // ignore panics
		neko.seek(startOffset)
	}()

	var uncompressed []byte

	for i := 0; i < numberOfSeq; i++ {
		token := readToken(neko)

		literals := neko.readBytes(token.nrOfLiterals)
		uncompressed = append(uncompressed, literals...)

		if token.tokenByte&0xF == 0 {
			beforeMatchOffset := neko.currentOffset()
			matchOffset := readMatchOffset(neko)
			if len(uncompressed)-matchOffset <= 0 {
				neko.seek(beforeMatchOffset)
				break
			}
			neko.seek(beforeMatchOffset)
		}

		matchOffset := readMatchOffset(neko)
		extendedMatchLen := readExtendedMatchLength(token, neko)

		uncompressed = appendMatches(uncompressed, token.nrOfMatches+extendedMatchLen, matchOffset)
	}

	return uncompressed
}

func uncompressNeko(neko *nekoData, cCond completeCond) []byte {
	var (
		uncompressed []byte
		complete     bool
	)

	for !cCond.Complete(neko, uncompressed) {
		uncompressed, complete = uncompressNekoBlock(neko, uncompressed, cCond)
		if complete {
			break
		}
	}

	return uncompressed
}

func uncompressNekoBlock(neko *nekoData, uncompressed []byte, completeCond completeCond) ([]byte, bool) {
	previouslyUncompressed := len(uncompressed)

	for !neko.fullyRead() {
		token := readToken(neko)

		literals := neko.readBytes(token.nrOfLiterals)
		uncompressed = append(uncompressed, literals...)

		// maximum block length is 0x8000 bytes. The file is not complete yet
		// but the lz4 decompression algorithm will reset and start again
		if len(uncompressed)-previouslyUncompressed == 0x8000 {
			return uncompressed, false
		}
		if completeCond.Complete(neko, uncompressed) {
			return uncompressed, true
		}

		matchOffset := readMatchOffset(neko)
		extendedMatchLen := readExtendedMatchLength(token, neko)
		nrOfMatches := token.nrOfMatches + extendedMatchLen
		uncompressed = appendMatches(uncompressed, nrOfMatches, matchOffset)
	}

	return uncompressed, true
}

type token struct {
	nrOfLiterals int
	nrOfMatches  int
	tokenByte    byte
}

func readToken(neko *nekoData) token {
	tokenByte := neko.readByte()

	litLength := int(tokenByte >> 4)
	matLength := int((tokenByte & 0xF) + 4)

	if litLength == 0xF {
		extendedLitLength := neko.readByte()

		litLength += int(extendedLitLength)
		for extendedLitLength == 0xFF {
			extendedLitLength = neko.readByte()
			litLength += int(extendedLitLength)
		}
	}

	return token{
		nrOfLiterals: litLength,
		nrOfMatches:  matLength,
		tokenByte:    tokenByte,
	}
}

func readMatchOffset(neko *nekoData) int {
	matchOffsetBytes := neko.readBytes(2)
	matchOffset := (int(matchOffsetBytes[1]) << 8) | int(matchOffsetBytes[0])

	return matchOffset
}

func readExtendedMatchLength(token token, neko *nekoData) int {
	if token.nrOfMatches != 0xF+4 {
		return 0
	}

	nextByte := neko.readByte()

	extendedMatLength := int(nextByte)

	for nextByte == 0xFF {
		nextByte = neko.readByte()
		extendedMatLength += int(nextByte)
	}

	return extendedMatLength
}

func appendMatches(uncompressed []byte, totalMatchLength int, matchOffset int) []byte {
	if totalMatchLength == 0 {
		return uncompressed
	}
	absoluteMatchOffset := len(uncompressed) - matchOffset

	for i := absoluteMatchOffset; i < absoluteMatchOffset+totalMatchLength; i++ {
		uncompressed = append(uncompressed, uncompressed[i])
	}

	return uncompressed
}
