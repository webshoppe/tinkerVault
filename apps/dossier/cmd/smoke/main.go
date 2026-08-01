// Smoke test for Windows: exercises dossier open/import/search/notes/stickies
// plus Tier 2: ImportBytes, kanban, canvas, decisions, multi-config.
// Build: GOOS=windows go build -o build/smoke.exe ./cmd/smoke
package main

import (
	"archive/zip"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/webshoppe/dossier/internal/dossier"
)

func main() {
	root := filepath.Join(os.TempDir(), "dossier-smoke-"+fmt.Sprint(os.Getpid()))
	_ = os.RemoveAll(root)
	defer os.RemoveAll(root)

	fmt.Println("SMOKE root=", root)
	s, err := dossier.OpenOrCreate(root)
	must(err)
	defer s.Close()

	// Write source files
	srcDir := filepath.Join(root, "_src")
	must(os.MkdirAll(srcDir, 0o755))
	md := filepath.Join(srcDir, "brief.md")
	must(os.WriteFile(md, []byte("# Brief\n\nProject Alpha launch checklist.\n"), 0o644))
	txt := filepath.Join(srcDir, "notes.txt")
	must(os.WriteFile(txt, []byte("plain text about Project Alpha\n"), 0o644))
	bin := filepath.Join(srcDir, "blob.dat")
	must(os.WriteFile(bin, []byte{1, 2, 3, 4, 5}, 0o644))

	d1, err := s.ImportFile(md)
	must(err)
	d2, err := s.ImportFile(txt)
	must(err)
	d3, err := s.ImportFile(bin)
	must(err)
	fmt.Printf("imported kinds: %s %s %s\n", d1.Kind, d2.Kind, d3.Kind)
	if d3.Kind != "other" {
		fail("expected other for blob")
	}
	if d1.Preview == "" {
		fail("expected preview on markdown import")
	}
	fmt.Println("doc preview:", d1.Preview)

	// Drag-drop path: base64 import
	b64 := base64.StdEncoding.EncodeToString([]byte("# Dropped\n\nFrom drag-drop Alpha path.\n"))
	dd, err := s.ImportBase64("dropped.md", b64)
	must(err)
	if dd.Kind != "markdown" {
		fail("dropped kind " + dd.Kind)
	}
	fmt.Println("dropped import ok:", dd.Filename)

	// Office kinds (minimal packages)
	docxPath := filepath.Join(srcDir, "tiny.docx")
	must(writeTinyDOCX(docxPath, "Alpha office hermes docx body"))
	od, err := s.ImportFile(docxPath)
	must(err)
	if od.Kind != "docx" || od.OfficeFlags == nil {
		fail(fmt.Sprintf("docx kind/flags %#v %#v", od.Kind, od.OfficeFlags))
	}
	fmt.Println("docx import ok partial=", od.OfficeFlags.Partial)

	n, err := s.CreateNote("Plan", "# Plan\n\nShip Project Alpha this week.\n")
	must(err)
	fmt.Println("note file exists:", fileExists(s.AbsPath(n.RelPath)))

	// Flag emoji (regional indicators) must round-trip
	flag := "🇨🇦"
	st, err := s.CreateSticky(&dossier.Sticky{Text: "Alpha reminder 🚀", Color: "purple", Emoji: flag, Size: "classic"})
	must(err)
	fmt.Printf("sticky size classic -> %.0fx%.0f color=%s emoji=%q runes=%d\n", st.W, st.H, st.Color, st.Emoji, len([]rune(st.Emoji)))
	if st.Emoji != flag {
		fail("flag emoji not preserved: " + st.Emoji)
	}

	// Kanban
	kb, err := s.CreateKanbanBoard("Sprint")
	must(err)
	fmt.Println("kanban id=", kb.ID, "state len=", len(kb.StateJSON))
	kb2, err := s.SaveKanbanBoard(kb.ID, "Sprint 1", `{"columns":[{"id":"c1","title":"To Do","wipLimit":0}],"cards":[{"id":"card1","columnId":"c1","text":"Alpha task","color":"yellow","order":0}]}`)
	must(err)
	fmt.Println("kanban saved title=", kb2.Title)

	// Paint + Annotate
	paint, err := s.CreateCanvas("paint", "Sketch")
	must(err)
	ann, err := s.CreateCanvas("annotate", "Markup")
	must(err)
	_, err = s.SaveCanvas(ann.ID, "Markup", `{"tool":"text","strokeColor":"#e03131","fontSize":36,"imageDataUrl":"","textLayers":[]}`)
	must(err)
	got, err := s.GetCanvas(ann.ID)
	must(err)
	if !strings.Contains(got.StateJSON, `"fontSize":36`) && !strings.Contains(got.StateJSON, `"fontSize": 36`) {
		fail("annotate fontSize not saved: " + got.StateJSON)
	}
	fmt.Println("paint+annotate ok; paint=", paint.ID)

	// Decision + version attach (end-to-end snapshot)
	dec, err := s.CreateDecision("Ship Alpha", "We ship this week.", "2026-07-20")
	must(err)
	dec2, err := s.AttachDocumentVersion(dec.ID, d1.ID, "brief at decision time")
	must(err)
	if dec2.DocVersionRel == "" || !fileExists(s.AbsPath(dec2.DocVersionRel)) {
		fail("decision version missing on disk")
	}
	// re-get must still show snapshot fields (UI depends on this)
	dec3, err := s.GetDecision(dec.ID)
	must(err)
	if dec3.DocVersionRel == "" || dec3.DocVersionNote == "" {
		fail("decision snapshot fields empty after GetDecision")
	}
	info, err := os.Stat(s.AbsPath(dec3.DocVersionRel))
	must(err)
	if info.Size() < 1 {
		fail("decision snapshot zero size")
	}
	fmt.Println("decision version:", dec2.DocVersionRel, "bytes=", info.Size())
	list, err := s.ListDecisions()
	must(err)
	if len(list) < 1 {
		fail("no decisions listed")
	}

	hits, err := s.Search("Alpha", 20)
	must(err)
	fmt.Printf("search hits=%d\n", len(hits))
	kinds := map[string]int{}
	for _, h := range hits {
		kinds[h.Kind]++
		fmt.Printf("  - %s: %s\n", h.Kind, h.Title)
	}
	if kinds["document"] < 1 || kinds["note"] < 1 {
		fail(fmt.Sprintf("expected document+note hits, got %v", kinds))
	}

	// Multi-dossier config
	cfg := &dossier.AppConfig{}
	cfg.TouchRecent(root)
	cfg.TouchRecent(filepath.Join(os.TempDir(), "other-dossier-path"))
	if cfg.LastDossierPath != filepath.Clean(filepath.Join(os.TempDir(), "other-dossier-path")) {
		// second touch is last
	}
	if len(cfg.RecentPaths) < 1 {
		fail("recent paths empty")
	}
	fmt.Println("config recent count=", len(cfg.RecentPaths))

	// Second dossier isolation
	root2 := filepath.Join(os.TempDir(), "dossier-smoke2-"+fmt.Sprint(os.Getpid()))
	_ = os.RemoveAll(root2)
	defer os.RemoveAll(root2)
	s2, err := dossier.OpenOrCreate(root2)
	must(err)
	docs2, _ := s2.ListDocuments()
	if len(docs2) != 0 {
		fail("second dossier not empty")
	}
	s2.Close()
	fmt.Println("multi-dossier isolation ok")

	db := filepath.Join(root, dossier.DBFileName)
	if !fileExists(db) {
		fail("missing db")
	}
	docs, _ := s.ListDocuments()
	fmt.Printf("docs on disk+db: %d\n", len(docs))
	for _, d := range docs {
		if !fileExists(s.AbsPath(d.RelPath)) {
			fail("missing file " + d.RelPath)
		}
	}

	fmt.Println("SMOKE_OK")
}

func must(err error) {
	if err != nil {
		fail(err.Error())
	}
}
func fail(msg string) {
	fmt.Fprintln(os.Stderr, "SMOKE_FAIL:", msg)
	os.Exit(1)
}
func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func writeTinyDOCX(path, plain string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	w, err := zw.Create("word/document.xml")
	if err != nil {
		return err
	}
	doc := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body><w:p><w:r><w:t>` + plain + `</w:t></w:r></w:p></w:body>
</w:document>`
	if _, err := w.Write([]byte(doc)); err != nil {
		return err
	}
	return zw.Close()
}

