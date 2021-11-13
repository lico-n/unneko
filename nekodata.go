package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

var pixelnekoFilesystemHeader = []byte("pixelneko filesystem")

func isPixelnekoFileHeader(bb []byte) bool {
	return bytes.Compare(bb, pixelnekoFilesystemHeader) == 0
}

func loadNekoData(path string) (*NekoData, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loading nekodata file at %s: %v", path, err)
	}

	if len(file) < 0x19  || !isPixelnekoFileHeader(file[:0x14]) {
		return nil, fmt.Errorf(" %s not a nekodata file", path)
	}

	isPatch := strings.HasSuffix(strings.ToLower(path), ".patch.nekodata")

	return NewNekoData(file[0x19:], isPatch), nil
}

type NekoData struct {
	data            []byte
	currentPosition int
	isPatch         bool
}

func NewNekoData(data []byte, isPatch bool) *NekoData {
	return &NekoData{
		data:            data,
		currentPosition: 0,
		isPatch:         isPatch,
	}
}

func (neko *NekoData) ReadBytes(size int) []byte {
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

func (neko *NekoData) ReadByte() byte {
	readByte := neko.data[neko.currentPosition]
	neko.currentPosition++

	return readByte
}

func (neko *NekoData) Seek(position int) {
	neko.currentPosition = position
}

func (neko *NekoData) CurrentOffset() int {
	return neko.currentPosition
}

func (neko *NekoData) SliceFromCurrentPos() *NekoData {
	return &NekoData{
		data:            neko.data[neko.currentPosition:],
		currentPosition: 0,
		isPatch:         neko.isPatch,
	}
}

func (neko *NekoData) Index(sep []byte) int {
	index := bytes.Index(neko.data[neko.currentPosition:], sep)
	if index == -1 {
		return -1
	}
	return neko.currentPosition + index
}

func (neko *NekoData) FullyRead() bool {
	return len(neko.data) <= neko.currentPosition
}
