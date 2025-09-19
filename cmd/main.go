package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/leandrowiemesfilho/markdown-converter/internal/converter"
	"github.com/leandrowiemesfilho/markdown-converter/internal/utils"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "doc2md",
		Usage: "Convert documents (PDF, DOCX, XLSX, PPTX) to Markdown",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file or directory",
			},
			&cli.StringFlag{
				Name:    "assets-dir",
				Aliases: []string{"a"},
				Value:   "assets",
				Usage:   "Directory for extracted assets",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Enable verbose output",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() == 0 {
				return fmt.Errorf("no input files specified")
			}

			outputOption := c.String("output")
			assetsDir := c.String("assets-dir")
			verbose := c.Bool("verbose")

			// Create assets directory
			if err := utils.EnsureDir(assetsDir); err != nil {
				return fmt.Errorf("failed to create assets directory: %v", err)
			}

			// Process each input file
			for _, inputPath := range c.Args().Slice() {
				if verbose {
					log.Printf("Processing: %s", inputPath)
				}

				if err := convertFile(inputPath, outputOption, assetsDir, verbose); err != nil {
					return fmt.Errorf("failed to convert %s: %v", inputPath, err)
				}

				if verbose {
					log.Printf("Successfully converted: %s", inputPath)
				}
			}

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func convertFile(inputPath, outputOption, assetsDir string, verbose bool) error {
	// Check if input file exists
	if !utils.FileExists(inputPath) {
		return fmt.Errorf("input file does not exist: %s", inputPath)
	}

	// Get appropriate converter
	conv, fileType, err := converter.GetConverter(inputPath, assetsDir)
	if err != nil {
		return err
	}

	if verbose {
		log.Printf("Detected file type: %s", fileType)
	}

	// Determine output path
	outputPath, err := utils.GetOutputPath(inputPath, outputOption)
	if err != nil {
		return fmt.Errorf("failed to determine output path: %v", err)
	}

	// Create output directory if needed
	if err := utils.EnsureDir(filepath.Dir(outputPath)); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	if verbose {
		log.Printf("Output: %s", outputPath)
	}

	// Perform conversion
	return conv.ToMarkdown(inputPath, outputPath)
}
