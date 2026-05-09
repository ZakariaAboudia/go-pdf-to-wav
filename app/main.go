package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"pdf-to-wav/internal/convert"
	"pdf-to-wav/internal/pdf"
	"pdf-to-wav/internal/tts"
)

func main() {
	fp := flag.String("filepath", "", "path to the pdf file")
	voice := flag.String("voice", "./piper/voices/en_US-libritts_r-medium.onnx", "path to piper voice model (.onnx)")
	workers := flag.Int("workers", max(1, runtime.NumCPU()/2), "number of parallel piper workers")
	flag.Parse()

	if *fp == "" {
		fmt.Fprintln(os.Stderr, "--filepath flag is required")
		os.Exit(1)
	}

	fnString := filepath.Base(strings.TrimSuffix(*fp, ".pdf"))
	txtFile := fnString + ".txt"

	if err := pdf.Extract(*fp, txtFile); err != nil {
		log.Fatal("read pdf:", err)
	}

	c := &convert.Converter{
		Piper: &tts.Piper{
			BinaryPath: "./piper/piper",
			ModelPath:  *voice,
		},
		ChunkSize:     20,
		MaxGoroutines: *workers,
	}

	if err := c.Run(txtFile, fnString+".wav"); err != nil {
		log.Fatal("convert:", err)
	}
}
