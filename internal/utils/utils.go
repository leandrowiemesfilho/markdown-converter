package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(dirPath string) error {
	return os.MkdirAll(dirPath, 0755)
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// GetOutputPath generates an output path based on input path and options
func GetOutputPath(inputPath, outputOption string) (string, error) {
	if outputOption == "" {
		// Use input filename with .md extension
		base := filepath.Base(inputPath)
		ext := filepath.Ext(base)
		return strings.TrimSuffix(base, ext) + ".md", nil
	}

	// Check if outputOption is a directory
	info, err := os.Stat(outputOption)
	if err == nil && info.IsDir() {
		base := filepath.Base(inputPath)
		ext := filepath.Ext(base)
		return filepath.Join(outputOption, strings.TrimSuffix(base, ext)+".md"), nil
	}

	// Output is a specific file path
	return outputOption, nil
}
