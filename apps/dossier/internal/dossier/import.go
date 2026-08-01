package dossier

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

// KindFromExt classifies a file by extension.
func KindFromExt(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".md", ".markdown", ".mdown":
		return "markdown"
	case ".txt", ".text", ".log", ".csv", ".tsv", ".json", ".yaml", ".yml", ".toml", ".xml", ".html", ".htm", ".css", ".js", ".ts", ".go", ".py", ".rs", ".c", ".h", ".cpp", ".java", ".sh", ".bat", ".ps1":
		return "text"
	case ".pdf":
		return "pdf"
	case ".docx":
		return "docx"
	case ".xlsx":
		return "xlsx"
	case ".odt":
		return "odt"
	case ".ods":
		return "ods"
	default:
		return "other"
	}
}

func mimeFromName(name string) string {
	ext := filepath.Ext(name)
	if ext == "" {
		return "application/octet-stream"
	}
	m := mime.TypeByExtension(ext)
	if m == "" {
		return "application/octet-stream"
	}
	return m
}

// uniqueName ensures dest filename doesn't collide.
func uniqueName(dir, name string) string {
	candidate := name
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	for i := 1; ; i++ {
		if _, err := os.Stat(filepath.Join(dir, candidate)); os.IsNotExist(err) {
			return candidate
		}
		candidate = fmt.Sprintf("%s (%d)%s", base, i, ext)
	}
}

// IsOfficeLockFile reports MS Office lock/temp siblings (~$report.docx) that must not be imported.
func IsOfficeLockFile(name string) bool {
	base := filepath.Base(name)
	return strings.HasPrefix(base, "~$")
}

// ImportFile copies a source file into the dossier documents folder and indexes it.
func (s *Store) ImportFile(srcPath string) (*Document, error) {
	srcPath = filepath.Clean(srcPath)
	info, err := os.Stat(srcPath)
	if err != nil {
		return nil, fmt.Errorf("stat source: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("cannot import a directory: %s", srcPath)
	}

	filename := filepath.Base(srcPath)
	if IsOfficeLockFile(filename) {
		return nil, fmt.Errorf("skip Office lock/temp file: %s", filename)
	}
	destDir := filepath.Join(s.Root, DocumentsDir)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return nil, err
	}
	destName := uniqueName(destDir, filename)
	destAbs := filepath.Join(destDir, destName)
	rel := filepath.ToSlash(filepath.Join(DocumentsDir, destName))

	if err := copyFile(srcPath, destAbs); err != nil {
		return nil, err
	}

	return s.indexDocumentFile(destAbs, destName, rel, info.Size(), info.ModTime().UTC().Format(timeRFC3339(info)))
}

// ImportBytes writes raw bytes as a new document (used by drag-and-drop import).
func (s *Store) ImportBytes(filename string, data []byte) (*Document, error) {
	filename = filepath.Base(strings.TrimSpace(filename))
	if filename == "" || filename == "." || filename == ".." {
		filename = "dropped-file"
	}
	// Strip any path separators that survived Base on exotic inputs
	filename = strings.ReplaceAll(filename, string(filepath.Separator), "_")
	if IsOfficeLockFile(filename) {
		return nil, fmt.Errorf("skip Office lock/temp file: %s", filename)
	}
	if len(data) == 0 {
		// allow empty files; still attach
		data = []byte{}
	}
	destDir := filepath.Join(s.Root, DocumentsDir)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return nil, err
	}
	destName := uniqueName(destDir, filename)
	destAbs := filepath.Join(destDir, destName)
	rel := filepath.ToSlash(filepath.Join(DocumentsDir, destName))
	if err := os.WriteFile(destAbs, data, 0o644); err != nil {
		return nil, err
	}
	info, err := os.Stat(destAbs)
	if err != nil {
		return nil, err
	}
	return s.indexDocumentFile(destAbs, destName, rel, info.Size(), info.ModTime().UTC().Format(timeRFC3339(info)))
}

// ImportBase64 decodes base64 payload then ImportBytes.
func (s *Store) ImportBase64(filename, b64 string) (*Document, error) {
	// Allow data URL prefix
	if i := strings.Index(b64, ","); i >= 0 && strings.Contains(b64[:i], "base64") {
		b64 = b64[i+1:]
	}
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		// try raw / no padding
		data, err = base64.RawStdEncoding.DecodeString(b64)
		if err != nil {
			return nil, fmt.Errorf("decode base64: %w", err)
		}
	}
	return s.ImportBytes(filename, data)
}

// AttachExisting indexes a file already inside the dossier documents folder
// without copying (used by watcher / rescan).
func (s *Store) AttachExisting(absPath string) (*Document, error) {
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}
	if IsOfficeLockFile(info.Name()) {
		return nil, fmt.Errorf("skip Office lock/temp file: %s", info.Name())
	}
	rel, err := filepath.Rel(s.Root, absPath)
	if err != nil {
		return nil, err
	}
	rel = filepath.ToSlash(rel)
	return s.indexDocumentFile(absPath, filepath.Base(absPath), rel, info.Size(), info.ModTime().UTC().Format(timeRFC3339(info)))
}

func timeRFC3339(info os.FileInfo) string {
	return info.ModTime().UTC().Format("2006-01-02T15:04:05Z")
}

func (s *Store) indexDocumentFile(abs, filename, rel string, size int64, mtime string) (*Document, error) {
	kind := KindFromExt(filename)
	var body string
	var flags *PDFFlags
	var office *OfficeFlags

	switch kind {
	case "markdown", "text":
		raw, err := os.ReadFile(abs)
		if err != nil {
			return nil, err
		}
		if !utf8.Valid(raw) {
			// Still attach; don't index binary garbage as text
			body = ""
			kind = "other"
		} else {
			body = string(raw)
		}
	case "pdf":
		text, pf, err := ExtractPDFText(abs)
		if err != nil {
			// Still attach the file
			flags = pf
			if flags == nil {
				flags = &PDFFlags{ExtractionNote: err.Error()}
			}
		} else {
			body = text
			flags = pf
		}
	case "docx", "xlsx", "odt", "ods":
		text, of, err := ExtractOfficeText(abs)
		office = of
		if err != nil {
			if office == nil {
				office = &OfficeFlags{Format: kind, Partial: true, ExtractionNote: err.Error()}
			}
			body = text // may be empty
		} else {
			body = text
		}
	default:
		body = ""
	}

	// Preserve existing ID if re-indexing same path
	existing, _ := s.GetDocumentByRelPath(rel)
	now := nowISO()
	d := &Document{
		ID:          uuid.NewString(),
		Filename:    filename,
		RelPath:     rel,
		MimeType:    mimeFromName(filename),
		SizeBytes:   size,
		Kind:        kind,
		PDFFlags:    flags,
		OfficeFlags: office,
		CreatedAt:   now,
		UpdatedAt:   now,
		Mtime:       mtime,
		IndexedAt:   now,
	}
	if existing != nil {
		d.ID = existing.ID
		d.CreatedAt = existing.CreatedAt
	}

	if err := s.UpsertDocument(d, body); err != nil {
		return nil, err
	}
	return d, nil
}

// ReadDocumentBody returns extracted/indexed body for preview (re-reads file).
func (s *Store) ReadDocumentBody(id string) (string, *Document, error) {
	d, err := s.GetDocument(id)
	if err != nil {
		return "", nil, err
	}
	abs := s.AbsPath(d.RelPath)
	switch d.Kind {
	case "markdown", "text":
		b, err := os.ReadFile(abs)
		if err != nil {
			return "", d, err
		}
		return string(b), d, nil
	case "pdf":
		text, flags, err := ExtractPDFText(abs)
		if flags != nil {
			d.PDFFlags = flags
		}
		if err != nil && text == "" {
			return "", d, err
		}
		return text, d, nil
	case "docx", "xlsx", "odt", "ods":
		text, of, err := ExtractOfficeText(abs)
		if of != nil {
			d.OfficeFlags = of
		}
		if err != nil && text == "" {
			return "", d, err
		}
		return text, d, nil
	default:
		return "", d, nil
	}
}

// RescanDocuments walks the documents folder and indexes new/changed files.
func (s *Store) RescanDocuments() (int, error) {
	dir := filepath.Join(s.Root, DocumentsDir)
	count := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		// Skip hidden and MS Office lock/temp files (~$report.docx)
		if strings.HasPrefix(info.Name(), ".") || IsOfficeLockFile(info.Name()) {
			return nil
		}
		rel, err := filepath.Rel(s.Root, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		existing, _ := s.GetDocumentByRelPath(rel)
		mtime := info.ModTime().UTC().Format("2006-01-02T15:04:05Z")
		if existing != nil && existing.Mtime == mtime && existing.SizeBytes == info.Size() {
			return nil
		}
		if _, err := s.AttachExisting(path); err != nil {
			return nil
		}
		count++
		return nil
	})
	return count, err
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
