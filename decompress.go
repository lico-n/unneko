package main


type CompleteCond interface {
	Complete(neko *NekoData, uncompressed []byte) bool
	InterruptBlock(neko *NekoData, uncompressedBlock []byte) bool
}

func tryUncompressHeader(neko *NekoData, numberOfSeq int) []byte {
	startOffset := neko.CurrentOffset()

	defer func(){
		if r := recover(); r != nil {
			// end of compressed data block reached, there are no more compressed files
			// reset to status before uncompressing attempt
			neko.Seek(startOffset)
		}
	}()

	var uncompressed []byte

	for i := 0; i < numberOfSeq; i++ {
		token := readToken(neko)

		literals := neko.ReadBytes(token.nrOfLiterals)
		uncompressed = append(uncompressed, literals...)

		matchOffset := readMatchOffset(neko)
		extendedMatchLen := readExtendedMatchLength(token, neko)

		uncompressed = appendMatches(uncompressed, token.nrOfMatches+extendedMatchLen, matchOffset)
	}

	return uncompressed
}

func uncompressNeko(neko *NekoData, completeCond CompleteCond) []byte {
	var uncompressed []byte

	for !completeCond.Complete(neko, uncompressed) {
		uncompressedBlock := uncompressNekoBlock(neko, completeCond)
		uncompressed = append(uncompressed, uncompressedBlock...)

	}

	return uncompressed
}

func uncompressNekoBlock(neko *NekoData, completeCond CompleteCond) []byte {
	var uncompressed []byte

	for !neko.FullyRead() {

		token := readToken(neko)

		literals := neko.ReadBytes(token.nrOfLiterals)
		uncompressed = append(uncompressed, literals...)

		if len(uncompressed) == 0x8000 || completeCond.InterruptBlock(neko, uncompressed) {
			break
		}

		matchOffset := readMatchOffset(neko)
		extendedMatchLen := readExtendedMatchLength(token, neko)

		uncompressed = appendMatches(uncompressed, token.nrOfMatches+extendedMatchLen, matchOffset)
	}

	return uncompressed
}

type Token struct {
	nrOfLiterals int
	nrOfMatches  int
	tokenByte    byte
}

func readToken(neko *NekoData) Token {
	tokenByte := neko.ReadByte()

	litLength := int(tokenByte >> 4)
	matLength := int((tokenByte & 0xF) + 4)

	if litLength == 0xF {
		extendedLitLength := neko.ReadByte()

		litLength += int(extendedLitLength)
		for extendedLitLength == 0xFF {
			extendedLitLength = neko.ReadByte()
			litLength += int(extendedLitLength)
		}
	}

	return Token{
		nrOfLiterals: litLength,
		nrOfMatches:  matLength,
		tokenByte:    tokenByte,
	}
}

func readMatchOffset(neko *NekoData) int {
	matchOffsetBytes := neko.ReadBytes(2)
	matchOffset := (int(matchOffsetBytes[1]) << 8) | int(matchOffsetBytes[0])

	return matchOffset
}

func readExtendedMatchLength(token Token, neko *NekoData) int {
	if token.nrOfMatches != 0xF+4 {
		return 0
	}

	nextByte := neko.ReadByte()

	extendedMatLength := int(nextByte)

	for nextByte == 0xFF {
		nextByte = neko.ReadByte()
		extendedMatLength += int(nextByte)
	}

	return extendedMatLength
}

func appendMatches(uncompressed []byte, totalMatchLength int, matchOffset int) []byte {
	absoluteMatchOffset := len(uncompressed) - matchOffset

	for i := absoluteMatchOffset; i < absoluteMatchOffset+totalMatchLength; i++ {
		uncompressed = append(uncompressed, uncompressed[i])
	}

	return uncompressed
}
