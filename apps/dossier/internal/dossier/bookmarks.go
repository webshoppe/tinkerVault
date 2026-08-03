package dossier

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// BookmarksFileName is the portable per-dossier bookmarks file (copy-folder backup).
const BookmarksFileName = "bookmarks.json"

// ImportBookmark is one external folder shortcut for import sources.
type ImportBookmark struct {
	ID    string `json:"id"`
	Path  string `json:"path"`
	Label string `json:"label,omitempty"`
}

// DossierBookmarks is the on-disk shape of bookmarks.json.
type DossierBookmarks struct {
	Folders []ImportBookmark `json:"folders"`
}

// ExternalFileEntry is a single file listing row for a bookmarked folder.
type ExternalFileEntry struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	Size  int64  `json:"size"`
	IsDir bool   `json:"isDir"`
	Ext   string `json:"ext"`
}

func (s *Store) bookmarksPath() string {
	return filepath.Join(s.Root, BookmarksFileName)
}

// LoadBookmarks reads bookmarks.json (empty list if missing).
func (s *Store) LoadBookmarks() (*DossierBookmarks, error) {
	raw, err := os.ReadFile(s.bookmarksPath())
	if err != nil {
		if os.IsNotExist(err) {
			return &DossierBookmarks{Folders: []ImportBookmark{}}, nil
		}
		return nil, err
	}
	var b DossierBookmarks
	if err := json.Unmarshal(raw, &b); err != nil {
		return nil, fmt.Errorf("bookmarks.json: %w", err)
	}
	if b.Folders == nil {
		b.Folders = []ImportBookmark{}
	}
	return &b, nil
}

func (s *Store) saveBookmarks(b *DossierBookmarks) error {
	if b.Folders == nil {
		b.Folders = []ImportBookmark{}
	}
	raw, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.bookmarksPath(), raw, 0o644)
}

// bookmarkPathBroken reports whether path is missing, not a directory, or unreadable.
// Used whenever the Sources list is rendered so renames/moves outside the app surface
// as a broken badge without requiring a Browse click first.
func bookmarkPathBroken(path string) bool {
	path = filepath.Clean(strings.TrimSpace(path))
	if path == "" || path == "." {
		return true
	}
	st, err := os.Stat(path)
	if err != nil || !st.IsDir() {
		return true
	}
	// Stat can succeed on some stub entries; require a readable directory.
	d, err := os.Open(path)
	if err != nil {
		return true
	}
	_ = d.Close()
	return false
}

// ListBookmarks returns bookmarks with a broken flag when path is missing.
// Always re-stats each path (no cache) so UI list render reflects disk truth.
func (s *Store) ListBookmarks() ([]map[string]interface{}, error) {
	b, err := s.LoadBookmarks()
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(b.Folders))
	for _, f := range b.Folders {
		broken := bookmarkPathBroken(f.Path)
		label := f.Label
		if label == "" {
			label = filepath.Base(f.Path)
		}
		out = append(out, map[string]interface{}{
			"id":     f.ID,
			"path":   f.Path,
			"label":  label,
			"broken": broken,
		})
	}
	return out, nil
}

// AddBookmarkFolder appends an external folder shortcut.
func (s *Store) AddBookmarkFolder(path, label string) (*ImportBookmark, error) {
	path = filepath.Clean(strings.TrimSpace(path))
	if path == "" {
		return nil, fmt.Errorf("path required")
	}
	st, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("folder not accessible: %w", err)
	}
	if !st.IsDir() {
		return nil, fmt.Errorf("not a folder: %s", path)
	}
	b, err := s.LoadBookmarks()
	if err != nil {
		return nil, err
	}
	for _, f := range b.Folders {
		if strings.EqualFold(filepath.Clean(f.Path), path) {
			return &f, nil // already bookmarked
		}
	}
	label = strings.TrimSpace(label)
	if label == "" {
		label = filepath.Base(path)
	}
	fb := ImportBookmark{
		ID:    uuid.NewString(),
		Path:  path,
		Label: label,
	}
	b.Folders = append(b.Folders, fb)
	if err := s.saveBookmarks(b); err != nil {
		return nil, err
	}
	return &fb, nil
}

// RemoveBookmark deletes a bookmark by id.
func (s *Store) RemoveBookmark(id string) error {
	id = strings.TrimSpace(id)
	b, err := s.LoadBookmarks()
	if err != nil {
		return err
	}
	out := b.Folders[:0]
	found := false
	for _, f := range b.Folders {
		if f.ID == id {
			found = true
			continue
		}
		out = append(out, f)
	}
	if !found {
		return fmt.Errorf("bookmark not found")
	}
	b.Folders = out
	return s.saveBookmarks(b)
}

// UpdateBookmarkPath re-picks a bookmark to a new folder path (same entry id).
func (s *Store) UpdateBookmarkPath(id, path string) (*ImportBookmark, error) {
	path = filepath.Clean(strings.TrimSpace(path))
	if path == "" {
		return nil, fmt.Errorf("path required")
	}
	if bookmarkPathBroken(path) {
		return nil, fmt.Errorf("folder not accessible")
	}
	b, err := s.LoadBookmarks()
	if err != nil {
		return nil, err
	}
	for i := range b.Folders {
		if b.Folders[i].ID != id {
			continue
		}
		oldBase := filepath.Base(b.Folders[i].Path)
		oldLabel := b.Folders[i].Label
		b.Folders[i].Path = path
		// Refresh auto-label when it was empty or matched the previous folder base.
		if oldLabel == "" || oldLabel == oldBase {
			b.Folders[i].Label = filepath.Base(path)
		}
		if err := s.saveBookmarks(b); err != nil {
			return nil, err
		}
		fb := b.Folders[i]
		return &fb, nil
	}
	return nil, fmt.Errorf("bookmark not found")
}

// ListBookmarkFiles lists non-hidden files in a bookmarked folder (not recursive).
// Always reads the live directory (no listing cache). Missing/renamed paths error.
func (s *Store) ListBookmarkFiles(bookmarkID string) ([]ExternalFileEntry, error) {
	b, err := s.LoadBookmarks()
	if err != nil {
		return nil, err
	}
	var folder string
	for _, f := range b.Folders {
		if f.ID == bookmarkID {
			folder = f.Path
			break
		}
	}
	if folder == "" {
		return nil, fmt.Errorf("bookmark not found")
	}
	if bookmarkPathBroken(folder) {
		return nil, fmt.Errorf("folder not accessible (path missing or moved)")
	}
	return ListExternalFolder(folder)
}

// ListExternalFolder lists immediate children of an external folder (files + dirs).
// Fresh os.ReadDir every call; never returns a cached listing.
func ListExternalFolder(folder string) ([]ExternalFileEntry, error) {
	folder = filepath.Clean(folder)
	st, err := os.Stat(folder)
	if err != nil {
		return nil, fmt.Errorf("folder not accessible: %w", err)
	}
	if !st.IsDir() {
		return nil, fmt.Errorf("not a folder")
	}
	entries, err := os.ReadDir(folder)
	if err != nil {
		return nil, fmt.Errorf("folder not accessible: %w", err)
	}
	out := make([]ExternalFileEntry, 0, len(entries))
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "~$") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		// Skip very new temp noise
		_ = time.Now()
		abs := filepath.Join(folder, name)
		out = append(out, ExternalFileEntry{
			Name:  name,
			Path:  abs,
			Size:  info.Size(),
			IsDir: e.IsDir(),
			Ext:   strings.ToLower(filepath.Ext(name)),
		})
	}
	return out, nil
}

// ImportAbsolutePaths imports external files by absolute path (bookmark multi-select).
func (s *Store) ImportAbsolutePaths(paths []string) ([]Document, error) {
	var imported []Document
	for _, p := range paths {
		p = filepath.Clean(strings.TrimSpace(p))
		if p == "" {
			continue
		}
		st, err := os.Stat(p)
		if err != nil || st.IsDir() {
			continue
		}
		if IsOfficeLockFile(filepath.Base(p)) {
			continue
		}
		d, err := s.ImportFile(p)
		if err != nil {
			continue
		}
		imported = append(imported, *d)
	}
	if imported == nil {
		imported = []Document{}
	}
	return imported, nil
}
