package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	command := fmt.Sprintf("echo \"$(< %s )\" | ./piper/piper --model ./piper/voices/en_US-libritts_r-medium.onnx  --output_file %v.wav", inputFile, outputFile)
	cmd := exec.Command("bash", "-c", command)

	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func prepareWavIndexFile() error {

	command := "find /app/voice_chunks/*.wav | sed 's:\\ :\\\\ :g'| sed 's/^/file /' > /app/voice_chunks/voices.txt"
	cmd := exec.Command("bash", "-c", command)

	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func combineMP3Files(outputFile string) error {
	command := fmt.Sprintf("ffmpeg -f concat -safe 0 -i /app/voice_chunks/voices.txt -c copy %s", outputFile)
	cmd := exec.Command("bash", "-c", command)

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

		chunkFile := filepath.Join("/app/tts_chunks", fmt.Sprintf("chunk_%d.txt", chunkCounter))
		chunkCounter++

		wg.Add(1)
		sem <- struct{}{}
		go func(lines []string, chunkFile string, chunkCounter int) {
			defer wg.Done()
			defer func() { <-sem }()

			// Write chunk to file
			err := os.WriteFile(chunkFile, []byte(joinLines(lines)), 0644)
			if err != nil {
				fmt.Printf("Error writing chunk file: %v\n", err)
				return
			}

			// Convert chunk to MP3 using FFmpeg
			outputMP3 := fmt.Sprintf("/app/voice_chunks/chunk_%d", chunkCounter)
			err = FFmpeg(chunkFile, outputMP3)
			if err != nil {
				fmt.Printf("Error converting chunk to MP3: %v\n", err)
				return
			}
		}(chunk, chunkFile, chunkCounter)
	}

	wg.Wait()

	if err := scanner.Err(); err != nil {
		return err
	}

	prepareWavIndexFile()

	combineMP3Files(outputFile)

	os.RemoveAll("/app/tts_chunks")
	os.RemoveAll("/app/voice_chunks")

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

	fnString := filepath.Base(strings.Replace(*fp, ".pdf", "", -1))
	fn := fmt.Sprintf("%s.txt", fnString)
	textFile, err := readFileLines(*fp, fn)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	err = textToSpeech(textFile, filepath.Join("", fmt.Sprintf("%s.wav", fnString)))
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
