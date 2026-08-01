package dossier

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var unsafeName = regexp.MustCompile(`[^a-zA-Z0-9._\-\s]+`)

// CreateNote creates a new markdown note file and DB row.
func (s *Store) CreateNote(title, body string) (*Note, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		title = "Untitled note"
	}
	id := uuid.NewString()
	filename := sanitizeFilename(title) + ".md"
	destDir := filepath.Join(s.Root, NotesDir)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return nil, err
	}
	filename = uniqueName(destDir, filename)
	rel := filepath.ToSlash(filepath.Join(NotesDir, filename))
	abs := s.AbsPath(rel)

	content := body
	if content == "" {
		content = "# " + title + "\n\n"
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		return nil, err
	}
	now := nowISO()
	n := &Note{
		ID:        id,
		Title:     title,
		RelPath:   rel,
		Body:      content,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.UpsertNote(n, content); err != nil {
		return nil, err
	}
	return n, nil
}

// SaveNote writes note body to disk and updates index.
func (s *Store) SaveNote(id, title, body string) (*Note, error) {
	n, err := s.GetNote(id)
	if err != nil {
		return nil, err
	}
	title = strings.TrimSpace(title)
	if title == "" {
		title = n.Title
	}
	// Derive title from first heading if still empty
	if title == "" || title == "Untitled note" {
		if t := firstHeading(body); t != "" {
			title = t
		}
	}
	abs := s.AbsPath(n.RelPath)
	if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
		return nil, err
	}
	n.Title = title
	n.Body = body
	n.UpdatedAt = nowISO()
	if err := s.UpsertNote(n, body); err != nil {
		return nil, err
	}
	return n, nil
}

// RescanNotes indexes notes folder.
func (s *Store) RescanNotes() (int, error) {
	dir := filepath.Join(s.Root, NotesDir)
	count := 0
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".md" && ext != ".txt" && ext != ".markdown" {
			continue
		}
		rel := filepath.ToSlash(filepath.Join(NotesDir, name))
		// Check if already tracked
		s.mu.RLock()
		var existingID string
		err := s.db.QueryRow(`SELECT id FROM notes WHERE rel_path=?`, rel).Scan(&existingID)
		s.mu.RUnlock()
		if err == nil {
			// Update index from file
			abs := s.AbsPath(rel)
			b, rerr := os.ReadFile(abs)
			if rerr != nil {
				continue
			}
			body := string(b)
			title := firstHeading(body)
			if title == "" {
				title = strings.TrimSuffix(name, filepath.Ext(name))
			}
			n := &Note{ID: existingID, Title: title, RelPath: rel, Body: body, CreatedAt: nowISO(), UpdatedAt: nowISO()}
			// preserve created_at
			if full, gerr := s.GetNote(existingID); gerr == nil {
				n.CreatedAt = full.CreatedAt
			}
			_ = s.UpsertNote(n, body)
			count++
			continue
		}
		// New file on disk → register
		abs := s.AbsPath(rel)
		b, rerr := os.ReadFile(abs)
		if rerr != nil {
			continue
		}
		body := string(b)
		title := firstHeading(body)
		if title == "" {
			title = strings.TrimSuffix(name, filepath.Ext(name))
		}
		id := uuid.NewString()
		now := nowISO()
		n := &Note{ID: id, Title: title, RelPath: rel, Body: body, CreatedAt: now, UpdatedAt: now}
		if err := s.UpsertNote(n, body); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func sanitizeFilename(s string) string {
	s = unsafeName.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "-")
	if s == "" {
		s = "note"
	}
	if len(s) > 60 {
		s = s[:60]
	}
	return s
}

func firstHeading(body string) string {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			t := strings.TrimSpace(strings.TrimLeft(line, "#"))
			if t != "" {
				return t
			}
		}
	}
	// first non-empty line
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			if len(line) > 80 {
				return line[:80]
			}
			return line
		}
	}
	return ""
}

// Ensure note title helper for UI.
func NoteTitleFromBody(body, fallback string) string {
	if t := firstHeading(body); t != "" {
		return t
	}
	if fallback != "" {
		return fallback
	}
	return "Untitled note"
}

// FormatBytes humanizes size.
func FormatBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}
