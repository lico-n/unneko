package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	return saveExtractedFiles(outputPath, extractedChan)
}

func extractFiles(neko *NekoData, keepOriginalLuacHeader bool) (chan *extractedFile) {
	extractedChan := make(chan *extractedFile, 1)

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

			fmt.Printf("stopped processing but there might be more data with file header\n%s\n", string(headerBytes))
			break
		}
	}()

	return extractedChan
}



func saveExtractedFiles(outputPath string, extractedChan chan *extractedFile) error {
	fileIndex := 0
	for file := range extractedChan {
		fileIndex++

		filePath := file.filePath
		if filePath == "" {
			filePath = fmt.Sprintf("%d%s", fileIndex, file.fileExtension)
		}

		outputFilePath := filepath.Join(outputPath, filePath)

		outputDir := filepath.Dir(outputFilePath)
		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			return fmt.Errorf("creating output dir %s: %v", outputDir, err)
		}

		if err := os.WriteFile(outputFilePath, file.data, os.ModePerm); err != nil {
			return fmt.Errorf("saving extracted file: %v", err)
		}
	}

	return nil
}

func getNekoDataBaseFileName(inputPath string) string {
	nekoDataFileName := filepath.Base(inputPath)

	return strings.Split(nekoDataFileName, ".")[0]
}
