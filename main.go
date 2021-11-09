package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	outputDir := flag.String("o", "./output", "output directory")

	keepOriginalLuacHeaders := flag.Bool("original-luac-header",
		false,
		"keep original luac header instead of replacing them with decompilable ones",
	)

	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "unneko v1.8.3 by Lico#6969")
		fmt.Fprintln(os.Stderr, "Usage: unneko <flags> input-file")
		flag.PrintDefaults()
		os.Exit(-1)
	}

	inputFile := args[0]
	if err := extractNekoData(inputFile, *outputDir, *keepOriginalLuacHeaders); err != nil {
		fmt.Printf("failed to extract nekodata: %v", err)
		os.Exit(-1)
	}
}
