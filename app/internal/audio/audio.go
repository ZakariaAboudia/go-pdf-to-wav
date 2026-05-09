package audio

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func chunkIndex(name string) int {
	name = strings.TrimSuffix(name, ".wav")
	parts := strings.Split(name, "_")
	n, _ := strconv.Atoi(parts[len(parts)-1])
	return n
}

func BuildWavList(dir, listFile string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var wavFiles []string
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".wav" {
			wavFiles = append(wavFiles, filepath.Join(dir, e.Name()))
		}
	}

	sort.Slice(wavFiles, func(i, j int) bool {
		return chunkIndex(filepath.Base(wavFiles[i])) < chunkIndex(filepath.Base(wavFiles[j]))
	})

	var sb strings.Builder
	for _, f := range wavFiles {
		sb.WriteString("file " + filepath.Base(f) + "\n")
	}
	return os.WriteFile(listFile, []byte(sb.String()), 0644)
}

func Combine(listFile, outputFile string) error {
	cmd := exec.Command(
		"ffmpeg", "-f", "concat", "-safe", "0",
		"-i", listFile,
		"-c", "copy", outputFile,
	)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
