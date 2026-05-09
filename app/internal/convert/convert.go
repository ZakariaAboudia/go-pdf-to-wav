package convert

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"pdf-to-wav/internal/audio"
	"pdf-to-wav/internal/tts"
)

const (
	chunkSize     = 20
	maxGoroutines = 3
)

func Run(inputFile string, outputFile string) error {
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
}
