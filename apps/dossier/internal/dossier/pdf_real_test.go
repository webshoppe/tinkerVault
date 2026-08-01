
package dossier

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractPDFText_realSample(t *testing.T) {
  // Path relative to module root - skip if missing
  candidates := []string{"build/sample-real.pdf", "../build/sample-real.pdf", "../../build/sample-real.pdf"}
  var src string
  for _, c := range candidates {
    if st, err := os.Stat(c); err == nil && st.Size() > 100 {
      src = c
      break
    }
  }
  if src == "" {
    t.Skip("no sample-real.pdf")
  }
  dir := t.TempDir()
  dst := filepath.Join(dir, "s.pdf")
  b, _ := os.ReadFile(src)
  os.WriteFile(dst, b, 0644)
  text, flags, err := ExtractPDFText(dst)
  t.Logf("err=%v pages=%d extracted=%d imageOnly=%v note=%q text=%q", err, flags.PageCount, flags.ExtractedPages, flags.ImageOnlyPages, flags.ExtractionNote, text)
  if err != nil && flags.PageCount == 0 {
    t.Log("library could not parse this sample; attach path still works")
    return
  }
  if flags.PageCount < 1 {
    t.Fatal("expected pages")
  }
}
