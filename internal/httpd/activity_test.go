package httpd

import (
	"testing"
	"time"
)

func TestSplitAcrossDays_SingleDay(t *testing.T) {
	loc := time.UTC
	start := time.Date(2024, 1, 15, 10, 0, 0, 0, loc)
	end := time.Date(2024, 1, 15, 11, 30, 0, 0, loc)

	acc := make(map[string]int)
	splitAcrossDays(start, end, acc)

	if len(acc) != 1 {
		t.Fatalf("expected 1 day entry, got %d", len(acc))
	}
	if acc["2024-01-15"] != 90 {
		t.Errorf("minutes = %d, want 90", acc["2024-01-15"])
	}
}

func TestSplitAcrossDays_SpansMidnight(t *testing.T) {
	loc := time.UTC
	start := time.Date(2024, 1, 15, 23, 0, 0, 0, loc)
	end := time.Date(2024, 1, 16, 1, 0, 0, 0, loc)

	acc := make(map[string]int)
	splitAcrossDays(start, end, acc)

	if acc["2024-01-15"] != 60 {
		t.Errorf("2024-01-15 = %d minutes, want 60", acc["2024-01-15"])
	}
	if acc["2024-01-16"] != 60 {
		t.Errorf("2024-01-16 = %d minutes, want 60", acc["2024-01-16"])
	}
}

func TestSplitAcrossDays_SpansThreeDays(t *testing.T) {
	loc := time.UTC
	// 23:00 on day 1 → 01:00 on day 3 (25 hours)
	start := time.Date(2024, 1, 1, 23, 0, 0, 0, loc)
	end := time.Date(2024, 1, 3, 0, 0, 0, 0, loc)

	acc := make(map[string]int)
	splitAcrossDays(start, end, acc)

	if acc["2024-01-01"] != 60 {
		t.Errorf("2024-01-01 = %d minutes, want 60", acc["2024-01-01"])
	}
	if acc["2024-01-02"] != 1440 {
		t.Errorf("2024-01-02 = %d minutes, want 1440 (full day)", acc["2024-01-02"])
	}
	if _, ok := acc["2024-01-03"]; ok {
		t.Errorf("2024-01-03 should not be present (end is midnight)")
	}
}

func TestSplitAcrossDays_ZeroDuration(t *testing.T) {
	loc := time.UTC
	t1 := time.Date(2024, 1, 15, 10, 0, 0, 0, loc)

	acc := make(map[string]int)
	splitAcrossDays(t1, t1, acc)

	if len(acc) != 0 {
		t.Errorf("expected no entries for zero duration, got %d", len(acc))
	}
}

func TestSplitAcrossDays_Accumulates(t *testing.T) {
	loc := time.UTC
	acc := make(map[string]int)

	// Two separate intervals on the same day.
	splitAcrossDays(
		time.Date(2024, 1, 15, 10, 0, 0, 0, loc),
		time.Date(2024, 1, 15, 11, 0, 0, 0, loc),
		acc,
	)
	splitAcrossDays(
		time.Date(2024, 1, 15, 14, 0, 0, 0, loc),
		time.Date(2024, 1, 15, 15, 30, 0, 0, loc),
		acc,
	)

	if acc["2024-01-15"] != 150 {
		t.Errorf("accumulated minutes = %d, want 150", acc["2024-01-15"])
	}
}

func TestParseChannelID_Valid(t *testing.T) {
	// 32-character hex string = 16 bytes
	id, err := parseChannelID("0102030405060708090a0b0c0d0e0f10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(id) != 16 {
		t.Errorf("len = %d, want 16", len(id))
	}
	if id[0] != 0x01 || id[15] != 0x10 {
		t.Errorf("decoded bytes incorrect: %x", id)
	}
}

func TestParseChannelID_Invalid(t *testing.T) {
	// parseChannelID is hex.DecodeString — only invalid hex chars cause errors.
	cases := []string{
		"notHex",
		"gg" + "00000000000000000000000000000000",
		"zz",
	}
	for _, tc := range cases {
		if _, err := parseChannelID(tc); err == nil {
			t.Errorf("parseChannelID(%q): expected error, got nil", tc)
		}
	}
}
