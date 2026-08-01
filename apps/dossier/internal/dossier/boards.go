package dossier

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// DefaultPaintState is the initial Paint canvas state (UI fills imageDataUrl on first draw).
func DefaultPaintState() map[string]interface{} {
	return map[string]interface{}{
		"tool":         "pen",
		"color":        "#1a1a1a",
		"brush":        4,
		"imageDataUrl": "",
		"bg":           "#ffffff",
	}
}

// DefaultAnnotateState is the initial Annotate board (image markup).
// fontSize is a first-class control — the reference whiteboard hard-coded 18px.
func DefaultAnnotateState() map[string]interface{} {
	return map[string]interface{}{
		"tool":         "highlight",
		"strokeColor":  "#e03131",
		"fontSize":     24,
		"imageDataUrl": "",
		"textLayers":   []interface{}{},
	}
}

// ListCanvases returns paint/annotate boards, newest first.
func (s *Store) ListCanvases(kind string) ([]Canvas, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var (
		rows interface {
			Next() bool
			Scan(dest ...interface{}) error
			Close() error
			Err() error
		}
		err error
	)
	if kind == "" {
		rows, err = s.db.Query(`SELECT id, kind, title, state_json, created_at, updated_at FROM canvases ORDER BY updated_at DESC`)
	} else {
		rows, err = s.db.Query(`SELECT id, kind, title, state_json, created_at, updated_at FROM canvases WHERE kind=? ORDER BY updated_at DESC`, kind)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Canvas
	for rows.Next() {
		var c Canvas
		if err := rows.Scan(&c.ID, &c.Kind, &c.Title, &c.StateJSON, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if out == nil {
		out = []Canvas{}
	}
	return out, rows.Err()
}

// CreateCanvas creates a paint or annotate board.
func (s *Store) CreateCanvas(kind, title string) (*Canvas, error) {
	kind = normalizeCanvasKind(kind)
	if title == "" {
		if kind == "annotate" {
			title = "Annotate"
		} else {
			title = "Paint"
		}
	}
	var state map[string]interface{}
	if kind == "annotate" {
		state = DefaultAnnotateState()
	} else {
		state = DefaultPaintState()
	}
	b, _ := json.Marshal(state)
	now := nowISO()
	c := &Canvas{
		ID:        uuid.NewString(),
		Kind:      kind,
		Title:     title,
		StateJSON: string(b),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.upsertCanvas(c); err != nil {
		return nil, err
	}
	return c, nil
}

// GetCanvas returns one board.
func (s *Store) GetCanvas(id string) (*Canvas, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var c Canvas
	err := s.db.QueryRow(`SELECT id, kind, title, state_json, created_at, updated_at FROM canvases WHERE id=?`, id).
		Scan(&c.ID, &c.Kind, &c.Title, &c.StateJSON, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// SaveCanvas updates title + state JSON.
func (s *Store) SaveCanvas(id, title, stateJSON string) (*Canvas, error) {
	c, err := s.GetCanvas(id)
	if err != nil {
		return nil, err
	}
	if title != "" {
		c.Title = title
	}
	if stateJSON != "" {
		var tmp interface{}
		if err := json.Unmarshal([]byte(stateJSON), &tmp); err != nil {
			return nil, fmt.Errorf("invalid state json: %w", err)
		}
		c.StateJSON = stateJSON
	}
	c.UpdatedAt = nowISO()
	if err := s.upsertCanvas(c); err != nil {
		return nil, err
	}
	return c, nil
}

// DeleteCanvas removes a board.
func (s *Store) DeleteCanvas(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM canvases WHERE id=?`, id)
	return err
}

func (s *Store) upsertCanvas(c *Canvas) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`
INSERT INTO canvases(id, kind, title, state_json, created_at, updated_at)
VALUES(?,?,?,?,?,?)
ON CONFLICT(id) DO UPDATE SET
  kind=excluded.kind, title=excluded.title, state_json=excluded.state_json, updated_at=excluded.updated_at
`, c.ID, c.Kind, c.Title, c.StateJSON, c.CreatedAt, c.UpdatedAt)
	return err
}

func normalizeCanvasKind(k string) string {
	switch k {
	case "annotate", "snip":
		return "annotate"
	default:
		return "paint"
	}
}
