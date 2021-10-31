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

	extractedChan := extractFiles(neko, keepOriginalLuacHeader)

	extractedChan = restoreFileNames(extractedChan)

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

var anotherFile = 0
func extractFiles(neko *NekoData, keepOriginalLuacHeader bool) chan *extractedFile {
	extractedChan := make(chan *extractedFile, 1)

	totalOffset := 0x18

	go func() {
		defer close(extractedChan)

		for !neko.FullyRead() {
			anotherFile++

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

			file := extractPlainFile(neko)
			extractedChan <- file
			totalOffset += neko.CurrentOffset()
			neko = neko.SliceFromCurrentPos()

		}

		//fmt.Println(anotherFile)
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
