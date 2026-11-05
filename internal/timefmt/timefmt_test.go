package timefmt

import (
	"testing"
	"time"
)

func jakarta(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		t.Fatalf("LoadLocation error: %v", err)
	}
	return loc
}

func TestShiftAtBoundaries(t *testing.T) {
	loc := jakarta(t)

	cases := []struct {
		hour, minute int
		want         string
	}{
		{6, 59, "malam"},
		{7, 0, "pagi"},
		{12, 0, "pagi"},
		{18, 59, "pagi"},
		{19, 0, "malam"},
		{23, 59, "malam"},
		{0, 0, "malam"},
	}

	for _, tc := range cases {
		now := time.Date(2026, 11, 5, tc.hour, tc.minute, 0, 0, loc)
		if got := ShiftAt(now, "07:00", "19:00", loc); got != tc.want {
			t.Errorf("%02d:%02d: expected %s, got %s", tc.hour, tc.minute, tc.want, got)
		}
	}
}

func TestShiftAtCustomHours(t *testing.T) {
	loc := jakarta(t)
	now := time.Date(2026, 11, 5, 6, 30, 0, 0, loc)

	if got := ShiftAt(now, "06:00", "18:00", loc); got != "pagi" {
		t.Fatalf("expected pagi, got %s", got)
	}
	if got := ShiftAt(now, "07:00", "19:00", loc); got != "malam" {
		t.Fatalf("expected malam, got %s", got)
	}
}

func TestRelativeID(t *testing.T) {
	now := time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		then time.Time
		want string
	}{
		{now.Add(-30 * time.Second), "Baru saja"},
		{now.Add(-5 * time.Minute), "5 menit lalu"},
		{now.Add(-3 * time.Hour), "3 jam lalu"},
		{now.Add(-50 * time.Hour), "2 hari lalu"},
		{now.Add(2 * time.Hour), "Baru saja"},
	}

	for _, tc := range cases {
		if got := RelativeID(tc.then, now); got != tc.want {
			t.Errorf("RelativeID(%v): expected %s, got %s", tc.then, tc.want, got)
		}
	}
}

func TestFormatUTCAndParseUTC(t *testing.T) {
	original := time.Date(2026, 11, 5, 10, 30, 0, 0, time.UTC)
	formatted := FormatUTC(original)

	parsed, err := ParseUTC(formatted)
	if err != nil {
		t.Fatalf("ParseUTC error: %v", err)
	}
	if !parsed.Equal(original) {
		t.Fatalf("Round-trip UTC gagal: %s != %s", parsed, original)
	}
}

func TestFormatClockAndDateTime(t *testing.T) {
	loc := jakarta(t)
	// 10:30 UTC = 17:30 WIB, Kamis.
	when := time.Date(2026, 11, 5, 10, 30, 0, 0, time.UTC)

	if got := FormatClock(when, loc); got != "Kam 17:30" {
		t.Fatalf("FormatClock: got %q", got)
	}
	if got := FormatDateTime(when, loc); got != "Kam, 5 Nov 17:30 WIB" {
		t.Fatalf("FormatDateTime: got %q", got)
	}
}
