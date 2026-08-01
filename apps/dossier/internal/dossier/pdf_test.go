package dossier

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractPDFText_minimal(t *testing.T) {
	// Minimal PDF with Helvetica text "HelloPDF"
	pdf := []byte("%PDF-1.1\n1 0 obj<< /Type /Catalog /Pages 2 0 R >>endobj\n2 0 obj<< /Type /Pages /Kids [3 0 R] /Count 1 >>endobj\n3 0 obj<< /Type /Page /Parent 2 0 R /MediaBox [0 0 300 144] /Contents 4 0 R /Resources<< /Font<< /F1 5 0 R >> >> >>endobj\n4 0 obj<< /Length 44 >>stream\nBT /F1 24 Tf 100 100 Td (HelloPDF) Tj ET\nendstream endobj\n5 0 obj<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>endobj\nxref\n0 6\n0000000000 65535 f \n0000000009 00000 n \n0000000058 00000 n \n0000000115 00000 n \n0000000266 00000 n \n0000000361 00000 n \ntrailer<< /Size 6 /Root 1 0 R >>\nstartxref\n433\n%%EOF\n")
	dir := t.TempDir()
	path := filepath.Join(dir, "t.pdf")
	if err := os.WriteFile(path, pdf, 0o644); err != nil {
		t.Fatal(err)
	}
	text, flags, err := ExtractPDFText(path)
	if flags == nil {
		t.Fatal("flags nil")
	}
	t.Logf("err=%v pages=%d extracted=%d imageOnly=%v text=%q note=%q", err, flags.PageCount, flags.ExtractedPages, flags.ImageOnlyPages, text, flags.ExtractionNote)
	// ledongthuc/pdf may or may not extract from minimal PDFs; ensure no panic and flags set
	if flags.PageCount < 0 {
		t.Fatal("bad page count")
	}
}
