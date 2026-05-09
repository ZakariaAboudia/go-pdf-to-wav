package convert

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"pdf-to-wav/internal/audio"
	"pdf-to-wav/internal/tts"
)

type Converter struct {
	Piper         *tts.Piper
	ChunkSize     int
	MaxGoroutines int
}

func (c *Converter) Run(inputFile, outputFile string) error {
	inFile, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("open input: %w", err)
	}
	defer inFile.Close()

	if err := os.MkdirAll("voice_chunks", 0775); err != nil {
		return fmt.Errorf("mkdir voice_chunks: %w", err)
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, c.MaxGoroutines)
	errCh := make(chan error, 1)

	scanner := bufio.NewScanner(inFile)
	chunkCounter := 0

	for {
		chunk := make([]string, 0, c.ChunkSize)
		for i := 0; i < c.ChunkSize && scanner.Scan(); i++ {
			chunk = append(chunk, scanner.Text())
		}

		if len(chunk) == 0 {
			break
		}

		outputWav := filepath.Join("voice_chunks", fmt.Sprintf("chunk_%d", chunkCounter))
		chunkCounter++

		wg.Add(1)
		sem <- struct{}{}
		go func(lines []string, outputWav string) {
			defer wg.Done()
			defer func() { <-sem }()

			text := strings.Join(lines, "\n")
			if err := c.Piper.Run(text, outputWav); err != nil {
				select {
				case errCh <- fmt.Errorf("piper %s: %w", outputWav, err):
				default:
				}
			}
		}(chunk, outputWav)
	}

	wg.Wait()

	if err := scanner.Err(); err != nil {
		return err
	}

	select {
	case err := <-errCh:
		return err
	default:
	}

	wavFiles, err := audio.ListWavFiles("voice_chunks")
	if err != nil {
		return fmt.Errorf("list wav files: %w", err)
	}

	if err := audio.Combine(wavFiles, outputFile); err != nil {
		return fmt.Errorf("combine wav: %w", err)
	}

	os.RemoveAll("voice_chunks")

	return nil
}
