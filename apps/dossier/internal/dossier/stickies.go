package dossier

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Sticky size presets (post-it inspired dimensions in CSS px).
var StickySizes = map[string][2]float64{
	"mini":     {150, 150},
	"standard": {200, 200},
	"classic":  {220, 220}, // ~3×3"
	"wide":     {280, 180},
	"tall":     {180, 280},
	"large":    {300, 300},
}

// StickyColors is the expanded palette (better than the reference 6-color set).
var StickyColors = []string{
	"yellow", "pink", "blue", "green", "orange", "red",
	"purple", "teal", "mint", "peach", "lavender", "white",
}

// CreateSticky creates a new sticky note with defaults.
func (s *Store) CreateSticky(partial *Sticky) (*Sticky, error) {
	now := nowISO()
	st := &Sticky{
		ID:        uuid.NewString(),
		Text:      "",
		Color:     "yellow",
		Size:      "standard",
		X:         40,
		Y:         40,
		W:         200,
		H:         200,
		Emoji:     "",
		ZIndex:    1,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if partial != nil {
		if partial.Text != "" {
			st.Text = partial.Text
		}
		if partial.Color != "" {
			st.Color = partial.Color
		}
		if partial.Size != "" {
			st.Size = partial.Size
			if dims, ok := StickySizes[partial.Size]; ok {
				st.W, st.H = dims[0], dims[1]
			}
		}
		if partial.X != 0 {
			st.X = partial.X
		}
		if partial.Y != 0 {
			st.Y = partial.Y
		}
		if partial.W > 0 {
			st.W = partial.W
		}
		if partial.H > 0 {
			st.H = partial.H
		}
		if partial.Emoji != "" {
			st.Emoji = partial.Emoji
		}
		if d := strings.TrimSpace(partial.DueAt); d != "" {
			st.DueAt = d
		}
	}
	// Stagger new notes so they don't stack perfectly
	list, _ := s.ListStickies()
	n := len(list)
	st.X = 40 + float64((n%6)*28)
	st.Y = 40 + float64((n%6)*28)
	if z, err := s.MaxStickyZ(); err == nil {
		st.ZIndex = z + 1
	}
	if err := s.UpsertSticky(st); err != nil {
		return nil, err
	}
	return st, nil
}

// UpdateSticky merges fields from the client payload (full sticky object).
// DueAt empty string clears schedule. LinkedKanban: non-nil with both ids sets link;
// non-nil with empty ids clears; nil keeps the previous link (safe for drag/resize
// payloads that omit the field). Prefer LinkStickyToKanban / UnlinkSticky for link ops.
func (s *Store) UpdateSticky(id string, patch *Sticky) (*Sticky, error) {
	list, err := s.ListStickies()
	if err != nil {
		return nil, err
	}
	var cur *Sticky
	for i := range list {
		if list[i].ID == id {
			cur = &list[i]
			break
		}
	}
	if cur == nil {
		return nil, errNotFound("sticky")
	}
	prevDue := strings.TrimSpace(cur.DueAt)
	cur.Text = patch.Text
	if patch.Color != "" {
		cur.Color = patch.Color
	}
	if patch.Size != "" {
		cur.Size = patch.Size
	}
	if patch.W > 0 {
		cur.W = patch.W
	}
	if patch.H > 0 {
		cur.H = patch.H
	}
	cur.X = patch.X
	cur.Y = patch.Y
	cur.Emoji = patch.Emoji
	if patch.ZIndex > 0 {
		cur.ZIndex = patch.ZIndex
	}
	// Always accept DueAt from client (empty = clear).
	cur.DueAt = strings.TrimSpace(patch.DueAt)
	if patch.LinkedKanban != nil {
		b := strings.TrimSpace(patch.LinkedKanban.BoardID)
		c := strings.TrimSpace(patch.LinkedKanban.CardID)
		if b != "" && c != "" {
			cur.LinkedKanban = &StickyLinkKanban{BoardID: b, CardID: c}
		} else {
			cur.LinkedKanban = nil
		}
	}
	cur.UpdatedAt = nowISO()
	if err := s.UpsertSticky(cur); err != nil {
		return nil, err
	}
	// Sync due date to linked kanban card when it changed.
	if cur.LinkedKanban != nil && cur.DueAt != prevDue {
		_ = s.syncDueToKanbanCard(cur.LinkedKanban.BoardID, cur.LinkedKanban.CardID, cur.DueAt)
	}
	return cur, nil
}

// LinkStickyToKanban creates a new card on board/column with sticky text (and due date),
// non-destructively. Original sticky is kept; both get denormalized link ids.
func (s *Store) LinkStickyToKanban(stickyID, boardID, columnID string) (*Sticky, *KanbanBoard, error) {
	stickyID = strings.TrimSpace(stickyID)
	boardID = strings.TrimSpace(boardID)
	columnID = strings.TrimSpace(columnID)
	if stickyID == "" || boardID == "" || columnID == "" {
		return nil, nil, fmt.Errorf("sticky, board, and column are required")
	}
	st, err := s.GetSticky(stickyID)
	if err != nil {
		return nil, nil, err
	}
	b, err := s.GetKanbanBoard(boardID)
	if err != nil {
		return nil, nil, fmt.Errorf("kanban board: %w", err)
	}
	var stt struct {
		Columns []map[string]interface{} `json:"columns"`
		Cards   []map[string]interface{} `json:"cards"`
	}
	if err := json.Unmarshal([]byte(b.StateJSON), &stt); err != nil {
		return nil, nil, fmt.Errorf("kanban state: %w", err)
	}
	colOK := false
	for _, col := range stt.Columns {
		if id, _ := col["id"].(string); id == columnID {
			colOK = true
			break
		}
	}
	if !colOK {
		return nil, nil, fmt.Errorf("column not found on board")
	}
	// Count cards already in column for order
	order := 0
	for _, c := range stt.Cards {
		if cid, _ := c["columnId"].(string); cid == columnID {
			order++
		}
	}
	cardID := uuid.NewString()
	card := map[string]interface{}{
		"id":             cardID,
		"columnId":       columnID,
		"text":           st.Text,
		"color":          st.Color,
		"order":          order,
		"linkedStickyId": stickyID,
	}
	if strings.TrimSpace(st.DueAt) != "" {
		card["dueAt"] = strings.TrimSpace(st.DueAt)
	}
	stt.Cards = append(stt.Cards, card)
	raw, err := json.Marshal(stt)
	if err != nil {
		return nil, nil, err
	}
	b, err = s.SaveKanbanBoard(boardID, b.Title, string(raw))
	if err != nil {
		return nil, nil, err
	}
	st.LinkedKanban = &StickyLinkKanban{BoardID: boardID, CardID: cardID}
	st.UpdatedAt = nowISO()
	if err := s.UpsertSticky(st); err != nil {
		return nil, nil, err
	}
	return st, b, nil
}

// UnlinkSticky clears the sticky's kanban schedule link and the card's linkedStickyId.
// Does not delete either item or change due dates.
func (s *Store) UnlinkSticky(stickyID string) (*Sticky, error) {
	st, err := s.GetSticky(stickyID)
	if err != nil {
		return nil, err
	}
	if st.LinkedKanban != nil {
		_ = s.clearCardStickyLink(st.LinkedKanban.BoardID, st.LinkedKanban.CardID)
		st.LinkedKanban = nil
		st.UpdatedAt = nowISO()
		if err := s.UpsertSticky(st); err != nil {
			return nil, err
		}
	}
	return st, nil
}

func (s *Store) clearCardStickyLink(boardID, cardID string) error {
	b, err := s.GetKanbanBoard(boardID)
	if err != nil {
		return err
	}
	var stt map[string]interface{}
	if err := json.Unmarshal([]byte(b.StateJSON), &stt); err != nil {
		return err
	}
	cards, _ := stt["cards"].([]interface{})
	changed := false
	for _, raw := range cards {
		cm, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		id, _ := cm["id"].(string)
		if id != cardID {
			continue
		}
		if _, had := cm["linkedStickyId"]; had {
			delete(cm, "linkedStickyId")
			changed = true
		}
	}
	if !changed {
		return nil
	}
	raw, err := json.Marshal(stt)
	if err != nil {
		return err
	}
	_, err = s.SaveKanbanBoard(boardID, b.Title, string(raw))
	return err
}

// syncDueToKanbanCard updates dueAt on a card without touching text or sticky link fields.
func (s *Store) syncDueToKanbanCard(boardID, cardID, dueAt string) error {
	b, err := s.GetKanbanBoard(boardID)
	if err != nil {
		return err
	}
	var stt map[string]interface{}
	if err := json.Unmarshal([]byte(b.StateJSON), &stt); err != nil {
		return err
	}
	cards, _ := stt["cards"].([]interface{})
	changed := false
	dueAt = strings.TrimSpace(dueAt)
	for _, raw := range cards {
		cm, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		id, _ := cm["id"].(string)
		if id != cardID {
			continue
		}
		prev, _ := cm["dueAt"].(string)
		if strings.TrimSpace(prev) == dueAt {
			return nil
		}
		if dueAt == "" {
			delete(cm, "dueAt")
		} else {
			cm["dueAt"] = dueAt
		}
		changed = true
		break
	}
	if !changed {
		return nil
	}
	raw, err := json.Marshal(stt)
	if err != nil {
		return err
	}
	// Use low-level upsert to avoid recursive sticky sync from SaveKanbanBoard hook
	_, err = s.saveKanbanStateOnly(boardID, b.Title, string(raw))
	return err
}

// SyncDueFromKanbanCard updates a sticky's due_at from a card (used after kanban save).
func (s *Store) SyncDueFromKanbanCard(stickyID, dueAt string) error {
	st, err := s.GetSticky(stickyID)
	if err != nil {
		return err
	}
	dueAt = strings.TrimSpace(dueAt)
	if strings.TrimSpace(st.DueAt) == dueAt {
		return nil
	}
	st.DueAt = dueAt
	st.UpdatedAt = nowISO()
	return s.UpsertSticky(st)
}

// dateOnly normalizes an ISO or date-only string to YYYY-MM-DD; empty if unparseable/empty.
func dateOnly(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if len(s) >= 10 && s[4] == '-' && s[7] == '-' {
		return s[:10]
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UTC().Format("2006-01-02")
	}
	return ""
}

// ListCalendarItems aggregates sticky due dates, kanban card due dates, and decision dates.
// Items with no date are omitted (never invented).
func (s *Store) ListCalendarItems() ([]CalendarItem, error) {
	var out []CalendarItem
	stickies, err := s.ListStickies()
	if err != nil {
		return nil, err
	}
	for _, st := range stickies {
		d := dateOnly(st.DueAt)
		if d == "" {
			continue
		}
		title := strings.TrimSpace(st.Text)
		if title == "" {
			title = "Sticky"
		} else if len(title) > 80 {
			title = title[:80] + "…"
		}
		out = append(out, CalendarItem{
			ID: st.ID, Kind: "sticky", Title: title, Date: d, DueAt: st.DueAt, Color: st.Color,
		})
	}
	boards, err := s.ListKanbanBoards()
	if err != nil {
		return nil, err
	}
	for _, b := range boards {
		var stt struct {
			Cards []struct {
				ID             string `json:"id"`
				Text           string `json:"text"`
				DueAt          string `json:"dueAt"`
				Color          string `json:"color"`
				LinkedStickyID string `json:"linkedStickyId"`
			} `json:"cards"`
		}
		_ = json.Unmarshal([]byte(b.StateJSON), &stt)
		for _, c := range stt.Cards {
			d := dateOnly(c.DueAt)
			if d == "" {
				continue
			}
			title := strings.TrimSpace(c.Text)
			if title == "" {
				title = "Card"
			} else if len(title) > 80 {
				title = title[:80] + "…"
			}
			if b.Title != "" {
				title = b.Title + ": " + title
			}
			out = append(out, CalendarItem{
				ID: c.ID, Kind: "kanban", Title: title, Date: d, DueAt: c.DueAt,
				BoardID: b.ID, CardID: c.ID, Color: c.Color,
			})
		}
	}
	decs, err := s.ListDecisions()
	if err != nil {
		return nil, err
	}
	for _, d := range decs {
		day := dateOnly(d.DecidedAt)
		if day == "" {
			continue
		}
		title := strings.TrimSpace(d.Title)
		if title == "" {
			title = "Decision"
		}
		out = append(out, CalendarItem{
			ID: d.ID, Kind: "decision", Title: title, Date: day, DueAt: d.DecidedAt,
		})
	}
	// Sort by date then kind then title
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].Date < out[i].Date ||
				(out[j].Date == out[i].Date && out[j].Kind < out[i].Kind) ||
				(out[j].Date == out[i].Date && out[j].Kind == out[i].Kind && out[j].Title < out[i].Title) {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	if out == nil {
		out = []CalendarItem{}
	}
	return out, nil
}

type notFoundError struct{ what string }

func (e notFoundError) Error() string { return e.what + " not found" }

func errNotFound(what string) error { return notFoundError{what: what} }
