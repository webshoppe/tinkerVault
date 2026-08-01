package dossier

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeMinimalDOCX(t *testing.T, path, plain string) {
	t.Helper()
	// Minimal OOXML package with one paragraph
	doc := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body><w:p><w:r><w:t>` + plain + `</w:t></w:r></w:p></w:body>
</w:document>`
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(doc)); err != nil {
		t.Fatal(err)
	}
	// [Content_Types] optional for our reader
	_ = zw.Close()
	_ = f.Close()
}

func writeMinimalXLSX(t *testing.T, path string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	ss, _ := zw.Create("xl/sharedStrings.xml")
	_, _ = ss.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="2" uniqueCount="2">
  <si><t>Alpha</t></si><si><t>Beta</t></si>
</sst>`))
	sh, _ := zw.Create("xl/worksheets/sheet1.xml")
	_, _ = sh.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
    <row r="1"><c r="A1" t="s"><v>0</v></c><c r="B1" t="s"><v>1</v></c></row>
  </sheetData>
</worksheet>`))
	_ = zw.Close()
	_ = f.Close()
}

func TestExtractDOCXAndXLSX(t *testing.T) {
	dir := t.TempDir()
	docx := filepath.Join(dir, "brief.docx")
	writeMinimalDOCX(t, docx, "Harbor Lane renovation plan")
	text, flags, err := ExtractOfficeText(docx)
	if err != nil {
		t.Fatal(err)
	}
	if flags == nil || flags.Format != "docx" || !flags.Partial {
		t.Fatalf("flags=%+v", flags)
	}
	if !strings.Contains(text, "Harbor Lane") {
		t.Fatalf("docx text=%q", text)
	}

	xlsx := filepath.Join(dir, "sheet.xlsx")
	writeMinimalXLSX(t, xlsx)
	xt, xf, err := ExtractOfficeText(xlsx)
	if err != nil {
		t.Fatal(err)
	}
	if xf == nil || xf.Format != "xlsx" {
		t.Fatalf("xlsx flags=%+v", xf)
	}
	if !strings.Contains(xt, "Alpha") || !strings.Contains(xt, "Beta") {
		t.Fatalf("xlsx text=%q", xt)
	}
}

func TestImportOfficeKinds(t *testing.T) {
	root := t.TempDir()
	s, err := OpenOrCreate(root)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	docx := filepath.Join(t.TempDir(), "a.docx")
	writeMinimalDOCX(t, docx, "Searchable hermes office body")
	d, err := s.ImportFile(docx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Kind != "docx" {
		t.Fatalf("kind=%s", d.Kind)
	}
	if d.OfficeFlags == nil {
		t.Fatal("expected office flags")
	}
	hits, err := s.Search("hermes", 10)
	if err != nil {
		t.Fatal(err)
	}
	ok := false
	for _, h := range hits {
		if h.Kind == "document" {
			ok = true
		}
	}
	if !ok {
		t.Fatalf("expected search hit, got %+v", hits)
	}
}

func writeMinimalODT(t *testing.T, path, plain string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	// ODF packages include mimetype (optional for our reader) + content.xml
	w, err := zw.Create("content.xml")
	if err != nil {
		t.Fatal(err)
	}
	doc := `<?xml version="1.0" encoding="UTF-8"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
 xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
  <office:body><office:text>
    <text:p>` + plain + `</text:p>
  </office:text></office:body>
</office:document-content>`
	if _, err := w.Write([]byte(doc)); err != nil {
		t.Fatal(err)
	}
	_ = zw.Close()
	_ = f.Close()
}

func writeMinimalODS(t *testing.T, path, cellA, cellB string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("content.xml")
	if err != nil {
		t.Fatal(err)
	}
	doc := `<?xml version="1.0" encoding="UTF-8"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
 xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0"
 xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0">
  <office:body><office:spreadsheet>
    <table:table table:name="Sheet1">
      <table:table-row>
        <table:table-cell><text:p>` + cellA + `</text:p></table:table-cell>
        <table:table-cell><text:p>` + cellB + `</text:p></table:table-cell>
      </table:table-row>
    </table:table>
  </office:spreadsheet></office:body>
</office:document-content>`
	if _, err := w.Write([]byte(doc)); err != nil {
		t.Fatal(err)
	}
	_ = zw.Close()
	_ = f.Close()
}

func TestExtractODTAndODS(t *testing.T) {
	dir := t.TempDir()
	odt := filepath.Join(dir, "brief.odt")
	writeMinimalODT(t, odt, "ZEPHYRODT2041 open document body")
	text, flags, err := ExtractOfficeText(odt)
	if err != nil {
		t.Fatal(err)
	}
	if flags == nil || flags.Format != "odt" || !flags.Partial {
		t.Fatalf("odt flags=%+v", flags)
	}
	if !strings.Contains(text, "ZEPHYRODT2041") {
		t.Fatalf("odt text=%q", text)
	}

	ods := filepath.Join(dir, "sheet.ods")
	writeMinimalODS(t, ods, "AlphaODS", "BetaODS")
	xt, xf, err := ExtractOfficeText(ods)
	if err != nil {
		t.Fatal(err)
	}
	if xf == nil || xf.Format != "ods" {
		t.Fatalf("ods flags=%+v", xf)
	}
	if !strings.Contains(xt, "AlphaODS") || !strings.Contains(xt, "BetaODS") {
		t.Fatalf("ods text=%q", xt)
	}
}

func TestImportOpenDocKindsAndRescan(t *testing.T) {
	root := t.TempDir()
	s, err := OpenOrCreate(root)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	odtSrc := filepath.Join(t.TempDir(), "memo.odt")
	writeMinimalODT(t, odtSrc, "Searchable odt keyword ORIONODT99")
	d, err := s.ImportFile(odtSrc)
	if err != nil {
		t.Fatal(err)
	}
	if d.Kind != "odt" {
		t.Fatalf("kind=%s", d.Kind)
	}
	if d.OfficeFlags == nil || !d.OfficeFlags.Partial {
		t.Fatalf("expected partial office flags, got %+v", d.OfficeFlags)
	}
	hits, err := s.Search("ORIONODT99", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) < 1 {
		t.Fatalf("expected FTS hit for odt, got %+v", hits)
	}

	// Rescan path: drop an .ods directly into documents/ then rescan
	odsName := "budget.ods"
	odsDest := filepath.Join(s.Root, DocumentsDir, odsName)
	writeMinimalODS(t, odsDest, "VEGAODS77", "sheet-cell")
	n, err := s.RescanDocuments()
	if err != nil {
		t.Fatal(err)
	}
	if n < 1 {
		t.Fatalf("expected rescan to pick up ods, count=%d", n)
	}
	hits2, err := s.Search("VEGAODS77", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits2) < 1 {
		t.Fatalf("expected FTS hit after rescan ods, got %+v", hits2)
	}
	// Confirm kind via list
	docs, err := s.ListDocuments()
	if err != nil {
		t.Fatal(err)
	}
	foundODS := false
	for _, doc := range docs {
		if doc.Filename == odsName {
			foundODS = true
			if doc.Kind != "ods" {
				t.Fatalf("rescan kind=%s want ods", doc.Kind)
			}
			if doc.OfficeFlags == nil {
				t.Fatal("rescan ods missing office flags")
			}
		}
	}
	if !foundODS {
		t.Fatal("ods not in document list after rescan")
	}
}

func TestKindFromExtOpenDoc(t *testing.T) {
	if KindFromExt("a.odt") != "odt" || KindFromExt("b.ODS") != "ods" {
		t.Fatal("KindFromExt odt/ods")
	}
}
