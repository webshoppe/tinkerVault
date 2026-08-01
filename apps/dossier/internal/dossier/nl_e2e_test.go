package dossier

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNLSearchVsStrictAND(t *testing.T) {
	root := t.TempDir()
	s, err := OpenOrCreate(root)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	p := filepath.Join(t.TempDir(), "dossier-test-plain-text.txt")
	body := "Dossier Tier 1 plain text import test. FTS5 search keyword for this file: ZEPHYRQUILL204 This is an ordinary .txt file with no markdown."
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ImportFile(p); err != nil {
		t.Fatal(err)
	}
	q := "What is the FTS5 test keyword in the plain-text smoke-test file?"
	strict, err := s.Search(q, 10)
	if err != nil {
		t.Fatal(err)
	}
	loose, err := s.SearchLoose(q, 10)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("strict hits=%d loose hits=%d looseQ=%s", len(strict), len(loose), buildFTSQueryLoose(q))
	if len(loose) < 1 {
		t.Fatal("loose must hit")
	}
	// Document the regression: strict AND of stopwords typically fails
	if len(strict) > 0 {
		t.Log("note: strict also hit (acceptable); loose still required for robustness")
	}
	// second question
	q2 := "Where can I find the ZEPHYRQUILL204 search testing keyword?"
	loose2, err := s.SearchLoose(q2, 10)
	if err != nil || len(loose2) < 1 {
		t.Fatalf("q2 hits=%d err=%v", len(loose2), err)
	}
}

func TestRescanSkipsOfficeLock(t *testing.T) {
	root := t.TempDir()
	s, err := OpenOrCreate(root)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	lock := filepath.Join(s.Root, DocumentsDir, "~$Book1.xlsx")
	if err := os.WriteFile(lock, []byte("not-a-real-xlsx"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.RescanDocuments(); err != nil {
		t.Fatal(err)
	}
	docs, _ := s.ListDocuments()
	for _, d := range docs {
		if strings.HasPrefix(d.Filename, "~$") {
			t.Fatal("imported lock", d.Filename)
		}
	}
}
