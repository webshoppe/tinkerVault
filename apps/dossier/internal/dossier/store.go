package dossier

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

const (
	DBFileName     = "dossier.db"
	DocumentsDir   = "documents"
	NotesDir       = "notes"
	BoardsDir      = "boards" // paint / annotate PNG + sidecars
	VersionsDir    = "decision-versions"
	MarkerFileName = ".dossier"
	AppDataRelPath = "Dossier"
	ConfigFileName = "config.json"
	SchemaVersion  = "3"
)

// Store is the SQLite-backed state for one dossier workspace.
type Store struct {
	mu   sync.RWMutex
	db   *sql.DB
	Root string
}

type Document struct {
	ID          string       `json:"id"`
	Filename    string       `json:"filename"`
	RelPath     string       `json:"relPath"`
	MimeType    string       `json:"mimeType"`
	SizeBytes   int64        `json:"sizeBytes"`
	Kind        string       `json:"kind"` // markdown, text, pdf, docx, xlsx, odt, ods, other
	PDFFlags    *PDFFlags    `json:"pdfFlags,omitempty"`
	OfficeFlags *OfficeFlags `json:"officeFlags,omitempty"`
	CreatedAt   string       `json:"createdAt"`
	UpdatedAt   string       `json:"updatedAt"`
	Mtime       string       `json:"mtime"`
	IndexedAt   string       `json:"indexedAt,omitempty"`
	Preview     string       `json:"preview,omitempty"`
}

type PDFFlags struct {
	PageCount       int   `json:"pageCount"`
	ExtractedPages  int   `json:"extractedPages"`
	ImageOnlyPages  []int `json:"imageOnlyPages"`
	HasImageOnly    bool  `json:"hasImageOnly"`
	ExtractionNote  string `json:"extractionNote,omitempty"`
}

type Note struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	RelPath   string `json:"relPath"`
	Body      string `json:"body,omitempty"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type Sticky struct {
	ID        string  `json:"id"`
	Text      string  `json:"text"`
	Color     string  `json:"color"`
	Size      string  `json:"size"` // mini, standard, classic, wide, tall, large
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	W         float64 `json:"w"`
	H         float64 `json:"h"`
	Emoji     string  `json:"emoji"`
	ZIndex    int     `json:"zIndex"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
}

type SearchHit struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"` // document, note, sticky
	Title    string `json:"title"`
	Snippet  string `json:"snippet"`
	RelPath  string `json:"relPath,omitempty"`
	Color    string `json:"color,omitempty"`
	Emoji    string `json:"emoji,omitempty"`
}

type AppConfig struct {
	LastDossierPath string   `json:"lastDossierPath"`
	RecentPaths     []string `json:"recentPaths,omitempty"`
	// DismissedIntros: view keys the user has dismissed (documents, notes, …).
	DismissedIntros []string `json:"dismissedIntros,omitempty"`
	// Settings is small app-wide prefs (not per-dossier).
	Settings AppSettings `json:"settings"`
}

// AppSettings holds global UI/behavior preferences.
type AppSettings struct {
	// NoteAutosaveMs debounce for notes editor (default 500).
	NoteAutosaveMs int `json:"noteAutosaveMs,omitempty"`
	// ConfirmDeletes shows confirm dialogs before destructive actions (default true).
	ConfirmDeletes *bool `json:"confirmDeletes,omitempty"`
	// ShowIntros enables first-open surface intros until dismissed (default true).
	ShowIntros *bool `json:"showIntros,omitempty"`
	// Theme is reserved; currently always dark shell.
	Theme string `json:"theme,omitempty"`

	// Optional local agent (client-only). When Host is empty the Ask panel is hidden.
	AgentHost  string `json:"agentHost,omitempty"`
	AgentPort  int    `json:"agentPort,omitempty"`
	AgentToken string `json:"agentToken,omitempty"`
	// AgentPath is the HTTP path on the agent host (default /v1/chat/completions).
	AgentPath string `json:"agentPath,omitempty"`

	// OpenWith maps file extension (".pdf") to a preferred app executable path.
	// Independent of the Windows default association; used by Open externally.
	OpenWith map[string]string `json:"openWith,omitempty"`

	// AutoOpenLast, when true, opens LastDossierPath on launch if it is still a
	// valid dossier folder. Default false (launcher shown). Separate from the
	// DOSSIER_AUTO_OPEN=1 env force-on used for debug/automation.
	AutoOpenLast *bool `json:"autoOpenLast,omitempty"`
}

// ExtKey normalizes a filename or extension to ".ext" lowercase for OpenWith keys.
func ExtKey(nameOrExt string) string {
	s := strings.TrimSpace(strings.ToLower(nameOrExt))
	if s == "" {
		return ""
	}
	// Prefer filepath.Ext for names like "foo.pdf" or paths; bare "pdf" / ".pdf" also work.
	if strings.Contains(s, string(filepath.Separator)) || strings.Contains(s, "/") || strings.Contains(s[1:], ".") {
		if e := filepath.Ext(s); e != "" {
			s = e
		}
	}
	if s == "" {
		return ""
	}
	if !strings.HasPrefix(s, ".") {
		s = "." + s
	}
	// ".foo.pdf" shouldn't happen; if it does, keep last ext
	if strings.Count(s, ".") > 1 {
		s = filepath.Ext(s)
		if s == "" {
			return ""
		}
	}
	return s
}

func (s AppSettings) ConfirmDeletesOrDefault() bool {
	if s.ConfirmDeletes == nil {
		return true
	}
	return *s.ConfirmDeletes
}

func (s AppSettings) ShowIntrosOrDefault() bool {
	if s.ShowIntros == nil {
		return true
	}
	return *s.ShowIntros
}

func (s AppSettings) NoteAutosaveMsOrDefault() int {
	if s.NoteAutosaveMs <= 0 {
		return 500
	}
	return s.NoteAutosaveMs
}

// AutoOpenLastOrDefault is false when unset (launcher is the normal double-click UX).
func (s AppSettings) AutoOpenLastOrDefault() bool {
	if s.AutoOpenLast == nil {
		return false
	}
	return *s.AutoOpenLast
}

// TouchRecent records path as last-opened and keeps a short recents list.
func (c *AppConfig) TouchRecent(path string) {
	path = filepath.Clean(path)
	c.LastDossierPath = path
	out := []string{path}
	for _, p := range c.RecentPaths {
		p = filepath.Clean(p)
		if p == "" || strings.EqualFold(p, path) {
			continue
		}
		out = append(out, p)
		if len(out) >= 12 {
			break
		}
	}
	c.RecentPaths = out
}

// Canvas is a Paint or Annotate board (state JSON + optional PNG under boards/).
type Canvas struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"` // paint | annotate
	Title     string `json:"title"`
	StateJSON string `json:"stateJson"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// KanbanBoard stores columns+cards as one JSON document (editable, portable).
type KanbanBoard struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	StateJSON string `json:"stateJson"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// Decision is a pinned decision on the chronological spine.
type Decision struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Body           string `json:"body"`
	DecidedAt      string `json:"decidedAt"`
	DocumentID     string `json:"documentId,omitempty"`
	DocVersionRel  string `json:"docVersionRel,omitempty"`
	DocVersionNote string `json:"docVersionNote,omitempty"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

func nowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func AppConfigPath() (string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		appData = filepath.Join(home, "AppData", "Roaming")
	}
	dir := filepath.Join(appData, AppDataRelPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, ConfigFileName), nil
}

func LoadConfig() (*AppConfig, error) {
	path, err := AppConfigPath()
	if err != nil {
		return &AppConfig{}, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &AppConfig{}, nil
		}
		return &AppConfig{}, err
	}
	var c AppConfig
	if err := json.Unmarshal(b, &c); err != nil {
		return &AppConfig{}, err
	}
	return &c, nil
}

func SaveConfig(c *AppConfig) error {
	path, err := AppConfigPath()
	if err != nil {
		return err
	}
	if c.RecentPaths == nil {
		c.RecentPaths = []string{}
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// ListRecentDossiers returns known recent paths that still look like dossiers.
func ListRecentDossiers() []string {
	cfg, err := LoadConfig()
	if err != nil {
		return []string{}
	}
	seen := map[string]bool{}
	var out []string
	candidates := append([]string{}, cfg.RecentPaths...)
	if cfg.LastDossierPath != "" {
		candidates = append([]string{cfg.LastDossierPath}, candidates...)
	}
	for _, p := range candidates {
		p = filepath.Clean(p)
		if p == "" || seen[strings.ToLower(p)] {
			continue
		}
		if !IsDossierFolder(p) {
			continue
		}
		seen[strings.ToLower(p)] = true
		out = append(out, p)
	}
	return out
}

// OpenOrCreate opens an existing dossier folder or initializes a new one.
func OpenOrCreate(root string) (*Store, error) {
	root = filepath.Clean(root)
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("create dossier folder: %w", err)
	}
	for _, sub := range []string{DocumentsDir, NotesDir, BoardsDir, VersionsDir} {
		if err := os.MkdirAll(filepath.Join(root, sub), 0o755); err != nil {
			return nil, err
		}
	}
	marker := filepath.Join(root, MarkerFileName)
	if _, err := os.Stat(marker); os.IsNotExist(err) {
		content := "Dossier workspace\nCreated: " + nowISO() + "\n"
		if err := os.WriteFile(marker, []byte(content), 0o644); err != nil {
			return nil, err
		}
	}

	dbPath := filepath.Join(root, DBFileName)
	// modernc sqlite: enable FTS5 (built-in)
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)", filepath.ToSlash(dbPath))
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)

	s := &Store{db: db, Root: root}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	return err
}

func (s *Store) migrate() error {
	schema := `
CREATE TABLE IF NOT EXISTS meta (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS documents (
  id TEXT PRIMARY KEY,
  filename TEXT NOT NULL,
  rel_path TEXT NOT NULL UNIQUE,
  mime_type TEXT,
  size_bytes INTEGER NOT NULL DEFAULT 0,
  kind TEXT NOT NULL DEFAULT 'other',
  pdf_flags TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  mtime TEXT,
  indexed_at TEXT,
  preview TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS notes (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  rel_path TEXT NOT NULL UNIQUE,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS stickies (
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
  due_at TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS canvases (
  id TEXT PRIMARY KEY,
  kind TEXT NOT NULL,
  title TEXT NOT NULL,
  state_json TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS kanban_boards (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  state_json TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS decisions (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  body TEXT NOT NULL DEFAULT '',
  decided_at TEXT NOT NULL,
  document_id TEXT,
  doc_version_rel TEXT,
  doc_version_note TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE VIRTUAL TABLE IF NOT EXISTS docs_fts USING fts5(
  doc_id UNINDEXED,
  kind UNINDEXED,
  title,
  body,
  tokenize = 'porter unicode61'
);
`
	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	// Tier-1 DBs may lack preview column
	_, _ = s.db.Exec(`ALTER TABLE documents ADD COLUMN preview TEXT NOT NULL DEFAULT ''`)
	// Schema v3: sticky due_at for a later Calendar tier (unused in UI this release).
	// ALTER is ignored when the column already exists (SQLite returns error; we drop it).
	_, _ = s.db.Exec(`ALTER TABLE stickies ADD COLUMN due_at TEXT`)
	_ = s.setMeta("schema_version", SchemaVersion)
	return nil
}

func (s *Store) AbsPath(rel string) string {
	return filepath.Join(s.Root, filepath.FromSlash(rel))
}

func (s *Store) setMeta(key, value string) error {
	_, err := s.db.Exec(`INSERT INTO meta(key,value) VALUES(?,?)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value`, key, value)
	return err
}

func (s *Store) getMeta(key string) (string, error) {
	var v string
	err := s.db.QueryRow(`SELECT value FROM meta WHERE key=?`, key).Scan(&v)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return v, err
}

// ---- Documents ----

func (s *Store) UpsertDocument(d *Document, body string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	pdfJSON := ""
	if d.PDFFlags != nil {
		b, _ := json.Marshal(d.PDFFlags)
		pdfJSON = string(b)
	}
	// Store office extract flags alongside pdf_flags as a tagged envelope when present
	// so we don't need a schema migration (Tier-1 DBs keep working).
	if d.OfficeFlags != nil {
		env := map[string]interface{}{"office": d.OfficeFlags}
		if d.PDFFlags != nil {
			env["pdf"] = d.PDFFlags
		}
		b, _ := json.Marshal(env)
		pdfJSON = string(b)
	}
	if d.Preview == "" {
		d.Preview = MakePreview(body, 140)
	}
	_, err := s.db.Exec(`
INSERT INTO documents(id, filename, rel_path, mime_type, size_bytes, kind, pdf_flags, created_at, updated_at, mtime, indexed_at, preview)
VALUES(?,?,?,?,?,?,?,?,?,?,?,?)
ON CONFLICT(id) DO UPDATE SET
  filename=excluded.filename,
  rel_path=excluded.rel_path,
  mime_type=excluded.mime_type,
  size_bytes=excluded.size_bytes,
  kind=excluded.kind,
  pdf_flags=excluded.pdf_flags,
  updated_at=excluded.updated_at,
  mtime=excluded.mtime,
  indexed_at=excluded.indexed_at,
  preview=excluded.preview
`, d.ID, d.Filename, d.RelPath, d.MimeType, d.SizeBytes, d.Kind, nullIfEmpty(pdfJSON),
		d.CreatedAt, d.UpdatedAt, d.Mtime, d.IndexedAt, d.Preview)
	if err != nil {
		return err
	}
	return s.reindexFTS(d.ID, "document", d.Filename, body)
}

// MakePreview collapses whitespace and truncates for list rows.
func MakePreview(body string, max int) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}
	// Flatten whitespace / newlines
	var b strings.Builder
	space := false
	for _, r := range body {
		if r == '\n' || r == '\r' || r == '\t' || r == ' ' {
			if !space && b.Len() > 0 {
				b.WriteByte(' ')
				space = true
			}
			continue
		}
		space = false
		b.WriteRune(r)
		if max > 0 && b.Len() >= max+1 {
			break
		}
	}
	out := strings.TrimSpace(b.String())
	if max > 0 && len(out) > max {
		// rune-safe truncate
		runes := []rune(out)
		if len(runes) > max {
			out = string(runes[:max]) + "…"
		}
	}
	return out
}

func (s *Store) GetDocument(id string) (*Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	row := s.db.QueryRow(`SELECT id, filename, rel_path, mime_type, size_bytes, kind, pdf_flags, created_at, updated_at, mtime, COALESCE(indexed_at,''), COALESCE(preview,'')
		FROM documents WHERE id=?`, id)
	return scanDocument(row)
}

func (s *Store) GetDocumentByRelPath(rel string) (*Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	row := s.db.QueryRow(`SELECT id, filename, rel_path, mime_type, size_bytes, kind, pdf_flags, created_at, updated_at, mtime, COALESCE(indexed_at,''), COALESCE(preview,'')
		FROM documents WHERE rel_path=?`, rel)
	return scanDocument(row)
}

func (s *Store) ListDocuments() ([]Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows, err := s.db.Query(`SELECT id, filename, rel_path, mime_type, size_bytes, kind, pdf_flags, created_at, updated_at, mtime, COALESCE(indexed_at,''), COALESCE(preview,'')
		FROM documents ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Document
	for rows.Next() {
		d, err := scanDocumentRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *d)
	}
	if out == nil {
		out = []Document{}
	}
	return out, rows.Err()
}

func (s *Store) DeleteDocument(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var rel string
	err := s.db.QueryRow(`SELECT rel_path FROM documents WHERE id=?`, id).Scan(&rel)
	if err != nil {
		return err
	}
	if _, err := s.db.Exec(`DELETE FROM documents WHERE id=?`, id); err != nil {
		return err
	}
	_ = s.removeFTS(id)
	abs := s.AbsPath(rel)
	_ = os.Remove(abs)
	return nil
}

func scanDocument(row *sql.Row) (*Document, error) {
	var d Document
	var pdf sql.NullString
	err := row.Scan(&d.ID, &d.Filename, &d.RelPath, &d.MimeType, &d.SizeBytes, &d.Kind, &pdf,
		&d.CreatedAt, &d.UpdatedAt, &d.Mtime, &d.IndexedAt, &d.Preview)
	if err != nil {
		return nil, err
	}
	parseExtractFlags(&d, pdf)
	return &d, nil
}

func scanDocumentRows(rows *sql.Rows) (*Document, error) {
	var d Document
	var pdf sql.NullString
	err := rows.Scan(&d.ID, &d.Filename, &d.RelPath, &d.MimeType, &d.SizeBytes, &d.Kind, &pdf,
		&d.CreatedAt, &d.UpdatedAt, &d.Mtime, &d.IndexedAt, &d.Preview)
	if err != nil {
		return nil, err
	}
	parseExtractFlags(&d, pdf)
	return &d, nil
}

func parseExtractFlags(d *Document, pdf sql.NullString) {
	if !pdf.Valid || pdf.String == "" {
		return
	}
	raw := []byte(pdf.String)
	// Envelope with office and/or pdf
	var env struct {
		Office *OfficeFlags `json:"office"`
		PDF    *PDFFlags    `json:"pdf"`
	}
	if json.Unmarshal(raw, &env) == nil && (env.Office != nil || env.PDF != nil) {
		d.OfficeFlags = env.Office
		d.PDFFlags = env.PDF
		return
	}
	var f PDFFlags
	if json.Unmarshal(raw, &f) == nil {
		d.PDFFlags = &f
	}
}

// ---- Notes ----

func (s *Store) UpsertNote(n *Note, body string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`
INSERT INTO notes(id, title, rel_path, created_at, updated_at) VALUES(?,?,?,?,?)
ON CONFLICT(id) DO UPDATE SET title=excluded.title, rel_path=excluded.rel_path, updated_at=excluded.updated_at
`, n.ID, n.Title, n.RelPath, n.CreatedAt, n.UpdatedAt)
	if err != nil {
		return err
	}
	return s.reindexFTS(n.ID, "note", n.Title, body)
}

func (s *Store) ListNotes() ([]Note, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows, err := s.db.Query(`SELECT id, title, rel_path, created_at, updated_at FROM notes ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Title, &n.RelPath, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	if out == nil {
		out = []Note{}
	}
	return out, rows.Err()
}

func (s *Store) GetNote(id string) (*Note, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var n Note
	err := s.db.QueryRow(`SELECT id, title, rel_path, created_at, updated_at FROM notes WHERE id=?`, id).
		Scan(&n.ID, &n.Title, &n.RelPath, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(s.AbsPath(n.RelPath))
	if err == nil {
		n.Body = string(b)
	}
	return &n, nil
}

func (s *Store) DeleteNote(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var rel string
	if err := s.db.QueryRow(`SELECT rel_path FROM notes WHERE id=?`, id).Scan(&rel); err != nil {
		return err
	}
	if _, err := s.db.Exec(`DELETE FROM notes WHERE id=?`, id); err != nil {
		return err
	}
	_ = s.removeFTS(id)
	_ = os.Remove(s.AbsPath(rel))
	return nil
}

// ---- Stickies ----

func (s *Store) ListStickies() ([]Sticky, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows, err := s.db.Query(`SELECT id, text, color, size, x, y, w, h, emoji, z_index, created_at, updated_at
		FROM stickies ORDER BY z_index ASC, updated_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Sticky
	for rows.Next() {
		var st Sticky
		if err := rows.Scan(&st.ID, &st.Text, &st.Color, &st.Size, &st.X, &st.Y, &st.W, &st.H,
			&st.Emoji, &st.ZIndex, &st.CreatedAt, &st.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	if out == nil {
		out = []Sticky{}
	}
	return out, rows.Err()
}

func (s *Store) UpsertSticky(st *Sticky) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`
INSERT INTO stickies(id, text, color, size, x, y, w, h, emoji, z_index, created_at, updated_at)
VALUES(?,?,?,?,?,?,?,?,?,?,?,?)
ON CONFLICT(id) DO UPDATE SET
  text=excluded.text, color=excluded.color, size=excluded.size,
  x=excluded.x, y=excluded.y, w=excluded.w, h=excluded.h,
  emoji=excluded.emoji, z_index=excluded.z_index, updated_at=excluded.updated_at
`, st.ID, st.Text, st.Color, st.Size, st.X, st.Y, st.W, st.H, st.Emoji, st.ZIndex, st.CreatedAt, st.UpdatedAt)
	if err != nil {
		return err
	}
	title := st.Emoji
	if title != "" {
		title += " "
	}
	preview := st.Text
	if len(preview) > 40 {
		preview = preview[:40] + "…"
	}
	title += preview
	if title == "" {
		title = "Sticky note"
	}
	return s.reindexFTS(st.ID, "sticky", title, st.Text)
}

func (s *Store) DeleteSticky(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.db.Exec(`DELETE FROM stickies WHERE id=?`, id); err != nil {
		return err
	}
	return s.removeFTS(id)
}

func (s *Store) MaxStickyZ() (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var z sql.NullInt64
	err := s.db.QueryRow(`SELECT MAX(z_index) FROM stickies`).Scan(&z)
	if err != nil {
		return 0, err
	}
	if !z.Valid {
		return 0, nil
	}
	return int(z.Int64), nil
}

// ---- FTS ----

func (s *Store) reindexFTS(id, kind, title, body string) error {
	if _, err := s.db.Exec(`DELETE FROM docs_fts WHERE doc_id=?`, id); err != nil {
		return err
	}
	_, err := s.db.Exec(`INSERT INTO docs_fts(doc_id, kind, title, body) VALUES(?,?,?,?)`, id, kind, title, body)
	return err
}

func (s *Store) removeFTS(id string) error {
	_, err := s.db.Exec(`DELETE FROM docs_fts WHERE doc_id=?`, id)
	return err
}

func (s *Store) Search(query string, limit int) ([]SearchHit, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	query = strings.TrimSpace(query)
	if query == "" {
		return []SearchHit{}, nil
	}
	if limit <= 0 {
		limit = 50
	}
	// Build a safe FTS query: quote terms, AND them
	ftsQuery := buildFTSQuery(query)
	if ftsQuery == "" {
		return []SearchHit{}, nil
	}

	rows, err := s.db.Query(`
SELECT doc_id, kind, title,
  snippet(docs_fts, 3, '«', '»', '…', 24) AS snip
FROM docs_fts
WHERE docs_fts MATCH ?
ORDER BY rank
LIMIT ?`, ftsQuery, limit)
	if err != nil {
		// Fallback: simple LIKE if FTS query fails
		return s.searchLike(query, limit)
	}
	defer rows.Close()
	var hits []SearchHit
	for rows.Next() {
		var h SearchHit
		if err := rows.Scan(&h.ID, &h.Kind, &h.Title, &h.Snippet); err != nil {
			return nil, err
		}
		hits = append(hits, h)
	}
	if hits == nil {
		hits = []SearchHit{}
	}
	// Enrich with paths/colors
	for i := range hits {
		switch hits[i].Kind {
		case "document":
			var rel string
			_ = s.db.QueryRow(`SELECT rel_path FROM documents WHERE id=?`, hits[i].ID).Scan(&rel)
			hits[i].RelPath = rel
		case "note":
			var rel string
			_ = s.db.QueryRow(`SELECT rel_path FROM notes WHERE id=?`, hits[i].ID).Scan(&rel)
			hits[i].RelPath = rel
		case "sticky":
			var color, emoji string
			_ = s.db.QueryRow(`SELECT color, emoji FROM stickies WHERE id=?`, hits[i].ID).Scan(&color, &emoji)
			hits[i].Color = color
			hits[i].Emoji = emoji
		}
	}
	return hits, rows.Err()
}

func (s *Store) searchLike(query string, limit int) ([]SearchHit, error) {
	q := "%" + query + "%"
	rows, err := s.db.Query(`
SELECT doc_id, kind, title, substr(body,1,120) FROM docs_fts
WHERE title LIKE ? OR body LIKE ?
LIMIT ?`, q, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var hits []SearchHit
	for rows.Next() {
		var h SearchHit
		if err := rows.Scan(&h.ID, &h.Kind, &h.Title, &h.Snippet); err != nil {
			return nil, err
		}
		hits = append(hits, h)
	}
	if hits == nil {
		hits = []SearchHit{}
	}
	return hits, nil
}

func buildFTSQuery(q string) string {
	// Split on whitespace; quote each token to avoid FTS syntax errors.
	// Space-separated FTS5 terms are AND-ed (good for the Search box as you type).
	parts := strings.Fields(q)
	var terms []string
	for _, p := range parts {
		p = strings.ReplaceAll(p, `"`, "")
		p = strings.Trim(p, "*")
		if p == "" {
			continue
		}
		// Prefix match for partial typing
		terms = append(terms, `"`+p+`"*`)
	}
	return strings.Join(terms, " ")
}

// englishStopwords for natural-language → FTS (agent / loose search).
var englishStopwords = map[string]bool{
	"a": true, "an": true, "the": true, "and": true, "or": true, "but": true,
	"is": true, "are": true, "was": true, "were": true, "be": true, "been": true, "being": true,
	"am": true, "do": true, "does": true, "did": true, "doing": true,
	"have": true, "has": true, "had": true, "having": true,
	"what": true, "which": true, "who": true, "whom": true, "whose": true,
	"where": true, "when": true, "why": true, "how": true,
	"this": true, "that": true, "these": true, "those": true,
	"in": true, "on": true, "at": true, "to": true, "for": true, "of": true, "from": true,
	"with": true, "by": true, "as": true, "into": true, "about": true, "over": true,
	"it": true, "its": true, "i": true, "me": true, "my": true, "we": true, "our": true,
	"you": true, "your": true, "they": true, "them": true, "their": true,
	"can": true, "could": true, "would": true, "should": true, "will": true, "just": true,
	"not": true, "no": true, "yes": true, "if": true, "then": true, "than": true,
	"so": true, "too": true, "very": true, "also": true, "any": true, "all": true,
	"please": true, "tell": true, "show": true, "find": true, "give": true, "get": true,
	"there": true, "here": true, "file": true, "document": true, "note": true, "dossier": true,
}

// meaningfulSearchTerms extracts content-bearing tokens from a natural-language question.
func meaningfulSearchTerms(q string) []string {
	// Split on non-alphanumeric (keeps FTS5-safe tokens)
	fields := strings.FieldsFunc(strings.ToLower(q), func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-')
	})
	var terms []string
	seen := map[string]bool{}
	for _, p := range fields {
		p = strings.Trim(p, "-_")
		if len(p) < 2 {
			continue
		}
		if englishStopwords[p] {
			continue
		}
		if seen[p] {
			continue
		}
		seen[p] = true
		terms = append(terms, p)
	}
	return terms
}

// buildFTSQueryLoose builds an FTS5 OR query over meaningful terms (for natural-language questions).
// Example: "What is the FTS5 test keyword in the plain-text file?" → "fts5"* OR "test"* OR "keyword"* OR "plain"* OR "text"*
func buildFTSQueryLoose(q string) string {
	terms := meaningfulSearchTerms(q)
	if len(terms) == 0 {
		// Fall back to strict builder without stopwords if everything was stopwords
		return buildFTSQuery(q)
	}
	var parts []string
	for _, p := range terms {
		p = strings.ReplaceAll(p, `"`, "")
		if p == "" {
			continue
		}
		parts = append(parts, `"`+p+`"*`)
	}
	return strings.Join(parts, " OR ")
}

// SearchLoose runs FTS with stopword stripping + OR matching (agent context / NL questions),
// ranked by bm25() (best first) and limited to `limit` rows.
// Falls back to strict Search if loose MATCH fails or returns nothing.
func (s *Store) SearchLoose(query string, limit int) ([]SearchHit, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []SearchHit{}, nil
	}
	if limit <= 0 {
		limit = 4
	}
	ftsQuery := buildFTSQueryLoose(query)
	if ftsQuery == "" {
		return s.Search(query, limit)
	}

	hits, err := s.searchFTSMatch(ftsQuery, limit)
	if err != nil || len(hits) == 0 {
		// Strict AND (good for bare keywords) or LIKE
		if hits2, err2 := s.Search(query, limit); err2 == nil && len(hits2) > 0 {
			return hits2, nil
		}
		if err != nil {
			return s.searchLike(query, limit)
		}
		return hits, nil
	}
	return hits, nil
}

// searchFTSMatch runs a pre-built FTS MATCH string, orders by bm25 (more negative = better),
// applies LIMIT, and enriches hits.
func (s *Store) searchFTSMatch(ftsQuery string, limit int) ([]SearchHit, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// bm25() is FTS5 Okapi BM25; ascending order puts strongest matches first.
	rows, err := s.db.Query(`
SELECT doc_id, kind, title,
  snippet(docs_fts, 3, '«', '»', '…', 24) AS snip
FROM docs_fts
WHERE docs_fts MATCH ?
ORDER BY bm25(docs_fts)
LIMIT ?`, ftsQuery, limit)
	if err != nil {
		// Fallback if bm25 unavailable
		rows, err = s.db.Query(`
SELECT doc_id, kind, title,
  snippet(docs_fts, 3, '«', '»', '…', 24) AS snip
FROM docs_fts
WHERE docs_fts MATCH ?
ORDER BY rank
LIMIT ?`, ftsQuery, limit)
		if err != nil {
			return nil, err
		}
	}
	defer rows.Close()
	var hits []SearchHit
	for rows.Next() {
		var h SearchHit
		if err := rows.Scan(&h.ID, &h.Kind, &h.Title, &h.Snippet); err != nil {
			return nil, err
		}
		hits = append(hits, h)
	}
	if hits == nil {
		hits = []SearchHit{}
	}
	for i := range hits {
		switch hits[i].Kind {
		case "document":
			var rel string
			_ = s.db.QueryRow(`SELECT rel_path FROM documents WHERE id=?`, hits[i].ID).Scan(&rel)
			hits[i].RelPath = rel
		case "note":
			var rel string
			_ = s.db.QueryRow(`SELECT rel_path FROM notes WHERE id=?`, hits[i].ID).Scan(&rel)
			hits[i].RelPath = rel
		case "sticky":
			var color, emoji string
			_ = s.db.QueryRow(`SELECT color, emoji FROM stickies WHERE id=?`, hits[i].ID).Scan(&color, &emoji)
			hits[i].Color = color
			hits[i].Emoji = emoji
		}
	}
	return hits, rows.Err()
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// IsDossierFolder reports whether path looks like a dossier workspace.
func IsDossierFolder(path string) bool {
	if path == "" {
		return false
	}
	if st, err := os.Stat(path); err != nil || !st.IsDir() {
		return false
	}
	if _, err := os.Stat(filepath.Join(path, MarkerFileName)); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(path, DBFileName)); err == nil {
		return true
	}
	return false
}
