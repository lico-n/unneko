package main

import (
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

	var extractedFiles []*extractedFile

	if neko.ContainsLuac() {
		extractedFiles, err = extractLuacFiles(neko, keepOriginalLuacHeader)
	} else {
		extractedFiles, err = extractFiles(neko)
	}

	if err != nil {
		return err
	}

	nekoBaseFileName := getNekoDataBaseFileName(inputPath)

	outputPath = filepath.Join(outputPath, nekoBaseFileName)

	for i, v := range extractedFiles {
		if err := saveExtractedFile(outputPath, neko, v, i); err != nil {
			return err
		}
	}

	return nil
}

func extractFiles(neko *NekoData) ([]*extractedFile, error) {
	var extracted []*extractedFile

	for !neko.FullyRead() {
		if nextFileIsUnityFile(neko) {
			fileSize := readUnityFileSize(neko)

			uncompressed := uncompressNeko(neko, newMaxUncompressedSizeCompleteCond(int(fileSize)))
			extracted = append(extracted, &extractedFile{
				data:          uncompressed,
				fileExtension: ".assetbundle",
			})

			neko = neko.SliceFromCurrentPos()
			continue
		}

		if nextFileIsJSONObject(neko) {
			uncompressed := uncompressNeko(neko, newBracketCounterCompleteCond('{', '}'))
			extracted = append(extracted, &extractedFile{
				data: uncompressed,
				fileExtension: ".json",
			})
			neko = neko.SliceFromCurrentPos()
			continue
		}

		if nextFileIsJSONArray(neko) {
			uncompressed := uncompressNeko(neko, newBracketCounterCompleteCond('[', ']'))
			extracted = append(extracted, &extractedFile{
				data: uncompressed,
				fileExtension: ".json",
			})
			neko = neko.SliceFromCurrentPos()
			continue
		}


		if bytes := tryUncompressHeader(neko, 1); len(bytes) > 0 {
			fmt.Printf("stopped processing but there might be more data with file header\n%s\n", string(bytes))
		}

		break
	}

	return extracted, nil
}

func saveExtractedFile(outputPath string, neko *NekoData, file *extractedFile, fileIndex int) error {

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

	return nil
}

func getNekoDataBaseFileName(inputPath string) string {
	nekoDataFileName := filepath.Base(inputPath)

	return strings.Split(nekoDataFileName, ".")[0]
}
