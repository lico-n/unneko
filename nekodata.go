package main

import (
	"bytes"
	"fmt"
	"os"
)


func loadNekoData(path string) (*NekoData, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loading nekodata file at %s: %v", path, err)
	}

	fsHeader := string(file[:0x14])
	if fsHeader != "pixelneko filesystem" || len(file) < 0x19 {
		return nil, fmt.Errorf(" %s not a nekodata file", path)
	}

	return NewNekoData(file[0x19:]), nil
}

type NekoData struct {
	data            []byte
	currentPosition int
}

func NewNekoData(data []byte) *NekoData {
	return &NekoData{
		data:            data,
		currentPosition: 0,
	}
}

func (neko *NekoData) ReadBytes(size int) []byte {
	if size == 0 {
		return nil
	}

	endPosition := neko.currentPosition + size
	readBytes := neko.data[neko.currentPosition:endPosition]

	neko.currentPosition = endPosition

	return readBytes
}

func (neko *NekoData) ReadByte() byte {
	readByte := neko.data[neko.currentPosition]
	neko.currentPosition++

	return readByte
}

func (neko *NekoData) StillContains(bytePattern []byte) bool {
	return bytes.Index(neko.data[neko.currentPosition:], bytePattern) != -1
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
	}
}

func (neko *NekoData) Index(sep []byte) int {
	index :=  bytes.Index(neko.data[neko.currentPosition:], sep)
	if index == -1 {
		return -1
	}
	return neko.currentPosition+index
}

func (neko *NekoData) FullyRead() bool {
	return len(neko.data) <= neko.currentPosition
}
