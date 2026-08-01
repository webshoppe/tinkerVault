package dossier

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// OfficeFlags describes best-effort Office text extraction (mirrors PDF honesty).
type OfficeFlags struct {
	Format         string `json:"format"` // docx | xlsx | odt | ods
	Partial        bool   `json:"partial"`
	ExtractionNote string `json:"extractionNote,omitempty"`
	SheetCount     int    `json:"sheetCount,omitempty"`
}

var xmlTagRe = regexp.MustCompile(`<[^>]+>`)
var multiSpace = regexp.MustCompile(`[ \t\r\f]+`)

// ExtractOfficeText pulls plain text from .docx / .xlsx / .odt / .ods for FTS indexing.
// Lossy by design: no formulas results, charts, images, or rich formatting.
func ExtractOfficeText(path string) (body string, flags *OfficeFlags, err error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".docx":
		return extractDOCX(path)
	case ".xlsx":
		return extractXLSX(path)
	case ".odt":
		return extractODT(path)
	case ".ods":
		return extractODS(path)
	default:
		return "", nil, fmt.Errorf("unsupported office type %s", ext)
	}
}

func extractDOCX(path string) (string, *OfficeFlags, error) {
	flags := &OfficeFlags{Format: "docx", Partial: true}
	zr, err := zip.OpenReader(path)
	if err != nil {
		flags.ExtractionNote = "not a valid .docx package: " + err.Error()
		return "", flags, err
	}
	defer zr.Close()

	var docXML []byte
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				flags.ExtractionNote = err.Error()
				return "", flags, err
			}
			docXML, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				flags.ExtractionNote = err.Error()
				return "", flags, err
			}
			break
		}
	}
	if docXML == nil {
		flags.ExtractionNote = "word/document.xml missing — empty or non-standard docx"
		return "", flags, fmt.Errorf("missing document.xml")
	}

	// Prefer w:t text nodes via lightweight scan
	text := extractWMLText(docXML)
	if strings.TrimSpace(text) == "" {
		// fallback strip tags
		text = stripXMLToText(string(docXML))
	}
	text = cleanExtracted(text)
	flags.ExtractionNote = "Best-effort text from Word body. Tables/images/headers/footers/comments and formatting are incomplete or omitted."
	if text == "" {
		flags.ExtractionNote += " No extractable text found."
	}
	return text, flags, nil
}

func extractWMLText(docXML []byte) string {
	// Decode as generic tokens; collect character data under w:t
	dec := xml.NewDecoder(bytes.NewReader(docXML))
	var b strings.Builder
	inT := false
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "t" {
				inT = true
			} else if t.Name.Local == "tab" {
				b.WriteByte('\t')
			} else if t.Name.Local == "br" || t.Name.Local == "cr" {
				b.WriteByte('\n')
			} else if t.Name.Local == "p" {
				if b.Len() > 0 {
					b.WriteByte('\n')
				}
			}
		case xml.EndElement:
			if t.Name.Local == "t" {
				inT = false
			}
		case xml.CharData:
			if inT {
				b.Write(t)
			}
		}
	}
	return b.String()
}

func extractXLSX(path string) (string, *OfficeFlags, error) {
	flags := &OfficeFlags{Format: "xlsx", Partial: true}
	zr, err := zip.OpenReader(path)
	if err != nil {
		flags.ExtractionNote = "not a valid .xlsx package: " + err.Error()
		return "", flags, err
	}
	defer zr.Close()

	shared := []string{}
	sheets := map[string][]byte{}
	for _, f := range zr.File {
		switch {
		case f.Name == "xl/sharedStrings.xml":
			rc, err := f.Open()
			if err != nil {
				continue
			}
			raw, _ := io.ReadAll(rc)
			rc.Close()
			shared = parseSharedStrings(raw)
		case strings.HasPrefix(f.Name, "xl/worksheets/sheet") && strings.HasSuffix(f.Name, ".xml"):
			rc, err := f.Open()
			if err != nil {
				continue
			}
			raw, _ := io.ReadAll(rc)
			rc.Close()
			sheets[f.Name] = raw
		}
	}
	flags.SheetCount = len(sheets)
	if len(sheets) == 0 {
		flags.ExtractionNote = "No worksheets found in workbook"
		return "", flags, fmt.Errorf("no worksheets")
	}

	var b strings.Builder
	// stable sheet order by name
	names := make([]string, 0, len(sheets))
	for n := range sheets {
		names = append(names, n)
	}
	// simple sort
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			if names[j] < names[i] {
				names[i], names[j] = names[j], names[i]
			}
		}
	}
	for _, name := range names {
		b.WriteString("[")
		b.WriteString(filepath.Base(name))
		b.WriteString("]\n")
		b.WriteString(parseSheetCells(sheets[name], shared))
		b.WriteString("\n")
	}
	text := cleanExtracted(b.String())
	flags.ExtractionNote = "Best-effort cell text from worksheets. Formulas (cached values only when present), charts, pivot tables, images, and formatting are incomplete or omitted."
	if text == "" {
		flags.ExtractionNote += " No extractable cell text found."
	}
	return text, flags, nil
}

func parseSharedStrings(raw []byte) []string {
	dec := xml.NewDecoder(bytes.NewReader(raw))
	var out []string
	var cur strings.Builder
	inSI, inT := false, false
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "si":
				inSI = true
				cur.Reset()
			case "t":
				if inSI {
					inT = true
				}
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "t":
				inT = false
			case "si":
				inSI = false
				out = append(out, cur.String())
			}
		case xml.CharData:
			if inT {
				cur.Write(t)
			}
		}
	}
	return out
}

func parseSheetCells(raw []byte, shared []string) string {
	dec := xml.NewDecoder(bytes.NewReader(raw))
	var b strings.Builder
	var cellType string
	var inV, inT, inIs bool
	var val strings.Builder
	rowHad := false
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "row":
				if rowHad {
					b.WriteByte('\n')
				}
				rowHad = true
			case "c":
				cellType = ""
				for _, a := range t.Attr {
					if a.Name.Local == "t" {
						cellType = a.Value
					}
				}
				val.Reset()
			case "v":
				inV = true
				val.Reset()
			case "t":
				// inline string
				inT = true
				if cellType == "inlineStr" || inIs {
					val.Reset()
				}
			case "is":
				inIs = true
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "v":
				inV = false
				s := val.String()
				if cellType == "s" {
					// shared string index
					if i, err := strconv.Atoi(strings.TrimSpace(s)); err == nil && i >= 0 && i < len(shared) {
						s = shared[i]
					}
				}
				if s != "" {
					if b.Len() > 0 && !strings.HasSuffix(b.String(), "\n") && !strings.HasSuffix(b.String(), "\t") {
						// between cells
					}
					// tab-separate cells on a row
					if rowHad && b.Len() > 0 {
						// check last char
						str := b.String()
						if len(str) > 0 && str[len(str)-1] != '\n' {
							b.WriteByte('\t')
						}
					}
					b.WriteString(s)
				}
			case "t":
				inT = false
				if cellType == "inlineStr" || inIs {
					s := val.String()
					if s != "" {
						str := b.String()
						if len(str) > 0 && str[len(str)-1] != '\n' {
							b.WriteByte('\t')
						}
						b.WriteString(s)
					}
				}
			case "is":
				inIs = false
			}
		case xml.CharData:
			if inV || inT {
				val.Write(t)
			}
		}
	}
	return b.String()
}

// extractODT pulls best-effort plain text from an OpenDocument Text (.odt) package.
func extractODT(path string) (string, *OfficeFlags, error) {
	flags := &OfficeFlags{Format: "odt", Partial: true}
	raw, err := readZipMember(path, "content.xml")
	if err != nil {
		flags.ExtractionNote = "not a valid .odt package: " + err.Error()
		return "", flags, err
	}
	text := extractODFText(raw, false)
	text = cleanExtracted(text)
	flags.ExtractionNote = "Best-effort text from OpenDocument body (content.xml). Styles, images, headers/footers, and embedded objects are incomplete or omitted."
	if text == "" {
		flags.ExtractionNote += " No extractable text found."
	}
	return text, flags, nil
}

// extractODS pulls best-effort cell text from an OpenDocument Spreadsheet (.ods) package.
func extractODS(path string) (string, *OfficeFlags, error) {
	flags := &OfficeFlags{Format: "ods", Partial: true}
	raw, err := readZipMember(path, "content.xml")
	if err != nil {
		flags.ExtractionNote = "not a valid .ods package: " + err.Error()
		return "", flags, err
	}
	text := extractODFText(raw, true)
	text = cleanExtracted(text)
	// Count table:table elements as sheet count (best-effort)
	flags.SheetCount = strings.Count(string(raw), "table:table ")
	if flags.SheetCount == 0 {
		flags.SheetCount = strings.Count(string(raw), "<table:table")
	}
	flags.ExtractionNote = "Best-effort cell text from OpenDocument spreadsheet (content.xml). Formulas, charts, images, and formatting are incomplete or omitted."
	if text == "" {
		flags.ExtractionNote += " No extractable cell text found."
	}
	return text, flags, nil
}

func readZipMember(path, member string) ([]byte, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	for _, f := range zr.File {
		if f.Name == member {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			raw, err := io.ReadAll(rc)
			rc.Close()
			return raw, err
		}
	}
	return nil, fmt.Errorf("%s missing", member)
}

// extractODFText walks ODF content.xml. When spreadsheet is true, table rows/cells
// get lightweight separators; otherwise paragraph breaks only.
func extractODFText(docXML []byte, spreadsheet bool) string {
	dec := xml.NewDecoder(bytes.NewReader(docXML))
	var b strings.Builder
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "s" {
				// text:s is a space run (often empty body)
				b.WriteByte(' ')
			}
		case xml.EndElement:
			local := t.Name.Local
			switch local {
			case "p", "h":
				if b.Len() > 0 {
					b.WriteByte('\n')
				}
			case "line-break":
				b.WriteByte('\n')
			case "table-cell":
				if spreadsheet {
					b.WriteByte('\t')
				}
			case "table-row":
				if spreadsheet {
					b.WriteByte('\n')
				}
			}
		case xml.CharData:
			b.Write(t)
		}
	}
	if b.Len() == 0 {
		// Fallback: strip all tags if structured walk found nothing
		return stripXMLToText(string(docXML))
	}
	return b.String()
}

func stripXMLToText(s string) string {
	s = xmlTagRe.ReplaceAllString(s, " ")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&apos;", "'")
	return s
}

func cleanExtracted(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	lines := strings.Split(s, "\n")
	var out []string
	for _, ln := range lines {
		ln = multiSpace.ReplaceAllString(ln, " ")
		ln = strings.TrimSpace(ln)
		if ln != "" {
			out = append(out, ln)
		}
	}
	return strings.Join(out, "\n")
}
