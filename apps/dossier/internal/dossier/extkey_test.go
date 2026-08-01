package dossier

import "testing"

func TestExtKey(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{".PDF", ".pdf"},
		{"pdf", ".pdf"},
		{"foo.PDF", ".pdf"},
		{`C:\a\b.docx`, ".docx"},
		{"report.md", ".md"},
		{"", ""},
	}
	for _, c := range cases {
		got := ExtKey(c.in)
		if got != c.want {
			t.Fatalf("ExtKey(%q)=%q want %q", c.in, got, c.want)
		}
	}
}
