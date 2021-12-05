package unneko

import "strings"

// isUncompressedFile returns whether the next file in nekoData is uncompressed
// for now this only has been observed for nekoData within nekoData as those are already compressed
// if the next few bytes looks like a nekoData header, then it's probably uncompressed
// a compressed nekoData would start with a lz4 tokenbyte instead
func isUncompressedFile(neko *nekoData) bool {
	startOffset := neko.currentOffset()
	defer func(){
		neko.seek(startOffset)
	}()

	return strings.HasPrefix(string(neko.readBytes(19)), "pixelneko")
}

// extractUncompressed uses the checksum file to try to extract an uncompressed file at the current nekoData position.
// With the checksum file we know the filesize of every file within nekoData.
// We only need to calculate the checksum at every possible filesize and if it matches one of the
// checksums, we will stop and count that as successfully extracted file.
func extractUncompressed(neko *nekoData, csumCond *checksumCompleteCond) []byte {
	possibleFileSizes := csumCond.PossibleFileSizes()

	var data []byte

	for _, v := range possibleFileSizes {
		toRead := v - len(data)
		newData := neko.readBytes(toRead)
		data = append(data, newData...)

		if csumCond.Complete(neko, data) {
			break
		}
	}

	return data
}