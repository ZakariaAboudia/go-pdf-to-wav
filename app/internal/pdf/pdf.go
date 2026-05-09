package pdf

import (
	"bufio"
	"fmt"
	"os"

	"github.com/cheggaaa/go-poppler"
)

func Extract(src, dst string) error {
	doc, err := poppler.Open(src)
	if err != nil {
		return fmt.Errorf("open pdf: %w", err)
	}
	defer doc.Close()

	file, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	defer w.Flush()

	for i := range doc.GetNPages() {
		if _, err := w.WriteString(doc.GetPage(i).Text()); err != nil {
			return fmt.Errorf("write page %d: %w", i, err)
		}
	}
	return nil
}
