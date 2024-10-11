package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/cheggaaa/go-poppler"
)

const (
	chunkSize     = 20 // Number of lines to process at once
	maxGoroutines = 3  // Maximum number of goroutines to use
)

func readFileLines(filePath string, outputFile string) (string, error) {
	pdf, err := poppler.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer pdf.Close()

	file, err := os.Create(outputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	defer w.Flush()

	pages := pdf.GetNPages()
	for i := range pages {
		p := pdf.GetPage(i)
		c := p.Text()
		_, err := w.WriteString(c)
		if err != nil {
			log.Fatal(err)
		}
	}

	// file, err = os.Open("example.txt")
	// if err != nil {
	// 	return nil, err
	// }
	// defer file.Close()

	// var lines []string
	// scanner := bufio.NewScanner(file)
	// for scanner.Scan() {
	// 	lines = append(lines, scanner.Text())
	// }
	// if err := scanner.Err(); err != nil {
	// 	return nil, err
	// }

	return outputFile, nil
}

func joinLines(lines []string) string {
	var result string
	for _, line := range lines {
		result += line + "\n"
	}

	return result
}

func FFmpeg(inputFile, outputFile string) error {
	command := fmt.Sprintf("echo \"$(< %s )\" | ./piper/piper --model ./piper/en_US-libritts_r-medium.onnx  --output_file ./%v.wav", inputFile, outputFile)
	cmd := exec.Command("bash", "-c", command)

	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func joinFilePaths(paths []string) string {
	return filepath.Join(paths...)
}

func combineMP3Files(inputFiles []string, outputFile string) error {
	args := []string{"-i", fmt.Sprintf("concat:%s", joinFilePaths(inputFiles)), "-acodec", "copy", outputFile}
	cmd := exec.Command("ffmpeg", args...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func textToSpeech(inputFile string, outputFile string) error {
	inFile, err := os.Open(inputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer inFile.Close()

	scanner := bufio.NewScanner(inFile)

	var wg sync.WaitGroup
	sem := make(chan struct{}, maxGoroutines)

	tempDir, err := os.MkdirTemp("/app/", "tts_chunks")

	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	var chunkFiles []string
	chunkCounter := 0

	for {
		chunk := make([]string, 0, chunkSize)
		for i := 0; i < chunkSize && scanner.Scan(); i++ {
			chunk = append(chunk, scanner.Text())
		}

		if len(chunk) == 0 {
			break
		}

		chunkFile := filepath.Join(fmt.Sprintf("chunk_%d.txt", chunkCounter))
		chunkCounter++

		wg.Add(1)
		sem <- struct{}{}
		go func(lines []string, chunkFile string) {
			defer wg.Done()
			defer func() { <-sem }()

			// Write chunk to file
			err := os.WriteFile(chunkFile, []byte(joinLines(lines)), 0644)
			if err != nil {
				fmt.Printf("Error writing chunk file: %v\n", err)
				return
			}

			// Convert chunk to MP3 using FFmpeg
			outputMP3 := chunkFile
			err = FFmpeg(chunkFile, outputMP3)
			if err != nil {
				fmt.Printf("Error converting chunk to MP3: %v\n", err)
				return
			}

			chunkFiles = append(chunkFiles, outputMP3)
		}(chunk, chunkFile)
	}

	wg.Wait()

	if err := scanner.Err(); err != nil {
		return err
	}

	return err

	// return combineMP3Files(chunkFiles, outputFile)

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

func main() {
	filePath := "js-dummies.pdf"
	textFile, err := readFileLines(filePath, "example.txt")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	err = textToSpeech(textFile, filepath.Join("", "output.mp3"))
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
