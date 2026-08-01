package dossier

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ListDecisions returns decisions in chronological spine order (decided_at ASC).
func (s *Store) ListDecisions() ([]Decision, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows, err := s.db.Query(`SELECT id, title, body, decided_at, COALESCE(document_id,''), COALESCE(doc_version_rel,''), COALESCE(doc_version_note,''), created_at, updated_at
		FROM decisions ORDER BY decided_at ASC, created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Decision
	for rows.Next() {
		var d Decision
		if err := rows.Scan(&d.ID, &d.Title, &d.Body, &d.DecidedAt, &d.DocumentID, &d.DocVersionRel, &d.DocVersionNote, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	if out == nil {
		out = []Decision{}
	}
	return out, rows.Err()
}

// GetDecision loads one decision.
func (s *Store) GetDecision(id string) (*Decision, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var d Decision
	err := s.db.QueryRow(`SELECT id, title, body, decided_at, COALESCE(document_id,''), COALESCE(doc_version_rel,''), COALESCE(doc_version_note,''), created_at, updated_at
		FROM decisions WHERE id=?`, id).
		Scan(&d.ID, &d.Title, &d.Body, &d.DecidedAt, &d.DocumentID, &d.DocVersionRel, &d.DocVersionNote, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// CreateDecision pins a new decision on the spine.
func (s *Store) CreateDecision(title, body, decidedAt string) (*Decision, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		title = "Untitled decision"
	}
	if decidedAt == "" {
		decidedAt = nowISO()
	}
	// Normalize date-only to full ISO for sorting stability
	if len(decidedAt) == 10 {
		decidedAt = decidedAt + "T12:00:00Z"
	}
	now := nowISO()
	d := &Decision{
		ID:        uuid.NewString(),
		Title:     title,
		Body:      body,
		DecidedAt: decidedAt,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.upsertDecision(d); err != nil {
		return nil, err
	}
	s.mu.Lock()
	_ = s.reindexFTS(d.ID, "decision", d.Title, d.Body)
	s.mu.Unlock()
	return d, nil
}

// UpdateDecision updates fields (empty documentId clears link when clearDoc is true).
func (s *Store) UpdateDecision(d *Decision) (*Decision, error) {
	cur, err := s.GetDecision(d.ID)
	if err != nil {
		return nil, err
	}
	if d.Title != "" {
		cur.Title = d.Title
	}
	cur.Body = d.Body
	if d.DecidedAt != "" {
		cur.DecidedAt = d.DecidedAt
		if len(cur.DecidedAt) == 10 {
			cur.DecidedAt = cur.DecidedAt + "T12:00:00Z"
		}
	}
	// Allow clearing by sending "-" sentinel for document id from API if needed;
	// normal path sets fields from client full object.
	cur.DocumentID = d.DocumentID
	if d.DocVersionRel != "" {
		cur.DocVersionRel = d.DocVersionRel
	}
	if d.DocVersionNote != "" {
		cur.DocVersionNote = d.DocVersionNote
	}
	cur.UpdatedAt = nowISO()
	if err := s.upsertDecision(cur); err != nil {
		return nil, err
	}
	s.mu.Lock()
	_ = s.reindexFTS(cur.ID, "decision", cur.Title, cur.Body)
	s.mu.Unlock()
	return cur, nil
}

// AttachDocumentVersion copies the current document file into decision-versions/
// and links it on the decision. Non-destructive to the live document.
func (s *Store) AttachDocumentVersion(decisionID, documentID, note string) (*Decision, error) {
	decisionID = strings.TrimSpace(decisionID)
	documentID = strings.TrimSpace(documentID)
	if decisionID == "" {
		return nil, fmt.Errorf("decision id required")
	}
	if documentID == "" {
		return nil, fmt.Errorf("select a document first")
	}
	d, err := s.GetDecision(decisionID)
	if err != nil {
		return nil, fmt.Errorf("decision: %w", err)
	}
	doc, err := s.GetDocument(documentID)
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}
	src := s.AbsPath(doc.RelPath)
	raw, err := os.ReadFile(src)
	if err != nil {
		return nil, fmt.Errorf("read document file: %w", err)
	}
	dir := filepath.Join(s.Root, VersionsDir, decisionID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create versions dir: %w", err)
	}
	stamp := time.Now().UTC().Format("20060102-150405")
	safeName := sanitizeFilename(doc.Filename)
	if safeName == "" || safeName == "note" {
		safeName = "document"
		if ext := filepath.Ext(doc.Filename); ext != "" {
			safeName += ext
		}
	}
	destName := stamp + "_" + safeName
	destAbs := filepath.Join(dir, destName)
	if err := os.WriteFile(destAbs, raw, 0o644); err != nil {
		return nil, fmt.Errorf("write snapshot: %w", err)
	}
	// Verify snapshot landed on disk
	if st, err := os.Stat(destAbs); err != nil || st.Size() != int64(len(raw)) {
		return nil, fmt.Errorf("snapshot verify failed for %s", destAbs)
	}
	rel := filepath.ToSlash(filepath.Join(VersionsDir, decisionID, destName))
	d.DocumentID = documentID
	d.DocVersionRel = rel
	if note == "" {
		note = fmt.Sprintf("Version of %s at %s (%d bytes)", doc.Filename, stamp, len(raw))
	}
	d.DocVersionNote = note
	d.UpdatedAt = nowISO()
	if err := s.upsertDecision(d); err != nil {
		return nil, fmt.Errorf("save decision: %w", err)
	}
	// Re-read to ensure JSON fields round-trip for the UI
	out, err := s.GetDecision(decisionID)
	if err != nil {
		return d, nil
	}
	return out, nil
}

// DeleteDecision removes the decision (version files kept on disk for audit).
func (s *Store) DeleteDecision(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.db.Exec(`DELETE FROM decisions WHERE id=?`, id); err != nil {
		return err
	}
	return s.removeFTS(id)
}

func (s *Store) upsertDecision(d *Decision) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`
INSERT INTO decisions(id, title, body, decided_at, document_id, doc_version_rel, doc_version_note, created_at, updated_at)
VALUES(?,?,?,?,?,?,?,?,?)
ON CONFLICT(id) DO UPDATE SET
  title=excluded.title, body=excluded.body, decided_at=excluded.decided_at,
  document_id=excluded.document_id, doc_version_rel=excluded.doc_version_rel,
  doc_version_note=excluded.doc_version_note, updated_at=excluded.updated_at
`, d.ID, d.Title, d.Body, d.DecidedAt, nullIfEmpty(d.DocumentID), nullIfEmpty(d.DocVersionRel),
		nullIfEmpty(d.DocVersionNote), d.CreatedAt, d.UpdatedAt)
	return err
}
