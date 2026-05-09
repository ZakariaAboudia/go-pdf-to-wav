package audio

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type wavHeader struct {
	ChunkID       [4]byte
	ChunkSize     uint32
	Format        [4]byte
	Subchunk1ID   [4]byte
	Subchunk1Size uint32
	AudioFormat   uint16
	NumChannels   uint16
	SampleRate    uint32
	ByteRate      uint32
	BlockAlign    uint16
	BitsPerSample uint16
	Subchunk2ID   [4]byte
	Subchunk2Size uint32
}

func chunkIndex(name string) int {
	name = strings.TrimSuffix(name, ".wav")
	parts := strings.Split(name, "_")
	n, _ := strconv.Atoi(parts[len(parts)-1])
	return n
}

func ListWavFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".wav" {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}

	sort.Slice(files, func(i, j int) bool {
		return chunkIndex(filepath.Base(files[i])) < chunkIndex(filepath.Base(files[j]))
	})

	return files, nil
}

func Combine(inputs []string, outputFile string) error {
	var allPCM []byte
	var hdr wavHeader

	for i, f := range inputs {
		r, err := os.Open(f)
		if err != nil {
			return fmt.Errorf("open %s: %w", f, err)
		}

		var h wavHeader
		if err := binary.Read(r, binary.LittleEndian, &h); err != nil {
			r.Close()
			return fmt.Errorf("read header %s: %w", f, err)
		}

		if i == 0 {
			hdr = h
		} else if h.SampleRate != hdr.SampleRate || h.NumChannels != hdr.NumChannels || h.BitsPerSample != hdr.BitsPerSample {
			r.Close()
			return fmt.Errorf("chunk %s format mismatch: got %dHz/%dch/%dbit, want %dHz/%dch/%dbit",
				f, h.SampleRate, h.NumChannels, h.BitsPerSample,
				hdr.SampleRate, hdr.NumChannels, hdr.BitsPerSample)
		}

		pcm := make([]byte, h.Subchunk2Size)
		if _, err := io.ReadFull(r, pcm); err != nil {
			r.Close()
			return fmt.Errorf("read pcm %s: %w", f, err)
		}
		r.Close()

		allPCM = append(allPCM, pcm...)
	}

	hdr.Subchunk2Size = uint32(len(allPCM))
	hdr.ChunkSize = 36 + hdr.Subchunk2Size

	out, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer out.Close()

	if err := binary.Write(out, binary.LittleEndian, hdr); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	_, err = out.Write(allPCM)
	return err
}
