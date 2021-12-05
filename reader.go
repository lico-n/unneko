package unneko

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type ExtractedFile struct {
	fileData []byte
	filePath string
}

func (f *ExtractedFile) Data() []byte {
	return f.fileData
}

func (f *ExtractedFile) Path() string {
	return f.filePath
}

type Reader struct {
	neko  *nekoData
	cCond *checksumCompleteCond
	namer *fileNamer
}

func NewReaderFromFile(file string) (*Reader, error) {
	inputFile, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("unable to read input file: %v", err)
	}

	isPatch := strings.HasSuffix(strings.ToLower(file), ".patch.nekodata")

	return NewReader(inputFile, isPatch)
}

func NewReader(data []byte, isPatchFile bool) (*Reader, error) {
	neko, err := newNekoData(data, isPatchFile)
	if err != nil {
		return nil, err
	}

	checksumFiles := findChecksumFiles(neko)
	if len(checksumFiles) == 0 {
		return nil, fmt.Errorf("unable to find checksum file")
	}

	return &Reader{
		neko:  neko,
		cCond: newChecksumCompleteCond(checksumFiles),
		namer: newFileNamer(checksumFiles),
	}, nil
}

func (r *Reader) HasNext() bool {
	return !r.cCond.FoundAll() && !r.neko.fullyRead()
}

func (r *Reader) Next() (*ExtractedFile, error) {
	if isUncompressedFile(r.neko) {
		data := extractUncompressed(r.neko, r.cCond)
		return r.newExtractedFile(data), nil
	}

	headerBytes := tryUncompressHeader(r.neko, 1)
	if len(headerBytes) == 0 {
		return nil, errors.New("file not extractable")
	}

	data := uncompressNeko(r.neko, r.cCond)

	return r.newExtractedFile(data), nil
}

func (r *Reader) newExtractedFile(data []byte) *ExtractedFile {
	r.cCond.MarkAsFound(data)

	return &ExtractedFile{
		fileData: data,
		filePath: r.namer.GetName(data),
	}
}
