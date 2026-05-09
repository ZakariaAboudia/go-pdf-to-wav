package tts

import (
	"os"
	"os/exec"
	"strings"
)

func Run(text, outputFile string) error {
	cmd := exec.Command(
		"./piper/piper",
		"--model", "./piper/voices/en_US-libritts_r-medium.onnx",
		"--output_file", outputFile+".wav",
	)
	cmd.Stdin = strings.NewReader(text)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
