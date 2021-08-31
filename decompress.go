package main

func uncompressHeader(neko *NekoData, numberOfSeq int) []byte {
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

func uncompressNeko(neko *NekoData, maxBytes int) []byte {
	var uncompressed []byte

	for !neko.FullyRead() {
		maxByteRemaining := maxBytes - len(uncompressed)
		if maxBytes > 0 && maxByteRemaining <= 0 {
			break
		}

		uncompressedBlock := uncompressNekoBlock(neko, maxByteRemaining)
		uncompressed = append(uncompressed, uncompressedBlock...)

	}

	return uncompressed
}

func uncompressNekoBlock(neko *NekoData, maxBytes int) []byte {
	var uncompressed []byte

	for !neko.FullyRead() {

		token := readToken(neko)

		literals := neko.ReadBytes(token.nrOfLiterals)
		uncompressed = append(uncompressed, literals...)

		if len(uncompressed) == 0x8000 || neko.FullyRead() {
			break
		}

		if maxBytes > 0 && len(uncompressed) >= maxBytes {
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
