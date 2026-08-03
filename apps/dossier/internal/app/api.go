package app

import (
	"encoding/base64"
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
	Name         string   `json:"name"`         // folder base name (always)
	DisplayName  string   `json:"displayName"`  // optional rename, or same as Name
	Version      string   `json:"version"`
	HasLastPath  bool     `json:"hasLastPath"`
	LastPath     string   `json:"lastPath"`
	LastName     string   `json:"lastName"` // display name for last path
	RecentPaths  []string `json:"recentPaths"`
	RecentNames  []string `json:"recentNames"` // display names (folder base if unset)
	// ActiveCollectionID is set when the open dossier path is a member of a collection.
	ActiveCollectionID   string `json:"activeCollectionId,omitempty"`
	ActiveCollectionName string `json:"activeCollectionName,omitempty"`
}

func (a *API) GetStatus() Status {
	cfg, _ := dossier.LoadConfig()
	if cfg == nil {
		cfg = &dossier.AppConfig{}
	}
	recents := dossier.ListRecentDossiers()
	names := make([]string, len(recents))
	for i, p := range recents {
		names[i] = cfg.DisplayNameFor(p)
	}
	st := Status{
		Version:     Version,
		HasLastPath: cfg.LastDossierPath != "" && dossier.IsDossierFolder(cfg.LastDossierPath),
		LastPath:    cfg.LastDossierPath,
		LastName:    cfg.DisplayNameFor(cfg.LastDossierPath),
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
		st.DisplayName = cfg.DisplayNameFor(a.store.Root)
		if col := dossier.CollectionContaining(a.store.Root); col != nil {
			st.ActiveCollectionID = col.ID
			st.ActiveCollectionName = col.Name
		}
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
	cfg, _ := dossier.LoadConfig()
	if cfg == nil {
		cfg = &dossier.AppConfig{}
	}
	paths := dossier.ListRecentDossiers()
	out := make([]map[string]string, 0, len(paths))
	for _, p := range paths {
		out = append(out, map[string]string{
			"path":        p,
			"name":        cfg.DisplayNameFor(p),
			"folderName":  filepath.Base(p),
			"displayName": cfg.DisplayNameFor(p),
		})
	}
	return out
}

// SetWorkspaceDisplayName sets or clears an optional display name for a dossier path.
// Empty name clears the override. Never renames the folder on disk.
func (a *API) SetWorkspaceDisplayName(path, name string) (Status, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return a.GetStatus(), fmt.Errorf("path required")
	}
	cfg, err := dossier.LoadConfig()
	if err != nil || cfg == nil {
		cfg = &dossier.AppConfig{}
	}
	cfg.SetWorkspaceDisplayName(path, name)
	if err := dossier.SaveConfig(cfg); err != nil {
		return a.GetStatus(), err
	}
	return a.GetStatus(), nil
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
	return s.ImportAbsolutePaths(paths)
}

// ImportPaths imports files by absolute path (bookmark multi-select).
func (a *API) ImportPaths(paths []string) ([]dossier.Document, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	return s.ImportAbsolutePaths(paths)
}

// --- Collections (app-wide) ---

func (a *API) ListCollections() []map[string]interface{} {
	cfg, _ := dossier.LoadConfig()
	if cfg == nil {
		cfg = &dossier.AppConfig{}
	}
	cols := dossier.ListCollections()
	out := make([]map[string]interface{}, 0, len(cols))
	openRoot := ""
	a.mu.Lock()
	if a.store != nil {
		openRoot = a.store.Root
	}
	a.mu.Unlock()
	for _, c := range cols {
		members := dossier.CollectionMemberViews(c)
		// Mark which member is currently open
		for i := range members {
			if openRoot != "" && strings.EqualFold(filepath.Clean(members[i]["path"]), filepath.Clean(openRoot)) {
				members[i]["open"] = "true"
			} else {
				members[i]["open"] = "false"
			}
		}
		out = append(out, map[string]interface{}{
			"id":      c.ID,
			"name":    c.Name,
			"paths":   c.Paths,
			"members": members,
		})
	}
	return out
}

func (a *API) CreateCollection(name string) (map[string]interface{}, error) {
	col, err := dossier.CreateCollection(name)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"id": col.ID, "name": col.Name, "paths": col.Paths,
		"members": dossier.CollectionMemberViews(*col),
	}, nil
}

func (a *API) DeleteCollection(id string) error {
	return dossier.DeleteCollection(id)
}

func (a *API) RenameCollection(id, name string) (map[string]interface{}, error) {
	col, err := dossier.RenameCollection(id, name)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"id": col.ID, "name": col.Name, "paths": col.Paths,
		"members": dossier.CollectionMemberViews(*col),
	}, nil
}

// AddDossierToCollection picks a folder (or uses path) and adds it to the collection.
// If path is empty, opens a folder picker.
func (a *API) AddDossierToCollection(collectionID, path string) (map[string]interface{}, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		p, err := dialog.PickFolder("Add dossier folder to collection")
		if err != nil {
			return nil, err
		}
		if p == "" {
			return nil, fmt.Errorf("cancelled")
		}
		path = p
	}
	// Initialize as dossier if needed
	if !dossier.IsDossierFolder(path) {
		st, err := dossier.OpenOrCreate(path)
		if err != nil {
			return nil, err
		}
		_ = st.Close()
	}
	col, err := dossier.AddPathToCollection(collectionID, path)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"id": col.ID, "name": col.Name, "paths": col.Paths,
		"members": dossier.CollectionMemberViews(*col),
	}, nil
}

func (a *API) RemoveDossierFromCollection(collectionID, path string) (map[string]interface{}, error) {
	col, err := dossier.RemovePathFromCollection(collectionID, path)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"id": col.ID, "name": col.Name, "paths": col.Paths,
		"members": dossier.CollectionMemberViews(*col),
	}, nil
}

// --- Import bookmarks (per-dossier portable file) ---

func (a *API) ListBookmarks() ([]map[string]interface{}, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	return s.ListBookmarks()
}

func (a *API) AddBookmarkFolder() (map[string]interface{}, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	path, err := dialog.PickFolder("Bookmark external folder for import")
	if err != nil {
		return nil, err
	}
	if path == "" {
		return nil, fmt.Errorf("cancelled")
	}
	fb, err := s.AddBookmarkFolder(path, "")
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"id": fb.ID, "path": fb.Path, "label": fb.Label, "broken": false,
	}, nil
}

func (a *API) RemoveBookmark(id string) error {
	s, err := a.requireStore()
	if err != nil {
		return err
	}
	return s.RemoveBookmark(id)
}

func (a *API) RepickBookmark(id string) (map[string]interface{}, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	path, err := dialog.PickFolder("Re-pick bookmarked folder")
	if err != nil {
		return nil, err
	}
	if path == "" {
		return nil, fmt.Errorf("cancelled")
	}
	fb, err := s.UpdateBookmarkPath(id, path)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"id": fb.ID, "path": fb.Path, "label": fb.Label, "broken": false,
	}, nil
}

func (a *API) ListBookmarkFiles(bookmarkID string) ([]dossier.ExternalFileEntry, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	return s.ListBookmarkFiles(bookmarkID)
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

// OpenURL opens http(s)/mailto links in the system browser (never navigates the WebView).
func (a *API) OpenURL(raw string) error {
	return dialog.OpenURL(raw)
}

// ResolveNoteAsset resolves a relative or absolute image path for a note preview.
// Returns a data: URL on success, or empty string if unreadable (no network fetch).
// Paths must resolve under the dossier root (or absolute under root after clean).
func (a *API) ResolveNoteAsset(noteID, href string) (string, error) {
	s, err := a.requireStore()
	if err != nil {
		return "", err
	}
	href = strings.TrimSpace(href)
	if href == "" {
		return "", nil
	}
	// Reject network and exotic schemes
	lower := strings.ToLower(href)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") ||
		strings.HasPrefix(lower, "data:") || strings.HasPrefix(lower, "//") {
		return "", nil
	}
	var abs string
	if filepath.IsAbs(href) {
		abs = filepath.Clean(href)
	} else {
		// Prefer note directory, then dossier root
		n, nerr := s.GetNote(noteID)
		baseDir := s.Root
		if nerr == nil && n != nil && n.RelPath != "" {
			baseDir = filepath.Dir(s.AbsPath(n.RelPath))
		}
		abs = filepath.Clean(filepath.Join(baseDir, filepath.FromSlash(href)))
		// If missing under note dir, try dossier root
		if _, err := os.Stat(abs); err != nil {
			abs = filepath.Clean(filepath.Join(s.Root, filepath.FromSlash(href)))
		}
	}
	// Must stay inside dossier root
	rootClean := filepath.Clean(s.Root) + string(filepath.Separator)
	absClean := filepath.Clean(abs)
	if absClean != filepath.Clean(s.Root) && !strings.HasPrefix(strings.ToLower(absClean+string(filepath.Separator)), strings.ToLower(rootClean)) {
		// Windows case-insensitive: also try EqualFold prefix
		if !strings.HasPrefix(strings.ToLower(absClean), strings.ToLower(filepath.Clean(s.Root))) {
			return "", nil
		}
	}
	raw, err := os.ReadFile(abs)
	if err != nil {
		return "", nil
	}
	// Cap image size for data URLs (~8 MiB)
	if len(raw) > 8<<20 {
		return "", nil
	}
	mime := "application/octet-stream"
	switch strings.ToLower(filepath.Ext(abs)) {
	case ".png":
		mime = "image/png"
	case ".jpg", ".jpeg":
		mime = "image/jpeg"
	case ".gif":
		mime = "image/gif"
	case ".webp":
		mime = "image/webp"
	case ".svg":
		mime = "image/svg+xml"
	case ".bmp":
		mime = "image/bmp"
	}
	enc := base64.StdEncoding.EncodeToString(raw)
	return "data:" + mime + ";base64," + enc, nil
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

// LinkStickyToKanban promotes sticky text into a new card (non-destructive).
func (a *API) LinkStickyToKanban(stickyID, boardID, columnID string) (map[string]interface{}, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	st, board, err := s.LinkStickyToKanban(stickyID, boardID, columnID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"sticky": st, "board": board}, nil
}

// UnlinkSticky clears the sticky↔kanban schedule link on both ends.
func (a *API) UnlinkSticky(stickyID string) (*dossier.Sticky, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	return s.UnlinkSticky(stickyID)
}

// ListCalendarItems returns agenda items (sticky/kanban/decision dates only).
func (a *API) ListCalendarItems() ([]dossier.CalendarItem, error) {
	s, err := a.requireStore()
	if err != nil {
		return nil, err
	}
	return s.ListCalendarItems()
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
		"noteAutosaveMs":   cfg.Settings.NoteAutosaveMsOrDefault(),
		"confirmDeletes":   confirm,
		"showIntros":       showIntros,
		"theme":            cfg.Settings.Theme,
		"dismissedIntros":  dismissed,
		"agentHost":        cfg.Settings.AgentHost,
		"agentPort":        cfg.Settings.AgentPort,
		"agentToken":       cfg.Settings.AgentToken,
		"agentPath":        path,
		"agentConfigured":  agentOn,
		"openWith":         openWith,
		"autoOpenLast":     cfg.Settings.AutoOpenLastOrDefault(),
		"notesEditorMode":  cfg.Settings.NotesEditorModeOrDefault(),
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
	NotesEditorMode *string           `json:"notesEditorMode"`
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
	if p.NotesEditorMode != nil {
		m := strings.ToLower(strings.TrimSpace(*p.NotesEditorMode))
		switch m {
		case "edit", "split", "preview":
			cfg.Settings.NotesEditorMode = m
		}
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
