package timefmt

import (
	"fmt"
	"time"
	_ "time/tzdata"
)

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now()
}

type FakeClock struct {
	Current time.Time
}

func (f *FakeClock) Now() time.Time {
	return f.Current
}

var indonesianDays = [...]string{
	"Min", "Sen", "Sel", "Rab", "Kam", "Jum", "Sab",
}

var indonesianMonths = [...]string{
	"", "Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des",
}

func ParseUTC(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

func FormatUTC(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func FormatClock(t time.Time, loc *time.Location) string {
	if loc != nil {
		t = t.In(loc)
	}
	dayName := indonesianDays[t.Weekday()]
	return fmt.Sprintf("%s %02d:%02d", dayName, t.Hour(), t.Minute())
}

func FormatDateTime(t time.Time, loc *time.Location) string {
	if loc != nil {
		t = t.In(loc)
	}
	dayName := indonesianDays[t.Weekday()]
	monthName := indonesianMonths[t.Month()]
	return fmt.Sprintf("%s, %d %s %02d:%02d WIB", dayName, t.Day(), monthName, t.Hour(), t.Minute())
}

func RelativeID(t, now time.Time) string {
	diff := now.Sub(t)
	if diff < 0 {
		return "Baru saja"
	}

	hours := int(diff.Hours())
	if hours < 1 {
		mins := int(diff.Minutes())
		if mins <= 1 {
			return "Baru saja"
		}
		return fmt.Sprintf("%d menit lalu", mins)
	}
	if hours < 24 {
		return fmt.Sprintf("%d jam lalu", hours)
	}
	days := hours / 24
	return fmt.Sprintf("%d hari lalu", days)
}

// ShiftAt returns "pagi" (07:00-19:00) or "malam" (19:00-07:00)
func ShiftAt(t time.Time, pagiStartStr, malamStartStr string, loc *time.Location) string {
	if loc != nil {
		t = t.In(loc)
	}

	pagiH, pagiM := parseTimeOfDay(pagiStartStr, 7, 0)
	malamH, malamM := parseTimeOfDay(malamStartStr, 19, 0)

	minuteOfDay := t.Hour()*60 + t.Minute()
	pagiMinutes := pagiH*60 + pagiM
	malamMinutes := malamH*60 + malamM

	if minuteOfDay >= pagiMinutes && minuteOfDay < malamMinutes {
		return "pagi"
	}
	return "malam"
}

func parseTimeOfDay(s string, defH, defM int) (int, int) {
	if len(s) != 5 || s[2] != ':' {
		return defH, defM
	}
	var h, m int
	_, err := fmt.Sscanf(s, "%02d:%02d", &h, &m)
	if err != nil || h < 0 || h > 23 || m < 0 || m > 59 {
		return defH, defM
	}
	return h, m
}
