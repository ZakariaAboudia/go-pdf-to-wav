package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"pdf-to-wav/internal/audio"
	"pdf-to-wav/internal/pdf"
	"pdf-to-wav/internal/tts"
)

const (
	chunkSize     = 20 // Number of lines to process at once
	maxGoroutines = 3  // Maximum number of goroutines to use
)



func textToSpeech(inputFile string, outputFile string) error {
	inFile, err := os.Open(inputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer inFile.Close()

	scanner := bufio.NewScanner(inFile)

	var wg sync.WaitGroup
	sem := make(chan struct{}, maxGoroutines)

	err = os.Mkdir("tts_chunks", 0775)

	if err != nil {
		log.Panic(err)
	}

	err = os.Mkdir("voice_chunks", 0775)

	if err != nil {
		log.Panic(err)
	}

	chunkCounter := 0

	for {
		chunk := make([]string, 0, chunkSize)
		for i := 0; i < chunkSize && scanner.Scan(); i++ {
			chunk = append(chunk, scanner.Text())
		}

		if len(chunk) == 0 {
			break
		}

		chunkFile := filepath.Join("tts_chunks", fmt.Sprintf("chunk_%d.txt", chunkCounter))
		outputWav := filepath.Join("voice_chunks", fmt.Sprintf("chunk_%d", chunkCounter))
		chunkCounter++

		wg.Add(1)
		sem <- struct{}{}
		go func(lines []string, chunkFile, outputWav string) {
			defer wg.Done()
			defer func() { <-sem }()

			text := strings.Join(lines, "\n")
			err := os.WriteFile(chunkFile, []byte(text), 0644)
			if err != nil {
				fmt.Printf("Error writing chunk file: %v\n", err)
				return
			}

			err = tts.Run(text, outputWav)
			if err != nil {
				fmt.Printf("Error converting chunk to wav: %v\n", err)
				return
			}
		}(chunk, chunkFile, outputWav)
	}

	wg.Wait()

	if err := scanner.Err(); err != nil {
		return err
	}

	audio.BuildWavList("voice_chunks", filepath.Join("voice_chunks", "voices.txt"))

	audio.Combine(filepath.Join("voice_chunks", "voices.txt"), outputFile)

	os.RemoveAll("tts_chunks")
	os.RemoveAll("voice_chunks")

	return err

	// command := fmt.Sprintf("echo %s | ./piper/piper --model ./piper/en_US-libritts_r-medium.onnx  --output_file %s", text, outputFilePath)
	// // fmt.Println(command)
	// cmd := exec.Command("bash", "-c", command)
	// stdout, err := cmd.Output()

	// if err != nil {
	// 	log.Println(err.Error())
	// 	return nil
	// }
	// log.Println(string(stdout))
	// return nil
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func main() {

	fp := flag.String("filepath", "", "path to the pdf file")

	flag.Parse()

	if !isFlagPassed("filepath") {
		fmt.Println("--filepath flag is required")
		os.Exit(0)
	}

	fnString := filepath.Base(strings.TrimSuffix(*fp, ".pdf"))
	txtFile := fnString + ".txt"

	if err := pdf.Extract(*fp, txtFile); err != nil {
		log.Fatal("read pdf:", err)
	}

	err := textToSpeech(txtFile, fnString+".wav")
	if err != nil {
		log.Fatal("could not convert to mp3 ", err)
	}

	// chunks := chunkLines(lines, chunkSize)
	// for i, chunk := range chunks {
	// 	text := strings.Join(chunk, " ")
	// 	outputFilePath := fmt.Sprintf("audio_chunk_%d.wav", i+1)
	// 	err := textToSpeech(text, outputFilePath)
	// 	if err != nil {
	// 		fmt.Println("Error creating audio file:", err)
	// 		return
	// 	}
	// 	fmt.Println("Audio file created:", outputFilePath)
	// }
}

// find *.mp4 | sed 's:\ :\\\ :g'| sed 's/^/file /' > fl.txt; ffmpeg -f concat -i fl.txt -c copy output.mp4; rm fl.txt
