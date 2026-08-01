// askprobe: open a temp dossier with fixture docs, point agent at a local listener,
// call AskDossier, print contextUsed count/titles. Used to verify BM25 top-N selection.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/webshoppe/dossier/internal/app"
	"github.com/webshoppe/dossier/internal/dossier"
)

func main() {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fail(err.Error())
	}
	port := ln.Addr().(*net.TCPAddr).Port

	var (
		mu      sync.Mutex
		gotBody []byte
		gotPath string
	)
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(io.LimitReader(r.Body, 2<<20))
		mu.Lock()
		gotBody = b
		gotPath = r.URL.Path
		mu.Unlock()
		// Minimal OpenAI-shaped reply
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"probe-ok"}}]}`))
	})}
	go srv.Serve(ln)
	defer srv.Close()

	// Isolate config
	cfgDir := filepath.Join(os.TempDir(), fmt.Sprintf("dossier-askprobe-cfg-%d", os.Getpid()))
	_ = os.RemoveAll(cfgDir)
	_ = os.MkdirAll(cfgDir, 0o755)
	_ = os.Setenv("APPDATA", cfgDir)

	root := filepath.Join(os.TempDir(), fmt.Sprintf("dossier-askprobe-%d", os.Getpid()))
	_ = os.RemoveAll(root)
	defer os.RemoveAll(root)

	store, err := dossier.OpenOrCreate(root)
	if err != nil {
		fail(err.Error())
	}
	// Unique answer doc
	mustWriteImport(store, "unique-zephyr.txt",
		"Dossier Tier 1 plain text import test.\nFTS5 search keyword for this file: ZEPHYRQUILL204 unique answer payload.\n")
	// Overlapping fixtures
	boilerplate := "Dossier Tier 1 plain text import test. This is an ordinary .txt file with no markdown for smoke-test coverage.\n"
	for i := 0; i < 8; i++ {
		mustWriteImport(store, fmt.Sprintf("fixture-%d.txt", i), boilerplate+fmt.Sprintf("fixture copy %d\n", i))
	}
	store.Close()

	api := app.NewAPI()
	if _, err := api.OpenDossierPath(root); err != nil {
		fail("open: " + err.Error())
	}
	host := "127.0.0.1"
	path := "/v1/chat/completions"
	if _, err := api.SaveAppSettings(app.SettingsPatch{
		AgentHost: &host,
		AgentPort: &port,
		AgentPath: &path,
	}); err != nil {
		fail("settings: " + err.Error())
	}

	// Narrow question
	narrow := api.AskDossier(app.AskDossierRequest{Question: "What is the ZEPHYRQUILL204 FTS5 search keyword?"})
	if narrow.Error != "" && !strings.Contains(narrow.Error, "probe") {
		// HTTP should succeed
		if narrow.Answer == "" && narrow.Error != "" {
			fail("narrow ask: " + narrow.Error)
		}
	}
	time.Sleep(50 * time.Millisecond)
	mu.Lock()
	body1 := append([]byte(nil), gotBody...)
	mu.Unlock()

	// Broad overlapping question
	broad := api.AskDossier(app.AskDossierRequest{Question: "Tell me about the Dossier Tier plain text import smoke-test coverage"})
	if broad.Error != "" && broad.Answer == "" {
		fail("broad ask: " + broad.Error)
	}
	time.Sleep(50 * time.Millisecond)
	mu.Lock()
	body2 := append([]byte(nil), gotBody...)
	mu.Unlock()

	fmt.Println("NARROW contextUsed=", len(narrow.ContextUsed))
	for _, c := range narrow.ContextUsed {
		fmt.Println("  N:", c["kind"], c["title"])
	}
	fmt.Println("BROAD contextUsed=", len(broad.ContextUsed))
	for _, c := range broad.ContextUsed {
		fmt.Println("  B:", c["kind"], c["title"])
	}

	// Parse context field from last payloads
	printContextDocs("NARROW_PAYLOAD", body1)
	printContextDocs("BROAD_PAYLOAD", body2)

	if len(narrow.ContextUsed) < 1 {
		fail("narrow expected ≥1 context doc")
	}
	if len(narrow.ContextUsed) > 4 {
		fail(fmt.Sprintf("narrow flooded: %d", len(narrow.ContextUsed)))
	}
	if len(broad.ContextUsed) > 4 {
		fail(fmt.Sprintf("broad flooded: %d", len(broad.ContextUsed)))
	}
	// Narrow should prefer unique-zephyr
	joined := ""
	for _, c := range narrow.ContextUsed {
		joined += c["title"] + " "
	}
	if !strings.Contains(strings.ToLower(joined), "unique") && !strings.Contains(strings.ToLower(joined), "zephyr") {
		fmt.Println("WARN: unique doc not obviously in narrow context titles:", joined)
	}
	fmt.Println("ASKPROBE_OK port=", port, "path=", gotPath)
}

func printContextDocs(label string, body []byte) {
	if len(body) == 0 {
		fmt.Println(label, "empty body")
		return
	}
	var m map[string]interface{}
	if json.Unmarshal(body, &m) != nil {
		fmt.Println(label, "non-json len=", len(body))
		return
	}
	ctx, _ := m["context"].(string)
	// Count ### headers = documents in context
	n := strings.Count(ctx, "### ")
	fmt.Printf("%s context_headers=%d context_runes≈%d\n", label, n, len([]rune(ctx)))
}

func mustWriteImport(s *dossier.Store, name, body string) {
	p := filepath.Join(s.Root, "_src_"+name)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		fail(err.Error())
	}
	if _, err := s.ImportFile(p); err != nil {
		fail(err.Error())
	}
}

func fail(msg string) {
	fmt.Fprintln(os.Stderr, "ASKPROBE_FAIL:", msg)
	os.Exit(1)
}
