package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	outputDir := flag.String("o", "./output", "output directory")

	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "unneko v1.10.0 by Lico#6969")
		fmt.Fprintln(os.Stderr, "Usage: unneko <flags> input-file")
		flag.PrintDefaults()
		os.Exit(-1)
	}

	inputFile := args[0]
	if err := extractNekoData(inputFile, *outputDir); err != nil {
		fmt.Printf("failed to extract nekodata: %v", err)
		os.Exit(-1)
	}
}
