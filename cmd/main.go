package main

import (
	"flag"
	"fmt"
	"github.com/lico-n/unneko"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func main() {
	outputDir := flag.String("o", "./output", "output directory")

	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "unneko v1.16.0 by Lico#6969")
		fmt.Fprintln(os.Stderr, "Usage: unneko <flags> input-file")
		flag.PrintDefaults()
		os.Exit(-1)
	}

	inputFiles, err := getInputFiles(args[0])
	if err != nil {
		fmt.Printf("unable to read input files: %v", err)
		os.Exit(-1)
	}

	for _, v := range inputFiles {
		fmt.Printf("extracting: %s\n", v)
		if err := performExtraction(v, *outputDir); err != nil {
			fmt.Printf("failed to extract nekodata: %v", err)
			os.Exit(-1)
		}
	}
}

func performExtraction(inputFilePath string, outputDir string) error {
	r, err := unneko.NewReaderFromFile(inputFilePath)
	if err != nil {
		return err
	}

	extractedChan := make(chan *unneko.ExtractedFile, 1)

	// run extraction thread
	go func() {
		defer close(extractedChan)
		for r.HasNext() {
			file, err := r.Next()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				break
			}

			extractedChan <- file
		}
	}()

	// run file storing threads
	numOfSaveWorker := 16
	wg := &sync.WaitGroup{}
	wg.Add(numOfSaveWorker)
	for i := 0; i < numOfSaveWorker; i++ {
		go startFileSavingWorker(wg, inputFilePath, outputDir, extractedChan)
	}

	wg.Wait()
	return nil
}

func startFileSavingWorker(wg *sync.WaitGroup, inputFile string, outputPath string, ch <-chan *unneko.ExtractedFile) {
	defer wg.Done()
	for file := range ch {

		outputFilePath := filepath.Join(outputPath, getFileNameWithoutExt(inputFile), file.Path())

		outputDir := filepath.Dir(outputFilePath)
		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			fmt.Fprintln(os.Stderr, fmt.Errorf("creating output dir %s: %v", outputDir, err))
			continue
		}

		if err := os.WriteFile(outputFilePath, file.Data(), os.ModePerm); err != nil {
			fmt.Fprintln(os.Stderr, fmt.Errorf("saving extracted file %s: %v", outputFilePath, err))
			continue
		}
	}
}

func getFileNameWithoutExt(inputFile string) string {
	baseFile := filepath.Base(inputFile)
	splitted := strings.Split(baseFile, ".")
	return splitted[0]
}

func getInputFiles(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat input path: %v", err)
	}
	if !info.IsDir() {
		return []string{path}, nil
	}

	var nekoFiles []string

	dirFiles, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("stat input dir: %v", err)
	}

	for _, v := range dirFiles {
		if v.IsDir() {
			continue
		}

		if strings.HasSuffix(v.Name(), ".nekodata") {
			nekoPath := filepath.Join(path, v.Name())
			nekoFiles = append(nekoFiles, nekoPath)
		}
	}

	return nekoFiles, nil
}
