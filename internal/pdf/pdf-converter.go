package pdf

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"

	"github.com/rsc/pdf"
)

type Converter struct {
	AssetsDir string
}

// TextElement represents a piece of text with its styling and position
type TextElement struct {
	Text   string
	Font   string
	Size   float64
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// TextLine represents a line of text with its elements
type TextLine struct {
	Elements []TextElement
	Y        float64
	FontSize float64
	IsBold   bool
}

func (c *Converter) ToMarkdown(inputPath, outputPath string) error {
	// Open the PDF file
	f, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open PDF: %v", err)
	}
	defer f.Close()

	// Get file size for PDF reader
	fileInfo, err := f.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	// Create PDF reader with file size
	reader, err := pdf.NewReader(f, fileInfo.Size())
	if err != nil {
		return fmt.Errorf("failed to create PDF reader: %v", err)
	}

	// Get number of pages
	numPages := reader.NumPage()
	if numPages == 0 {
		return fmt.Errorf("PDF contains no pages")
	}

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	// Process each page
	for pageNum := 1; pageNum <= numPages; pageNum++ {
		page := reader.Page(pageNum)
		if page.V.IsNull() {
			continue // Skip empty pages
		}

		// Extract structured text from the page
		markdown, err := c.extractStructuredText(page)
		if err != nil {
			return fmt.Errorf("failed to extract text from page %d: %v", pageNum, err)
		}

		if strings.TrimSpace(markdown) == "" {
			continue
		}

		// Write the structured content
		outFile.WriteString(markdown)
		outFile.WriteString("\n\n")

		// Add page separator (except for last page)
		if pageNum < numPages {
			outFile.WriteString("---\n\n")
		}
	}

	return nil
}

func (c *Converter) extractStructuredText(page pdf.Page) (string, error) {
	content := page.Content()

	// Extract all text elements with their properties
	var elements []TextElement
	for _, text := range content.Text {
		decoded := decodePDFText(text.S)
		if decoded == "" {
			continue
		}

		element := TextElement{
			Text:  decoded,
			Font:  text.Font,
			Size:  text.FontSize,
			X:     text.X,
			Y:     text.Y,
			Width: text.W,
		}
		elements = append(elements, element)
	}

	if len(elements) == 0 {
		return "", nil
	}

	// Group elements into lines
	lines := c.groupElementsIntoLines(elements)

	// Detect document structure and convert to Markdown
	markdown := c.convertLinesToMarkdown(lines)

	return markdown, nil
}

func (c *Converter) groupElementsIntoLines(elements []TextElement) []TextLine {
	// Group elements by Y coordinate (same line)
	lineMap := make(map[float64][]TextElement)
	for _, element := range elements {
		lineMap[element.Y] = append(lineMap[element.Y], element)
	}

	// Sort elements within each line by X coordinate
	var lines []TextLine
	for y, lineElements := range lineMap {
		sort.Slice(lineElements, func(i, j int) bool {
			return lineElements[i].X < lineElements[j].X
		})

		// Calculate line properties
		line := TextLine{
			Elements: lineElements,
			Y:        y,
			FontSize: lineElements[0].Size, // Use first element's size as reference
			IsBold:   isBoldFont(lineElements[0].Font),
		}
		lines = append(lines, line)
	}

	// Sort lines by Y coordinate (top to bottom)
	sort.Slice(lines, func(i, j int) bool {
		return lines[i].Y > lines[j].Y // Higher Y means lower on page
	})

	return lines
}

func (c *Converter) convertLinesToMarkdown(lines []TextLine) string {
	var result strings.Builder
	var previousLine *TextLine
	var inList bool

	for i, line := range lines {
		lineText := c.extractLineText(line)
		if strings.TrimSpace(lineText) == "" {
			continue
		}

		// Detect heading based on font size and style
		if c.isHeading(line, previousLine) {
			level := c.getHeadingLevel(line)
			result.WriteString(strings.Repeat("#", level) + " " + lineText + "\n")
			inList = false
		} else if c.isListItem(lineText) {
			// Detect list item
			if !inList {
				result.WriteString("\n")
			}
			result.WriteString("- " + strings.TrimSpace(lineText[1:]) + "\n")
			inList = true
		} else if c.isTableRow(line) {
			// Detect table row (based on alignment and multiple elements)
			if i == 0 || !c.isTableRow(lines[i-1]) {
				result.WriteString("\n") // Table separator before first row
			}
			result.WriteString("| " + lineText + " |\n")
			if i == 0 {
				// Add header separator
				separator := "|" + strings.Repeat(" --- |", len(line.Elements)) + "\n"
				result.WriteString(separator)
			}
			inList = false
		} else {
			// Regular paragraph
			if inList {
				result.WriteString("\n")
				inList = false
			}

			result.WriteString(lineText + "\n\n")
		}

		previousLine = &line
	}

	return result.String()
}

func (c *Converter) extractLineText(line TextLine) string {
	var text strings.Builder
	for _, element := range line.Elements {
		text.WriteString(element.Text)
	}
	return text.String()
}

func (c *Converter) isHeading(line TextLine, previousLine *TextLine) bool {
	// Heading detection logic
	if line.FontSize > 14 {
		return true
	}
	if line.IsBold && line.FontSize > 12 {
		return true
	}
	if previousLine != nil && line.FontSize > previousLine.FontSize+2 {
		return true
	}
	return false
}

func (c *Converter) getHeadingLevel(line TextLine) int {
	// Simple heading level detection based on font size
	switch {
	case line.FontSize >= 20:
		return 1
	case line.FontSize >= 18:
		return 2
	case line.FontSize >= 16:
		return 3
	case line.FontSize >= 14:
		return 4
	default:
		return 5
	}
}

func (c *Converter) isListItem(lineText string) bool {
	// Detect list items (bullet points, numbers, etc.)
	trimmed := strings.TrimSpace(lineText)
	if len(trimmed) == 0 {
		return false
	}

	// Check for bullet points
	if strings.HasPrefix(trimmed, "•") || strings.HasPrefix(trimmed, "▪") ||
		strings.HasPrefix(trimmed, "‣") || strings.HasPrefix(trimmed, "-") ||
		strings.HasPrefix(trimmed, "𐀚") {
		return true
	}

	// Check for numbered lists (1., 2., a., b., etc.)
	if len(trimmed) > 2 && trimmed[1] == '.' {
		firstChar := trimmed[0]
		if (firstChar >= '0' && firstChar <= '9') ||
			(firstChar >= 'a' && firstChar <= 'z') ||
			(firstChar >= 'A' && firstChar <= 'Z') {
			return true
		}
	}

	return false
}

func (c *Converter) isTableRow(line TextLine) bool {
	// Simple table detection - multiple elements with similar Y positions
	if len(line.Elements) < 2 {
		return false
	}

	// Check if elements are aligned in a table-like structure
	for i := 1; i < len(line.Elements); i++ {
		prev := line.Elements[i-1]
		curr := line.Elements[i]

		// If elements are too close together, probably not a table
		if curr.X-prev.X-prev.Width < 5.0 {
			return false
		}
	}

	fmt.Println(line.Y, "is a table")

	return true
}

func isBoldFont(fontName string) bool {
	// Simple bold detection based on font name
	fontName = strings.ToLower(fontName)
	return strings.Contains(fontName, "bold") ||
		strings.Contains(fontName, "black") ||
		strings.Contains(fontName, "heavy")
}

func decodePDFText(text string) string {
	var result strings.Builder

	for _, r := range text {
		decoded := r + 29

		if unicode.IsPrint(decoded) {
			result.WriteRune(decoded)
		} else if decoded == 0x0A || decoded == 0x0D {
			result.WriteRune('\n')
		}
	}

	return result.String()
}
