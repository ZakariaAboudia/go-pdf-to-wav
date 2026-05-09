package tts

import (
	"os"
	"os/exec"
	"strings"
)

func Run(text, outputFile, model string) error {
	cmd := exec.Command(
		"./piper/piper",
		"--model", model,
		"--output_file", outputFile+".wav",
	)
	cmd.Stdin = strings.NewReader(text)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
