package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type extractedFile struct {
	data          []byte
	filePath      string
	fileExtension string
}

func extractNekoData(inputPath string, outputPath string) error {
	neko, err := loadNekoData(inputPath)
	if err != nil {
		return err
	}

	checksumFiles := findChecksumFiles(neko)
	if len(checksumFiles) == 0 {
		return fmt.Errorf("unable to find checksum file")
	}

	neko.Seek(0)

	extractedChan := extractFiles(neko, checksumFiles)

	extractedChan = restoreFileNames(checksumFiles, extractedChan)

	nekoBaseFileName := getNekoDataBaseFileName(inputPath)
	outputPath = filepath.Join(outputPath, nekoBaseFileName)

	numOfSaveWorker := 16

	wg := &sync.WaitGroup{}
	wg.Add(numOfSaveWorker)

	for i := 0; i < numOfSaveWorker; i++ {
		startFileSavingWorker(wg, outputPath, extractedChan)
	}

	wg.Wait()
	return nil
}

func extractFiles(neko *NekoData, checksumFiles []*ChecksumFile) chan *extractedFile {
	extractedChan := make(chan *extractedFile, 1)

	csumCond := newChecksumCompleteCond(checksumFiles)

	go func() {
		defer close(extractedChan)

		for !neko.FullyRead() && !csumCond.FoundAll() {
			startOffset := neko.CurrentOffset()

			if isUncompressedFile(neko) {
				file := extractUncompressed(neko, csumCond)
				extractedChan <- file
				neko = neko.SliceFromCurrentPos()
				continue
			}

			neko.Seek(startOffset)
			headerBytes := tryUncompressHeader(neko, 1)
			if len(headerBytes) == 0 {
				break
			}

			neko.Seek(startOffset)
			file := extractWithChecksum(neko, csumCond)
			extractedChan <- file
			neko = neko.SliceFromCurrentPos()
		}

		if !csumCond.FoundAll() {
			fmt.Fprint(os.Stderr, "extract finished but not all files were found\n")
		}
	}()

	return extractedChan
}

func startFileSavingWorker(wg *sync.WaitGroup, outputPath string, extractedChan chan *extractedFile) {
	go func() {
		defer wg.Done()
		for file := range extractedChan {
			outputFilePath := filepath.Join(outputPath, file.filePath)

			outputDir := filepath.Dir(outputFilePath)
			if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
				panic(fmt.Errorf("creating output dir %s: %v", outputDir, err))
			}

			if err := os.WriteFile(outputFilePath, file.data, os.ModePerm); err != nil {
				panic(fmt.Errorf("saving extracted file: %v", err))
			}
		}
	}()
}

func getNekoDataBaseFileName(inputPath string) string {
	nekoDataFileName := filepath.Base(inputPath)

	return strings.Split(nekoDataFileName, ".")[0]
}
