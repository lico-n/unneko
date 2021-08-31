package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ExtractedFile interface {
	Data() []byte
	FilePath() string
}

func extractNekoData(inputPath string, outputPath string, keepOriginalLuacHeader bool) error {
	neko, err := loadNekoData(inputPath)
	if err != nil {
		return err
	}

	var extractedFiles []ExtractedFile

	switch neko.dataType {
	case NekoDataTypeLuac:
		extractedFiles, err = extractLuacFiles(neko, keepOriginalLuacHeader)
	case NekoDataTypeUnity:
		extractedFiles, err = extractUnityFiles(neko)
	default:
		return fmt.Errorf("unhandled neko data type %s", neko.dataType)
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

func saveExtractedFile(outputPath string, neko *NekoData, file ExtractedFile, fileIndex int) error {
	data := file.Data()

	filePath := file.FilePath()
	if filePath == "" {
		filePath = fmt.Sprintf("%d%s", fileIndex, getFileExtensionForNekoData(neko))
	}

	outputFilePath := filepath.Join(outputPath, filePath)

	outputDir := filepath.Dir(outputFilePath)
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("creating output dir %s: %v", outputDir, err)
	}

	if err := os.WriteFile(outputFilePath, data, os.ModePerm); err != nil {
		return fmt.Errorf("saving extracted file: %v", err)
	}

	return nil
}

func getFileExtensionForNekoData(neko *NekoData) string {
	switch neko.DataType() {
	case NekoDataTypeLuac:
		return ".luac"
	case NekoDataTypeUnity:
		return ".assetbundle"
	}

	return ""
}

func getNekoDataBaseFileName(inputPath string) string {
	nekoDataFileName := filepath.Base(inputPath)

	return strings.Split(nekoDataFileName, ".")[0]
}
