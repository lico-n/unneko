package main

import (
	"flag"
	"fmt"
	"github.com/lico-n/unneko"
	"os"
	"path/filepath"
	"sync"
)

func main() {
	outputDir := flag.String("o", "./output", "output directory")

	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "unneko v1.14.0 by Lico#6969")
		fmt.Fprintln(os.Stderr, "Usage: unneko <flags> input-file")
		flag.PrintDefaults()
		os.Exit(-1)
	}

	inputFile := args[0]
	if err := performExtraction(inputFile, *outputDir); err != nil {
		fmt.Printf("failed to extract nekodata: %v", err)
		os.Exit(-1)
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
		go startFileSavingWorker(wg, outputDir, extractedChan)
	}

	wg.Wait()
	return nil
}

func startFileSavingWorker(wg *sync.WaitGroup, outputPath string, ch <-chan *unneko.ExtractedFile) {
	defer wg.Done()
	for file := range ch {
		outputFilePath := filepath.Join(outputPath, file.Path())

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
