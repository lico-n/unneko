package main

func extractUncompressed(neko *NekoData, csumCond *checksumCompleteCond) *extractedFile {
	possibleFileSizes := csumCond.PossibleFileSizes()

	var data []byte

	for _, v := range possibleFileSizes {
		toRead := v - len(data)
		newData := neko.ReadBytes(toRead)
		data = append(data, newData...)

		if csumCond.Complete(neko, data) {
			csumCond.MarkAsFound(data)
			break
		}
	}

	return &extractedFile{
		data: data,
	}
}

func containsChecksum(checksumFiles []*ChecksumFile, checksum uint32) bool {
	for _, v := range checksumFiles {
		for _, f := range v.Files {
			if f.Crc32 == checksum {
				return true
			}
		}
	}

	return false
}
