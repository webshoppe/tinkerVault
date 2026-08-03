package dossier

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
)

// DefaultKanbanState: three columns, empty cards — improved over reference with
// WIP limits optional field and card colors.
func DefaultKanbanState() map[string]interface{} {
	todo := uuid.NewString()
	doing := uuid.NewString()
	done := uuid.NewString()
	return map[string]interface{}{
		"columns": []map[string]interface{}{
			{"id": todo, "title": "To Do", "wipLimit": 0},
			{"id": doing, "title": "In Progress", "wipLimit": 5},
			{"id": done, "title": "Done", "wipLimit": 0},
		},
		"cards": []map[string]interface{}{},
	}
}

// ListKanbanBoards returns all kanban boards newest first.
func (s *Store) ListKanbanBoards() ([]KanbanBoard, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows, err := s.db.Query(`SELECT id, title, state_json, created_at, updated_at FROM kanban_boards ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []KanbanBoard
	for rows.Next() {
		var b KanbanBoard
		if err := rows.Scan(&b.ID, &b.Title, &b.StateJSON, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	if out == nil {
		out = []KanbanBoard{}
	}
	return out, rows.Err()
}

// CreateKanbanBoard creates a board with default columns.
func (s *Store) CreateKanbanBoard(title string) (*KanbanBoard, error) {
	if title == "" {
		title = "Kanban"
	}
	b, _ := json.Marshal(DefaultKanbanState())
	now := nowISO()
	board := &KanbanBoard{
		ID:        uuid.NewString(),
		Title:     title,
		StateJSON: string(b),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.upsertKanban(board); err != nil {
		return nil, err
	}
	return board, nil
}

// GetKanbanBoard loads one board.
func (s *Store) GetKanbanBoard(id string) (*KanbanBoard, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var b KanbanBoard
	err := s.db.QueryRow(`SELECT id, title, state_json, created_at, updated_at FROM kanban_boards WHERE id=?`, id).
		Scan(&b.ID, &b.Title, &b.StateJSON, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// SaveKanbanBoard updates title + full state JSON.
// After save, syncs dueAt from cards that have linkedStickyId onto those stickies.
func (s *Store) SaveKanbanBoard(id, title, stateJSON string) (*KanbanBoard, error) {
	b, err := s.saveKanbanStateOnly(id, title, stateJSON)
	if err != nil {
		return nil, err
	}
	// Sync schedule links: card dueAt → sticky dueAt when linked.
	var stt struct {
		Cards []struct {
			ID             string `json:"id"`
			DueAt          string `json:"dueAt"`
			LinkedStickyID string `json:"linkedStickyId"`
		} `json:"cards"`
	}
	_ = json.Unmarshal([]byte(b.StateJSON), &stt)
	for _, c := range stt.Cards {
		sid := strings.TrimSpace(c.LinkedStickyID)
		if sid == "" {
			continue
		}
		// Card dueAt (including empty clear) is source of truth for this write path.
		if err := s.SyncDueFromKanbanCard(sid, c.DueAt); err != nil {
			// Non-fatal for the board save; sticky may have been deleted.
			fmt.Fprintf(os.Stderr, "sync card→sticky due %s: %v\n", sid, err)
		}
	}
	return b, nil
}

// saveKanbanStateOnly persists board JSON without sticky due-date side effects
// (used when sticky is the source of truth for a due-date write).
func (s *Store) saveKanbanStateOnly(id, title, stateJSON string) (*KanbanBoard, error) {
	b, err := s.GetKanbanBoard(id)
	if err != nil {
		return nil, err
	}
	if title != "" {
		b.Title = title
	}
	if stateJSON != "" {
		var tmp interface{}
		if err := json.Unmarshal([]byte(stateJSON), &tmp); err != nil {
			return nil, fmt.Errorf("invalid kanban state: %w", err)
		}
		b.StateJSON = stateJSON
	}
	b.UpdatedAt = nowISO()
	if err := s.upsertKanban(b); err != nil {
		return nil, err
	}
	_ = s.indexKanbanFTS(b)
	return b, nil
}

// DeleteKanbanBoard removes a board.
func (s *Store) DeleteKanbanBoard(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.db.Exec(`DELETE FROM kanban_boards WHERE id=?`, id); err != nil {
		return err
	}
	return s.removeFTS(id)
}

func (s *Store) upsertKanban(b *KanbanBoard) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`
INSERT INTO kanban_boards(id, title, state_json, created_at, updated_at)
VALUES(?,?,?,?,?)
ON CONFLICT(id) DO UPDATE SET title=excluded.title, state_json=excluded.state_json, updated_at=excluded.updated_at
`, b.ID, b.Title, b.StateJSON, b.CreatedAt, b.UpdatedAt)
	return err
}

func (s *Store) indexKanbanFTS(b *KanbanBoard) error {
	var st struct {
		Cards []struct {
			Text string `json:"text"`
		} `json:"cards"`
	}
	_ = json.Unmarshal([]byte(b.StateJSON), &st)
	var body string
	for _, c := range st.Cards {
		if c.Text != "" {
			body += c.Text + "\n"
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.reindexFTS(b.ID, "kanban", b.Title, body)
}
