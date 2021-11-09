package main

import (
	"bytes"
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

func extractNekoData(inputPath string, outputPath string, keepOriginalLuacHeader bool) error {
	neko, err := loadNekoData(inputPath)
	if err != nil {
		return err
	}

	checksumFile := findChecksumFile(neko)

	if checksumFile == nil {
		return fmt.Errorf("unable to find checksum file")
	}
	neko.Seek(0)

	extractedChan := extractFiles(neko, checksumFile, keepOriginalLuacHeader)

	extractedChan = restoreFileNames(checksumFile, extractedChan)

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


func extractFiles(neko *NekoData, checksumFile *ChecksumFile, keepOriginalLuacHeader bool) chan *extractedFile {
	extractedChan := make(chan *extractedFile, 1)

	csumCond := newChecksumCompleteCond(checksumFile)

	go func() {
		defer close(extractedChan)

		for !neko.FullyRead() {

			startOffset := neko.CurrentOffset()
			headerBytes := tryUncompressHeader(neko, 1)
			if len(headerBytes) == 0 {
				break
			}

			neko.Seek(startOffset)

			if len(headerBytes) >= 7 && string(headerBytes[:7]) == "UnityFS" {
				file := extractUnityFile(neko)
				extractedChan <- file
				neko = neko.SliceFromCurrentPos()
				continue
			}

			if headerBytes[0] == '{' {
				file := extractJSONObjectFile(neko)
				extractedChan <- file
				neko = neko.SliceFromCurrentPos()
				continue
			}

			if headerBytes[0] == '[' {
				file := extractJSONArrayFile(neko)
				extractedChan <- file
				neko = neko.SliceFromCurrentPos()
				continue
			}

			if len(headerBytes) >= 5 && bytes.Compare(headerBytes[:5], luacFileHeader) == 0 {
				file := extractLuacFile(neko, keepOriginalLuacHeader)
				extractedChan <- file
				neko = neko.SliceFromCurrentPos()
				continue
			}

			file := extractPlainFile(neko, csumCond)
			extractedChan <- file
			neko = neko.SliceFromCurrentPos()

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
