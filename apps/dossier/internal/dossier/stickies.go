package dossier

import (
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

// UpdateSticky merges fields from the client payload.
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
	// Client always sends the full sticky object on update.
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
	cur.UpdatedAt = nowISO()
	if err := s.UpsertSticky(cur); err != nil {
		return nil, err
	}
	return cur, nil
}

type notFoundError struct{ what string }

func (e notFoundError) Error() string { return e.what + " not found" }

func errNotFound(what string) error { return notFoundError{what: what} }
