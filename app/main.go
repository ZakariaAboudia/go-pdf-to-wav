package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"pdf-to-wav/internal/convert"
	"pdf-to-wav/internal/pdf"
)

func main() {
	fp := flag.String("filepath", "", "path to the pdf file")
	voice := flag.String("voice", "./piper/voices/en_US-libritts_r-medium.onnx", "path to piper voice model (.onnx)")
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

	if err := convert.Run(txtFile, fnString+".wav", *voice); err != nil {
		log.Fatal("convert:", err)
	}
}
