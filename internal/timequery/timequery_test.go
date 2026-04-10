package timequery

import (
	"testing"
	"time"
)

func TestDefaultSpecUsesLast7Days(t *testing.T) {
	spec := DefaultSpec()
	if spec.Kind != KindPreset {
		t.Fatalf("unexpected kind: %q", spec.Kind)
	}
	if spec.Period != PresetLast7Days {
		t.Fatalf("unexpected period: %q", spec.Period)
	}
}

func TestResolveSingleDay(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	now := time.Date(2026, 4, 9, 15, 30, 0, 0, loc)

	window, err := ResolveWithLanguage(Spec{Kind: KindSingleDay, On: "2026-04-09"}, loc, now, "zh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := window.Start.Format("2006-01-02 15:04:05"); got != "2026-04-09 00:00:00" {
		t.Fatalf("unexpected start: %s", got)
	}
	if got := window.End.Format("2006-01-02 15:04:05"); got != "2026-04-09 23:59:59" {
		t.Fatalf("unexpected end: %s", got)
	}
	if window.Label != "工作日报" {
		t.Fatalf("unexpected label: %q", window.Label)
	}
}

func TestResolvePresetThisWeek(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	now := time.Date(2026, 4, 9, 15, 30, 0, 0, loc)

	window, err := ResolveWithLanguage(Spec{Kind: KindPreset, Period: PresetThisWeek}, loc, now, "zh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := window.Start.Format("2006-01-02 15:04:05"); got != "2026-04-06 00:00:00" {
		t.Fatalf("unexpected start: %s", got)
	}
	if got := window.End.Format("2006-01-02 15:04:05"); got != "2026-04-09 15:30:00" {
		t.Fatalf("unexpected end: %s", got)
	}
	if window.Label != "工作周报" {
		t.Fatalf("unexpected label: %q", window.Label)
	}
}

func TestResolvePresetLastWeekMonthYear(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	now := time.Date(2026, 4, 9, 15, 30, 0, 0, loc)

	tests := []struct {
		period string
		start  string
		end    string
		label  string
	}{
		{PresetLastWeek, "2026-03-30 00:00:00", "2026-04-05 23:59:59", "Weekly Report"},
		{PresetLastMonth, "2026-03-01 00:00:00", "2026-03-31 23:59:59", "Monthly Report"},
		{PresetLastYear, "2025-01-01 00:00:00", "2025-12-31 23:59:59", "Yearly Report"},
	}

	for _, tt := range tests {
		window, err := ResolveWithLanguage(Spec{Kind: KindPreset, Period: tt.period}, loc, now, "en")
		if err != nil {
			t.Fatalf("period %s unexpected error: %v", tt.period, err)
		}
		if got := window.Start.Format("2006-01-02 15:04:05"); got != tt.start {
			t.Fatalf("period %s unexpected start: %s", tt.period, got)
		}
		if got := window.End.Format("2006-01-02 15:04:05"); got != tt.end {
			t.Fatalf("period %s unexpected end: %s", tt.period, got)
		}
		if window.Label != tt.label {
			t.Fatalf("period %s unexpected label: %q", tt.period, window.Label)
		}
	}
}

func TestResolveRangeRejectsIncompleteBounds(t *testing.T) {
	_, err := Resolve(Spec{Kind: KindRange, From: "2026-04-01"}, time.UTC, time.Now().UTC())
	if err == nil {
		t.Fatal("expected error for incomplete range")
	}
}

func TestResolveRangeRejectsReverseBounds(t *testing.T) {
	_, err := ResolveWithLanguage(
		Spec{Kind: KindRange, From: "2026-04-10", To: "2026-04-01"},
		time.UTC,
		time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC),
		"en",
	)
	if err == nil {
		t.Fatal("expected error for reverse range bounds")
	}
}
