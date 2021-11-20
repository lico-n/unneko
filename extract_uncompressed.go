package main

import "strings"

// isUncompressedFile returns whether the next file in NekoData is uncompressed
// for now this only has been observed for NekoData within NekoData as those are already compressed
// if the next few bytes looks like a NekoData header, then it's probably uncompressed
// a compressed NekoData would start with a lz4 tokenbyte instead
func isUncompressedFile(neko *NekoData) bool {
	startOffset := neko.CurrentOffset()

	isUncompressed := strings.HasPrefix(string(neko.ReadBytes(19)), "pixelneko")
	neko.Seek(startOffset)

	return isUncompressed
}

// extractUncompressed uses the checksum file to try to extract an uncompressed file at the current NekoData position.
// With the checksum file we know the filesize of every file within NekoData.
// We only need to calculate the checksum at every possible filesize and if it matches one of the
// checksums, we will stop and count that as successfully extracted file.
func extractUncompressed(neko *NekoData, csumCond *checksumCompleteCond) *extractedFile {
	possibleFileSizes := csumCond.PossibleFileSizes()

	var data []byte

	for _, v := range possibleFileSizes {
		toRead := v - len(data)
		newData := neko.ReadBytes(toRead)
		data = append(data, newData...)

		if csumCond.Complete(neko, data) {
			break
		}
	}

	csumCond.MarkAsFound(data)

	return &extractedFile{
		data: data,
	}
}