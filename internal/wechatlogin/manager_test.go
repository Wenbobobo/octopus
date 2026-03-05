package wechatlogin

import (
	"testing"
	"time"
)

func TestParseClock(t *testing.T) {
	h, m, err := parseClock("03:15")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if h != 3 || m != 15 {
		t.Fatalf("unexpected parse result: %d:%d", h, m)
	}

	if _, _, err := parseClock("24:00"); err == nil {
		t.Fatalf("expected error for invalid hour")
	}
}

func TestNextClockTime(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	now := time.Date(2026, 2, 26, 2, 59, 0, 0, loc)
	next := nextClockTime(now, 3, 0)
	if next.Day() != now.Day() || next.Hour() != 3 || next.Minute() != 0 {
		t.Fatalf("unexpected next time: %v", next)
	}

	now = time.Date(2026, 2, 26, 3, 0, 1, 0, loc)
	next = nextClockTime(now, 3, 0)
	if !next.After(now) {
		t.Fatalf("next should be after now, got now=%v next=%v", now, next)
	}
	if next.Day() == now.Day() {
		t.Fatalf("expected next day trigger, got now=%v next=%v", now, next)
	}
}
