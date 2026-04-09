package report

import (
	"strings"
	"testing"
	"time"

	"github.com/kway-teow/git-digest/internal/git"
	"github.com/kway-teow/git-digest/internal/timequery"
)

func TestGenerateTextReportUsesWindowLabel(t *testing.T) {
	var out strings.Builder
	g := NewGenerator(FormatText, &out)
	window := timequery.Window{
		Start: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2026, 4, 9, 23, 59, 59, 0, time.UTC),
		Label: "Iterative Report",
	}

	err := g.GenerateReport("summary", []git.CommitInfo{{Hash: "12345678", Author: "cola", Date: time.Date(2026, 4, 9, 10, 0, 0, 0, time.UTC), Message: "feat: test"}}, window, timequery.LanguageEnglish)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := out.String()
	if !strings.Contains(text, "Iterative Report (2026-04-01 to 2026-04-09)") {
		t.Fatalf("missing window title: %s", text)
	}
	if !strings.Contains(text, "## AI Summary") {
		t.Fatalf("missing english summary heading: %s", text)
	}
}
