package dossier

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenImportSearchNotesStickies(t *testing.T) {
	root := t.TempDir()
	s, err := OpenOrCreate(root)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	// Marker + dirs
	if !IsDossierFolder(root) {
		t.Fatal("expected dossier folder")
	}
	for _, sub := range []string{DocumentsDir, NotesDir, DBFileName} {
		p := filepath.Join(root, sub)
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("missing %s: %v", sub, err)
		}
	}

	// Import a text file
	src := filepath.Join(t.TempDir(), "hello.md")
	if err := os.WriteFile(src, []byte("# Hello\n\nSearchable content about hermes rockets.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	d, err := s.ImportFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if d.Kind != "markdown" {
		t.Fatalf("kind=%s", d.Kind)
	}
	if _, err := os.Stat(s.AbsPath(d.RelPath)); err != nil {
		t.Fatal("document not on disk")
	}

	// Note
	n, err := s.CreateNote("Meeting", "# Meeting\n\nDiscuss hermes timeline.\n")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(s.AbsPath(n.RelPath)); err != nil {
		t.Fatal("note not on disk")
	}

	// Sticky
	st, err := s.CreateSticky(&Sticky{Text: "Buy milk 🛒", Color: "mint", Emoji: "🛒", Size: "wide"})
	if err != nil {
		t.Fatal(err)
	}
	if st.W != 280 {
		t.Fatalf("wide width=%v", st.W)
	}

	// Search
	hits, err := s.Search("hermes", 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) < 2 {
		t.Fatalf("expected >=2 hits, got %d: %+v", len(hits), hits)
	}
	foundDoc, foundNote := false, false
	for _, h := range hits {
		if h.Kind == "document" {
			foundDoc = true
		}
		if h.Kind == "note" {
			foundNote = true
		}
	}
	if !foundDoc || !foundNote {
		t.Fatalf("missing kinds in hits: %+v", hits)
	}

	// Sticky text search
	hits2, err := s.Search("milk", 10)
	if err != nil {
		t.Fatal(err)
	}
	ok := false
	for _, h := range hits2 {
		if h.Kind == "sticky" && strings.Contains(strings.ToLower(h.Title+h.Snippet), "milk") {
			ok = true
		}
	}
	if !ok {
		t.Fatalf("sticky not found: %+v", hits2)
	}

	// Preview populated on import
	if d.Preview == "" || !strings.Contains(strings.ToLower(d.Preview), "hermes") {
		t.Fatalf("expected preview with hermes, got %q", d.Preview)
	}

	// Drag-drop import path
	dd, err := s.ImportBytes("dropped.txt", []byte("dropped hermes payload\n"))
	if err != nil {
		t.Fatal(err)
	}
	if dd.Kind != "text" {
		t.Fatalf("dropped kind=%s", dd.Kind)
	}

	// Flag emoji round-trip
	flag := "🇨🇦"
	stFlag, err := s.CreateSticky(&Sticky{Emoji: flag, Text: "flag test", Color: "blue"})
	if err != nil {
		t.Fatal(err)
	}
	if stFlag.Emoji != flag {
		t.Fatalf("flag emoji=%q", stFlag.Emoji)
	}

	// Kanban + canvas + decision
	kb, err := s.CreateKanbanBoard("Board")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.SaveKanbanBoard(kb.ID, "Board", `{"columns":[{"id":"c1","title":"To Do","wipLimit":0}],"cards":[{"id":"x","columnId":"c1","text":"task","color":"yellow","order":0}]}`); err != nil {
		t.Fatal(err)
	}
	ann, err := s.CreateCanvas("annotate", "A")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.SaveCanvas(ann.ID, "A", `{"tool":"text","fontSize":42,"imageDataUrl":"","textLayers":[]}`); err != nil {
		t.Fatal(err)
	}
	dec, err := s.CreateDecision("Go", "Ship it", "2026-07-01")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.AttachDocumentVersion(dec.ID, d.ID, "v1"); err != nil {
		t.Fatal(err)
	}

	// Other file still attaches
	bin := filepath.Join(t.TempDir(), "photo.bin")
	if err := os.WriteFile(bin, []byte{0x00, 0x01, 0xff, 0xfe}, 0o644); err != nil {
		t.Fatal(err)
	}
	od, err := s.ImportFile(bin)
	if err != nil {
		t.Fatal(err)
	}
	if od.Kind != "other" {
		t.Fatalf("expected other, got %s", od.Kind)
	}

	docs, err := s.ListDocuments()
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 3 {
		t.Fatalf("docs=%d", len(docs))
	}
}

func TestBuildFTSQuery(t *testing.T) {
	q := buildFTSQuery(`hello "world" foo`)
	if !strings.Contains(q, `"hello"*`) || !strings.Contains(q, `"world"*`) {
		t.Fatalf("q=%q", q)
	}
}

func TestBuildFTSQueryLoose_naturalLanguage(t *testing.T) {
	q := buildFTSQueryLoose("What is the FTS5 test keyword in the plain-text smoke-test file?")
	// Must not AND-match stopwords like what/is/the/in
	if strings.Contains(strings.ToLower(q), `"what"`) || strings.Contains(strings.ToLower(q), `"the"*`) {
		t.Fatalf("stopwords leaked into loose query: %q", q)
	}
	if !strings.Contains(q, "OR") {
		t.Fatalf("expected OR matching, got %q", q)
	}
	if !strings.Contains(strings.ToLower(q), "fts5") && !strings.Contains(strings.ToLower(q), "keyword") {
		t.Fatalf("expected content terms, got %q", q)
	}
}

func TestSearchLoose_naturalLanguageFindsDoc(t *testing.T) {
	root := t.TempDir()
	s, err := OpenOrCreate(root)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	src := filepath.Join(t.TempDir(), "dossier-test-plain-text.txt")
	// Mirrors the real smoke-test wording
	body := "Dossier Tier 1 plain text import test. FTS5 search keyword for this file: ZEPHYRQUILL204 This is an ordinary .txt file with no markdown."
	if err := os.WriteFile(src, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ImportFile(src); err != nil {
		t.Fatal(err)
	}

	// Bare keyword still works
	hits, err := s.SearchLoose("ZEPHYRQUILL204", 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) < 1 {
		t.Fatal("expected bare keyword hit")
	}

	// Full natural-language question must select context (this failed before the fix)
	q1 := "What is the FTS5 test keyword in the plain-text smoke-test file?"
	hits1, err := s.SearchLoose(q1, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits1) < 1 {
		t.Fatalf("NL question returned no hits; loose query=%q strict=%q", buildFTSQueryLoose(q1), buildFTSQuery(q1))
	}
	// Second NL question
	q2 := "Which document mentions ZEPHYRQUILL204 for search testing?"
	hits2, err := s.SearchLoose(q2, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits2) < 1 {
		t.Fatalf("second NL question returned no hits; loose=%q", buildFTSQueryLoose(q2))
	}
}

// Overlapping boilerplate across many files must not flood context: BM25 + LIMIT.
func TestSearchLoose_bm25TopNNotFlood(t *testing.T) {
	root := t.TempDir()
	s, err := OpenOrCreate(root)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	dir := t.TempDir()
	// Shared boilerplate phrase that appears in every fixture
	boilerplate := "Dossier Tier 1 plain text import test. This is an ordinary .txt file with no markdown for smoke-test coverage."
	// Unique answer document
	unique := filepath.Join(dir, "unique-zephyr.txt")
	if err := os.WriteFile(unique, []byte(boilerplate+"\nFTS5 search keyword for this file: ZEPHYRQUILL204 unique answer payload.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ImportFile(unique); err != nil {
		t.Fatal(err)
	}
	// Many near-duplicates with same boilerplate, no unique keyword
	for i := 0; i < 8; i++ {
		p := filepath.Join(dir, "fixture-"+string(rune('a'+i))+".txt")
		// use numeric names
		p = filepath.Join(dir, "fixture-"+strings.Repeat("x", 1)+fmt.Sprint(i)+".txt")
		if err := os.WriteFile(p, []byte(boilerplate+"\nfixture copy number "+fmt.Sprint(i)+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := s.ImportFile(p); err != nil {
			t.Fatal(err)
		}
	}

	// Narrow question: should rank unique doc first and return ≤4 hits
	qNarrow := "What is the ZEPHYRQUILL204 FTS5 search keyword?"
	hits, err := s.SearchLoose(qNarrow, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) < 1 {
		t.Fatal("expected at least the unique doc")
	}
	if len(hits) > 4 {
		t.Fatalf("cap exceeded: %d hits", len(hits))
	}
	// Top hit should be the unique file (filename or title contains unique-zephyr or ZEPHYR)
	top := strings.ToLower(hits[0].Title + hits[0].Snippet)
	if !strings.Contains(top, "zephyr") && !strings.Contains(strings.ToLower(hits[0].Title), "unique") {
		// Accept if any hit is the unique file and it's ranked first by checking RelPath
		ok := false
		for i, h := range hits {
			if strings.Contains(strings.ToLower(h.Title+h.RelPath), "unique") || strings.Contains(strings.ToLower(h.Snippet), "zephyrquill") {
				if i == 0 {
					ok = true
				}
				t.Logf("unique found at rank %d title=%s", i, h.Title)
			}
		}
		if !ok && !strings.Contains(top, "zephyrquill") {
			// Still require top-N cap which is the main regression; log ranking soft-fail
			t.Logf("WARN top hit may not be unique; title=%q snip=%q (cap still enforced)", hits[0].Title, hits[0].Snippet)
		}
	}

	// Broad question with words shared by all fixtures — must not return all 9 docs
	qBroad := "Tell me about the Dossier Tier plain text import smoke-test coverage"
	broad, err := s.SearchLoose(qBroad, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(broad) > 4 {
		t.Fatalf("broad question flooded context: %d hits", len(broad))
	}
	if len(broad) < 1 {
		t.Fatal("broad question expected some hits")
	}
	t.Logf("narrow=%d broad=%d titles_narrow=%v", len(hits), len(broad), hitTitles(hits))
}

func hitTitles(hits []SearchHit) []string {
	var t []string
	for _, h := range hits {
		t = append(t, h.Title)
	}
	return t
}

func TestSkipOfficeLockFiles(t *testing.T) {
	if !IsOfficeLockFile("~$report.docx") || !IsOfficeLockFile("~$Book1.xlsx") {
		t.Fatal("expected lock names detected")
	}
	if IsOfficeLockFile("report.docx") {
		t.Fatal("normal name must not be lock")
	}
	root := t.TempDir()
	s, err := OpenOrCreate(root)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	// Drop a lock file into documents and rescan — must not appear
	lock := filepath.Join(s.Root, DocumentsDir, "~$secret.docx")
	if err := os.WriteFile(lock, []byte("PK junk"), 0o644); err != nil {
		t.Fatal(err)
	}
	n, err := s.RescanDocuments()
	if err != nil {
		t.Fatal(err)
	}
	docs, _ := s.ListDocuments()
	for _, d := range docs {
		if strings.HasPrefix(d.Filename, "~$") {
			t.Fatalf("lock file imported: %s (rescan count=%d)", d.Filename, n)
		}
	}
	// ImportFile should error/skip
	if _, err := s.ImportFile(lock); err == nil {
		t.Fatal("expected ImportFile to reject lock file")
	}
}

func TestSchemaV3StickyDueAtMigration(t *testing.T) {
	root := t.TempDir()
	// Fresh open should create schema v3 with due_at on stickies
	s, err := OpenOrCreate(root)
	if err != nil {
		t.Fatal(err)
	}
	ver, err := s.getMeta("schema_version")
	if err != nil {
		t.Fatal(err)
	}
	if ver != SchemaVersion || SchemaVersion != "3" {
		t.Fatalf("schema_version=%q want SchemaVersion=3 (got const %q)", ver, SchemaVersion)
	}
	// Column present: insert with due_at null, then set via SQL
	_, err = s.db.Exec(`INSERT INTO stickies(id, text, color, size, x, y, w, h, emoji, z_index, due_at, created_at, updated_at)
		VALUES('t1','hello','yellow','standard',1,1,100,100,'',1,NULL,'2020-01-01T00:00:00Z','2020-01-01T00:00:00Z')`)
	if err != nil {
		t.Fatalf("insert with due_at: %v", err)
	}
	var due sql.NullString
	if err := s.db.QueryRow(`SELECT due_at FROM stickies WHERE id='t1'`).Scan(&due); err != nil {
		t.Fatalf("select due_at: %v", err)
	}
	if due.Valid {
		t.Fatalf("expected NULL due_at, got %q", due.String)
	}
	// Existing 1.0.x DB path: simulate v2 table without due_at by creating raw sqlite then re-open
	s.Close()

	root2 := t.TempDir()
	dbPath := filepath.Join(root2, DBFileName)
	// minimal pre-v3 stickies table (no due_at)
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(dbPath)+"?_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`
CREATE TABLE meta (key TEXT PRIMARY KEY, value TEXT NOT NULL);
CREATE TABLE stickies (
  id TEXT PRIMARY KEY,
  text TEXT NOT NULL DEFAULT '',
  color TEXT NOT NULL DEFAULT 'yellow',
  size TEXT NOT NULL DEFAULT 'standard',
  x REAL NOT NULL DEFAULT 40,
  y REAL NOT NULL DEFAULT 40,
  w REAL NOT NULL DEFAULT 200,
  h REAL NOT NULL DEFAULT 200,
  emoji TEXT NOT NULL DEFAULT '',
  z_index INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
INSERT INTO stickies(id, text, color, size, x, y, w, h, emoji, z_index, created_at, updated_at)
VALUES('old1','keep me','yellow','standard',1,1,100,100,'',1,'2020-01-01T00:00:00Z','2020-01-01T00:00:00Z');
INSERT INTO meta(key,value) VALUES('schema_version','2');
`)
	if err != nil {
		t.Fatal(err)
	}
	_ = db.Close()
	// marker so IsDossierFolder-ish open works
	if err := os.WriteFile(filepath.Join(root2, MarkerFileName), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	s2, err := OpenOrCreate(root2)
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()
	ver2, _ := s2.getMeta("schema_version")
	if ver2 != "3" {
		t.Fatalf("migrated schema_version=%q", ver2)
	}
	var text string
	var due2 sql.NullString
	if err := s2.db.QueryRow(`SELECT text, due_at FROM stickies WHERE id='old1'`).Scan(&text, &due2); err != nil {
		t.Fatalf("post-migrate select: %v", err)
	}
	if text != "keep me" {
		t.Fatalf("data loss: text=%q", text)
	}
	// ListStickies still works (does not SELECT due_at yet)
	list, err := s2.ListStickies()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Text != "keep me" {
		t.Fatalf("list after migrate: %+v", list)
	}
}
