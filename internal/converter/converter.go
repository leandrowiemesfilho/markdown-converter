package converter

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/leandrowiemesfilho/markdown-converter/internal/pdf"
)

type Converter interface {
	ToMarkdown(inputPath, outputPath string) error
}

// FileType represents supported file types
type FileType string

const (
	PDF  FileType = "pdf"
	DOCX FileType = "docx"
	XLSX FileType = "xlsx"
	PPTX FileType = "pptx"
)

// GetConverter returns the appropriate converter based on file extension
func GetConverter(filePath string, assetsDir string) (Converter, FileType, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".pdf":
		return &pdf.Converter{AssetsDir: assetsDir}, PDF, nil
	default:
		return nil, "", fmt.Errorf("unsupported file type: %s", ext)
	}
}
