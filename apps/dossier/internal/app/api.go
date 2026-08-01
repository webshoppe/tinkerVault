package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/webshoppe/dossier/internal/dialog"
	"github.com/webshoppe/dossier/internal/dossier"
)

// API is the set of methods bound into the webview JavaScript context.
type API struct {
	mu      sync.Mutex
	store   *dossier.Store
	watcher *dossier.Watcher
}

func NewAPI() *API {
	return &API{}
}

func (a *API) Close() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.watcher != nil {
		_ = a.watcher.Close()
		a.watcher = nil
	}
	if a.store != nil {
		_ = a.store.Close()
		a.store = nil
	}
}

// --- Setup / dossier lifecycle ---

type Status struct {
	Open         bool     `json:"open"`
	Root         string   `json:"root"`
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	HasLastPath  bool     `json:"hasLastPath"`
	LastPath     string   `json:"lastPath"`
	RecentPaths  []string `json:"recentPaths"`
	RecentNames  []string `json:"recentNames"`
}

func (a *API) GetStatus() Status {
	cfg, _ := dossier.LoadConfig()
	recents := dossier.ListRecentDossiers()
	names := make([]string, len(recents))
	for i, p := range recents {
		names[i] = filepath.Base(p)
	}
	st := Status{
		Version:     Version,
		HasLastPath: cfg.LastDossierPath != "" && dossier.IsDossierFolder(cfg.LastDossierPath),
		LastPath:    cfg.LastDossierPath,
		RecentPaths: recents,
		RecentNames: names,
	}
	if st.RecentPaths == nil {
		st.RecentPaths = []string{}
	}
	if st.RecentNames == nil {
		st.RecentNames = []string{}
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.store != nil {
		st.Open = true
		st.Root = a.store.Root
		st.Name = filepath.Base(a.store.Root)
	}
	return st
}

func (a *API) OpenLastDossier() (Status, error) {
	cfg, err := dossier.LoadConfig()
	if err != nil || cfg.LastDossierPath == "" {
		return a.GetStatus(), fmt.Errorf("no previous dossier")
	}
	return a.openPath(cfg.LastDossierPath)
}

func (a *API) PickAndOpenDossier() (Status, error) {
	path, err := dialog.PickFolder("Open or create a Dossier folder")
	if err != nil {
		return a.GetStatus(), err
	}
	if path == "" {
		return a.GetStatus(), fmt.Errorf("cancelled")
	}
	return a.openPath(path)
}

func (a *API) OpenDossierPath(path string) (Status, error) {
	if path == "" {
		return a.GetStatus(), fmt.Errorf("path required")
	}
	return a.openPath(path)
}

func (a *API) CreateDossier() (Status, error) {
	path, err := dialog.PickFolder("Choose folder for new Dossier (will be initialized here)")
	if err != nil {
		return a.GetStatus(), err
	}
	if path == "" {
		return a.GetStatus(), fmt.Errorf("cancelled")
	}
	return a.openPath(path)
}

// ListRecentDossiers returns recents for the switcher UI.
func (a *API) ListRecentDossiers() []map[string]string {
	paths := dossier.ListRecentDossiers()
	out := make([]map[string]string, 0, len(paths))
	for _, p := range paths {
		out = append(out, map[string]string{"path": p, "name": filepath.Base(p)})
	}
	return out
}

// SwitchDossier opens another workspace (closes current).
func (a *API) SwitchDossier(path string) (Status, error) {
	return a.OpenDossierPath(path)
}

// CloseDossier unloads the current workspace and returns to setup.
func (a *API) CloseDossier() Status {
	a.mu.Lock()
	if a.watcher != nil {
		_ = a.watcher.Close()
		a.watcher = nil
	}
	if a.store != nil {
		_ = a.store.Close()
		a.store = nil
	}
	a.mu.Unlock()
	return a.GetStatus()
}

func (a *API) openPath(path string) (Status, error) {
	a.mu.Lock()
	if a.watcher != nil {
		_ = a.watcher.Close()
		a.watcher = nil
	}
	if a.store != nil {
		_ = a.store.Close()
		a.store = nil
	}
	a.mu.Unlock()

	store, err := dossier.OpenOrCreate(path)
	if err != nil {
		return a.GetStatus(), err
	}
	_, _ = store.RescanDocuments()
	_, _ = store.RescanNotes()

	w, err := dossier.NewWatcher(store, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "watcher unavailable: %v\n", err)
	}

	a.mu.Lock()
	a.store = store
	a.watcher = w
	a.mu.Unlock()

	cfg, _ := dossier.LoadConfig()
	if cfg == nil {
		cfg = &dossier.AppConfig{}
	}
	cfg.TouchRecent(store.Root)
	_ = dossier.SaveConfig(cfg)
	return a.GetStatus(), nil
}

func (a *API) requireStore() (*dossier.Store, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.store == nil {
		return nil, fmt.Errorf("no dossier open")
	}
	return a.store, nil
}

// --- Documents ---

func (a *API) ListDocuments() ([]dossier.Document, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	return s.ListDocuments()
}

func (a *API) ImportFiles() ([]dossier.Document, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	paths, err := dialog.PickFiles("Import files into Dossier")
	if err != nil {
		return nil, err
	}
	var imported []dossier.Document
	for _, p := range paths {
		d, err := s.ImportFile(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "import %s: %v\n", p, err)
			continue
		}
		imported = append(imported, *d)
	}
	if imported == nil {
		imported = []dossier.Document{}
	}
	return imported, nil
}

// DroppedFile is one file from HTML5 drag-and-drop (content as base64).
type DroppedFile struct {
	Filename string `json:"filename"`
	DataB64  string `json:"dataB64"`
}

// ImportDropped imports files dropped onto the window (same path as ImportFiles).
func (a *API) ImportDropped(files []DroppedFile) ([]dossier.Document, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	var imported []dossier.Document
	for _, f := range files {
		if f.Filename == "" || f.DataB64 == "" {
			// empty DataB64 may still be empty file; allow if filename set
			if f.Filename == "" {
				continue
			}
		}
		d, err := s.ImportBase64(f.Filename, f.DataB64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "import dropped %s: %v\n", f.Filename, err)
			continue
		}
		imported = append(imported, *d)
	}
	if imported == nil {
		imported = []dossier.Document{}
	}
	return imported, nil
}

func (a *API) GetDocument(id string) (map[string]interface{}, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	body, d, err := s.ReadDocumentBody(id)
	if err != nil && d == nil {
		return nil, err
	}
	out := map[string]interface{}{
		"document": d,
		"body":     body,
	}
	return out, nil
}

func (a *API) DeleteDocument(id string) error {
	s, err := a.requireStore()
	if err != nil {
		return err
	}
	return s.DeleteDocument(id)
}

// OpenDocumentExternally opens a document with Dossier's preferred app for that
// extension when set; otherwise the Windows default association.
// forceDefault=true always uses the OS default (ignores preferred).
func (a *API) OpenDocumentExternally(id string, forceDefault bool) error {
	s, err := a.requireStore()
	if err != nil {
		return err
	}
	d, err := s.GetDocument(id)
	if err != nil {
		return err
	}
	abs := s.AbsPath(d.RelPath)
	if !forceDefault {
		if app := a.preferredAppFor(d.Filename); app != "" {
			return dialog.OpenWithApp(abs, app)
		}
	}
	return dialog.OpenExternally(abs)
}

// OpenDocumentWithApp launches path with a one-shot chosen executable (does not save preference).
func (a *API) OpenDocumentWithApp(id, appExe string) error {
	s, err := a.requireStore()
	if err != nil {
		return err
	}
	d, err := s.GetDocument(id)
	if err != nil {
		return err
	}
	appExe = strings.TrimSpace(appExe)
	if appExe == "" {
		return dialog.OpenExternally(s.AbsPath(d.RelPath))
	}
	return dialog.OpenWithApp(s.AbsPath(d.RelPath), appExe)
}

// PickOpenApp shows an executable picker; returns path or empty if cancelled.
func (a *API) PickOpenApp() (string, error) {
	p, err := dialog.PickExecutable("Choose application for Open externally")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(p), nil
}

// SetOpenWithPreference remembers preferred app for an extension (".pdf" or filename).
// Empty appExe clears the preference.
func (a *API) SetOpenWithPreference(extOrName, appExe string) (map[string]interface{}, error) {
	key := dossier.ExtKey(extOrName)
	if key == "" {
		return a.GetAppSettings(), fmt.Errorf("extension required")
	}
	cfg, err := dossier.LoadConfig()
	if err != nil || cfg == nil {
		cfg = &dossier.AppConfig{}
	}
	if cfg.Settings.OpenWith == nil {
		cfg.Settings.OpenWith = map[string]string{}
	}
	appExe = strings.TrimSpace(appExe)
	if appExe == "" {
		delete(cfg.Settings.OpenWith, key)
	} else {
		cfg.Settings.OpenWith[key] = appExe
	}
	if err := dossier.SaveConfig(cfg); err != nil {
		return nil, err
	}
	return a.GetAppSettings(), nil
}

func (a *API) preferredAppFor(filename string) string {
	cfg, _ := dossier.LoadConfig()
	if cfg == nil || cfg.Settings.OpenWith == nil {
		return ""
	}
	key := dossier.ExtKey(filename)
	if key == "" {
		return ""
	}
	return strings.TrimSpace(cfg.Settings.OpenWith[key])
}

func (a *API) Rescan() (map[string]int, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	dc, err := s.RescanDocuments()
	if err != nil {
		return nil, err
	}
	nc, err := s.RescanNotes()
	if err != nil {
		return nil, err
	}
	return map[string]int{"documents": dc, "notes": nc}, nil
}

// --- Notes ---

func (a *API) ListNotes() ([]dossier.Note, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	return s.ListNotes()
}

func (a *API) CreateNote(title string) (*dossier.Note, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	return s.CreateNote(title, "")
}

func (a *API) GetNote(id string) (*dossier.Note, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	return s.GetNote(id)
}

func (a *API) SaveNote(id, title, body string) (*dossier.Note, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	return s.SaveNote(id, title, body)
}

func (a *API) DeleteNote(id string) error {
	s, err := a.requireStore()
	if err != nil {
		return err
	}
	return s.DeleteNote(id)
}

// --- Stickies ---

func (a *API) ListStickies() ([]dossier.Sticky, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	return s.ListStickies()
}

func (a *API) CreateSticky(color, size, emoji string) (*dossier.Sticky, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	return s.CreateSticky(&dossier.Sticky{Color: color, Size: size, Emoji: emoji})
}

func (a *API) UpdateSticky(st dossier.Sticky) (*dossier.Sticky, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	return s.UpdateSticky(st.ID, &st)
}

func (a *API) DeleteSticky(id string) error {
	s, err := a.requireStore()
	if err != nil {
		return err
	}
	return s.DeleteSticky(id)
}

func (a *API) StickyMeta() map[string]interface{} {
	return map[string]interface{}{
		"colors": dossier.StickyColors,
		"sizes":  dossier.StickySizes,
	}
}

// --- Paint / Annotate canvases ---

func (a *API) ListCanvases(kind string) ([]dossier.Canvas, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	return s.ListCanvases(kind)
}

func (a *API) CreateCanvas(kind, title string) (*dossier.Canvas, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	return s.CreateCanvas(kind, title)
}

func (a *API) GetCanvas(id string) (*dossier.Canvas, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	return s.GetCanvas(id)
}

func (a *API) SaveCanvas(id, title, stateJSON string) (*dossier.Canvas, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	return s.SaveCanvas(id, title, stateJSON)
}

func (a *API) DeleteCanvas(id string) error {
	s, err := a.requireStore()
	if err != nil {
		return err
	}
	return s.DeleteCanvas(id)
}

// --- Kanban ---

func (a *API) ListKanbanBoards() ([]dossier.KanbanBoard, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	return s.ListKanbanBoards()
}

func (a *API) CreateKanbanBoard(title string) (*dossier.KanbanBoard, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	return s.CreateKanbanBoard(title)
}

func (a *API) GetKanbanBoard(id string) (*dossier.KanbanBoard, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	return s.GetKanbanBoard(id)
}

func (a *API) SaveKanbanBoard(id, title, stateJSON string) (*dossier.KanbanBoard, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	return s.SaveKanbanBoard(id, title, stateJSON)
}

func (a *API) DeleteKanbanBoard(id string) error {
	s, err := a.requireStore()
	if err != nil {
		return err
	}
	return s.DeleteKanbanBoard(id)
}

// --- Decisions ---

func (a *API) ListDecisions() ([]dossier.Decision, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	return s.ListDecisions()
}

func (a *API) CreateDecision(title, body, decidedAt string) (*dossier.Decision, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	return s.CreateDecision(title, body, decidedAt)
}

func (a *API) UpdateDecision(d dossier.Decision) (*dossier.Decision, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	return s.UpdateDecision(&d)
}

func (a *API) AttachDocumentVersion(decisionID, documentID, note string) (*dossier.Decision, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	return s.AttachDocumentVersion(decisionID, documentID, note)
}

func (a *API) DeleteDecision(id string) error {
	s, err := a.requireStore()
	if err != nil {
		return err
	}
	return s.DeleteDecision(id)
}

func (a *API) OpenDecisionVersion(id string) error {
	s, err := a.requireStore()
	if err != nil {
		return err
	}
	d, err := s.GetDecision(id)
	if err != nil {
		return err
	}
	if d.DocVersionRel == "" {
		return fmt.Errorf("no version attached")
	}
	return dialog.OpenExternally(s.AbsPath(d.DocVersionRel))
}

// --- Search ---

func (a *API) Search(query string) ([]dossier.SearchHit, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	return s.Search(query, 60)
}

// --- Utility ---

func (a *API) FormatSize(n float64) string {
	return dossier.FormatBytes(int64(n))
}

func (a *API) OpenDossierFolder() error {
	s, err := a.requireStore()
	if err != nil {
		return err
	}
	return dialog.OpenExternally(s.Root)
}

// --- App settings & intros (global, outside dossier) ---

func (a *API) GetAppSettings() map[string]interface{} {
	cfg, _ := dossier.LoadConfig()
	if cfg == nil {
		cfg = &dossier.AppConfig{}
	}
	confirm := cfg.Settings.ConfirmDeletesOrDefault()
	showIntros := cfg.Settings.ShowIntrosOrDefault()
	dismissed := cfg.DismissedIntros
	if dismissed == nil {
		dismissed = []string{}
	}
	agentOn := AgentConfigured(cfg)
	path := strings.TrimSpace(cfg.Settings.AgentPath)
	if path == "" {
		path = "/v1/chat/completions"
	}
	openWith := cfg.Settings.OpenWith
	if openWith == nil {
		openWith = map[string]string{}
	}
	return map[string]interface{}{
		"noteAutosaveMs":  cfg.Settings.NoteAutosaveMsOrDefault(),
		"confirmDeletes":  confirm,
		"showIntros":      showIntros,
		"theme":           cfg.Settings.Theme,
		"dismissedIntros": dismissed,
		"agentHost":       cfg.Settings.AgentHost,
		"agentPort":       cfg.Settings.AgentPort,
		"agentToken":      cfg.Settings.AgentToken,
		"agentPath":       path,
		"agentConfigured": agentOn,
		"openWith":        openWith,
		"autoOpenLast":    cfg.Settings.AutoOpenLastOrDefault(),
	}
}

type SettingsPatch struct {
	NoteAutosaveMs  *int              `json:"noteAutosaveMs"`
	ConfirmDeletes  *bool             `json:"confirmDeletes"`
	ShowIntros      *bool             `json:"showIntros"`
	Theme           *string           `json:"theme"`
	DismissedIntros []string          `json:"dismissedIntros"`
	AgentHost       *string           `json:"agentHost"`
	AgentPort       *int              `json:"agentPort"`
	AgentToken      *string           `json:"agentToken"`
	AgentPath       *string           `json:"agentPath"`
	OpenWith        map[string]string `json:"openWith"`
	AutoOpenLast    *bool             `json:"autoOpenLast"`
}

func (a *API) SaveAppSettings(p SettingsPatch) (map[string]interface{}, error) {
	cfg, err := dossier.LoadConfig()
	if err != nil || cfg == nil {
		cfg = &dossier.AppConfig{}
	}
	if p.NoteAutosaveMs != nil {
		cfg.Settings.NoteAutosaveMs = *p.NoteAutosaveMs
	}
	if p.ConfirmDeletes != nil {
		cfg.Settings.ConfirmDeletes = p.ConfirmDeletes
	}
	if p.ShowIntros != nil {
		cfg.Settings.ShowIntros = p.ShowIntros
	}
	if p.AutoOpenLast != nil {
		cfg.Settings.AutoOpenLast = p.AutoOpenLast
	}
	if p.Theme != nil {
		cfg.Settings.Theme = *p.Theme
	}
	if p.DismissedIntros != nil {
		cfg.DismissedIntros = p.DismissedIntros
	}
	if p.AgentHost != nil {
		cfg.Settings.AgentHost = strings.TrimSpace(*p.AgentHost)
	}
	if p.AgentPort != nil {
		cfg.Settings.AgentPort = *p.AgentPort
	}
	// Host is the enable/hide gate (help text + toast agree): empty host → panel off.
	// Do not auto-fill host when empty — that made "clear host to hide" impossible.
	if p.AgentToken != nil {
		cfg.Settings.AgentToken = strings.TrimSpace(*p.AgentToken)
	}
	if p.AgentPath != nil {
		cfg.Settings.AgentPath = strings.TrimSpace(*p.AgentPath)
	}
	if p.OpenWith != nil {
		// Full replace of the map when provided (UI sends the whole table).
		clean := map[string]string{}
		for k, v := range p.OpenWith {
			key := dossier.ExtKey(k)
			v = strings.TrimSpace(v)
			if key == "" || v == "" {
				continue
			}
			clean[key] = v
		}
		cfg.Settings.OpenWith = clean
	}
	if err := dossier.SaveConfig(cfg); err != nil {
		return nil, err
	}
	return a.GetAppSettings(), nil
}

// DismissIntro records that the user dismissed a surface intro.
func (a *API) DismissIntro(view string) (map[string]interface{}, error) {
	cfg, err := dossier.LoadConfig()
	if err != nil || cfg == nil {
		cfg = &dossier.AppConfig{}
	}
	view = strings.TrimSpace(view)
	if view == "" {
		return a.GetAppSettings(), nil
	}
	found := false
	for _, v := range cfg.DismissedIntros {
		if v == view {
			found = true
			break
		}
	}
	if !found {
		cfg.DismissedIntros = append(cfg.DismissedIntros, view)
		_ = dossier.SaveConfig(cfg)
	}
	return a.GetAppSettings(), nil
}

// ResetIntros clears all dismissed intros (settings control).
func (a *API) ResetIntros() (map[string]interface{}, error) {
	cfg, err := dossier.LoadConfig()
	if err != nil || cfg == nil {
		cfg = &dossier.AppConfig{}
	}
	cfg.DismissedIntros = []string{}
	if err := dossier.SaveConfig(cfg); err != nil {
		return nil, err
	}
	return a.GetAppSettings(), nil
}
