package dossier

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/ledongthuc/pdf"
)

// ExtractPDFText extracts text from a PDF file. Image-only pages are reported
// in PDFFlags.ImageOnlyPages (1-based page numbers).
func ExtractPDFText(path string) (text string, flags *PDFFlags, err error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", &PDFFlags{
			ExtractionNote: "Could not open PDF: " + err.Error(),
		}, err
	}
	defer f.Close()

	total := r.NumPage()
	flags = &PDFFlags{
		PageCount:      total,
		ImageOnlyPages: []int{},
	}

	var b strings.Builder
	for i := 1; i <= total; i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			flags.ImageOnlyPages = append(flags.ImageOnlyPages, i)
			continue
		}
		content, err := p.GetPlainText(nil)
		if err != nil {
			flags.ImageOnlyPages = append(flags.ImageOnlyPages, i)
			continue
		}
		content = strings.TrimSpace(content)
		if isEffectivelyEmpty(content) {
			flags.ImageOnlyPages = append(flags.ImageOnlyPages, i)
			continue
		}
		flags.ExtractedPages++
		if b.Len() > 0 {
			b.WriteString("\n\n")
		}
		fmt.Fprintf(&b, "--- Page %d ---\n%s", i, content)
	}

	flags.HasImageOnly = len(flags.ImageOnlyPages) > 0
	if flags.ExtractedPages == 0 && total > 0 {
		flags.ExtractionNote = "No extractable text found. All pages appear to be image-only (scanned). OCR is not available in Tier 1."
	} else if flags.HasImageOnly {
		flags.ExtractionNote = fmt.Sprintf(
			"%d of %d pages had no extractable text (likely image-only/scanned).",
			len(flags.ImageOnlyPages), total,
		)
	}

	return b.String(), flags, nil
}

func isEffectivelyEmpty(s string) bool {
	if s == "" {
		return true
	}
	letters := 0
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			letters++
			if letters >= 3 {
				return false
			}
		}
	}
	return true
}

// ReadFileBytes is a tiny helper used by tests/import.
func ReadFileBytes(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// ExtractTextBytes is used when content is already in memory (unused currently).
func ExtractTextBytes(data []byte) (string, error) {
	// Not used for PDF; plain-text path only.
	return string(bytes.TrimSpace(data)), nil
}
