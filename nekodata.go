package main

import (
	"bytes"
	"fmt"
	"os"
)

type NekoDataType string

const (
	NekoDataTypeUnknown NekoDataType = "unknown"
	NekoDataTypeUnity   NekoDataType = "unity"
	NekoDataTypeLuac    NekoDataType = "luac"
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
	dataType        NekoDataType
}

func NewNekoData(data []byte) *NekoData {
	neko := &NekoData{
		data:            data,
		currentPosition: 0,
	}

	neko.dataType = determineNekoDataType(neko)

	neko.Reset()

	return neko
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

func (neko *NekoData) Reset() {
	neko.currentPosition = 0
}

func (neko *NekoData) DataType() NekoDataType {
	return neko.dataType
}

func (neko *NekoData) AllPatternIndices(bytePattern []byte) []int {
	var indices []int
	pos := 0

	for {
		relativeIndex := bytes.Index(neko.data[pos:], bytePattern)
		if relativeIndex == -1 {
			break
		}

		absoluteIndex := pos + relativeIndex

		indices = append(indices, absoluteIndex)
		pos = absoluteIndex + 1
	}

	return indices
}

func (neko *NekoData) Seek(position int) {
	neko.currentPosition = position
}

func (neko *NekoData) CurrentOffset() int {
	return neko.currentPosition
}

func (neko *NekoData) Slice(start int, end int) *NekoData {
	return &NekoData{
		data:            neko.data[start:end],
		currentPosition: 0,
		dataType:        neko.dataType,
	}
}

func (neko *NekoData) FullyRead() bool {
	return len(neko.data) <= neko.currentPosition
}

func (neko *NekoData) RemainingBytes() int {
	return len(neko.data) - neko.currentPosition
}

func determineNekoDataType(nd *NekoData) NekoDataType {
	header := uncompressHeader(nd, 1)

	if len(header) >= 5 && bytes.Compare(header[:5], luacFileHeader) == 0 {
		return NekoDataTypeLuac
	}

	if len(header) >= 7 && bytes.Compare(header[:7], unityFileHeader) == 0 {
		return NekoDataTypeUnity
	}

	return NekoDataTypeUnknown
}
