package unneko

import (
	"bytes"
	"errors"
)

var pixelnekoFilesystemHeader = []byte("pixelneko filesystem")

func isPixelnekoFileHeader(bb []byte) bool {
	return bytes.Compare(bb, pixelnekoFilesystemHeader) == 0
}

func newNekoData(file []byte, isPatchFile bool) (*nekoData, error) {
	if len(file) < 0x19  || !isPixelnekoFileHeader(file[:0x14]) {
		return nil, errors.New("invalid nekoData file header")
	}

	return &nekoData{
		data:            file[0x19:],
		currentPosition: 0,
		isPatch:         isPatchFile,
	}, nil
}

type nekoData struct {
	data            []byte
	currentPosition int
	isPatch         bool
}

func (neko *nekoData) readBytes(size int) []byte {
	if size == 0 {
		return nil
	}

	endPosition := neko.currentPosition + size
	if len(neko.data) < endPosition {
		endPosition = len(neko.data)
	}

	readBytes := neko.data[neko.currentPosition:endPosition]

	neko.currentPosition = endPosition

	return readBytes
}

func (neko *nekoData) readByte() byte {
	readByte := neko.data[neko.currentPosition]
	neko.currentPosition++
	return readByte
}

func (neko *nekoData) seek(position int) {
	neko.currentPosition = position
}

func (neko *nekoData) currentOffset() int {
	return neko.currentPosition
}

func (neko *nekoData) index(sep []byte) int {
	index := bytes.Index(neko.data[neko.currentPosition:], sep)
	if index == -1 {
		return -1
	}
	return neko.currentPosition + index
}

func (neko *nekoData) fullyRead() bool {
	return len(neko.data) <= neko.currentPosition
}
