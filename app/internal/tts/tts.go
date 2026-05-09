package tts

import (
	"os"
	"os/exec"
	"strings"
)

type Piper struct {
	BinaryPath string
	ModelPath  string
}

func (p *Piper) Run(text, outputFile string) error {
	cmd := exec.Command(
		p.BinaryPath,
		"--model", p.ModelPath,
		"--output_file", outputFile+".wav",
	)
	cmd.Stdin = strings.NewReader(text)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
