package session

import "time"

const (
	Premarket  = "premarket"
	Regular    = "regular"
	Postmarket = "postmarket"
	Offhours   = "offhours"
)

func ClassifyUS(t time.Time) string {
	if t.IsZero() {
		return Offhours
	}
	ny := t.In(newYork())
	minutes := ny.Hour()*60 + ny.Minute()
	switch {
	case minutes >= 4*60 && minutes < 9*60+30:
		return Premarket
	case minutes >= 9*60+30 && minutes < 16*60:
		return Regular
	case minutes >= 16*60 && minutes < 20*60:
		return Postmarket
	default:
		return Offhours
	}
}

func TradingDate(t time.Time) time.Time {
	if t.IsZero() {
		return time.Time{}
	}
	ny := t.In(newYork())
	return time.Date(ny.Year(), ny.Month(), ny.Day(), 0, 0, 0, 0, time.UTC)
}

func IsPreOpenFiveMinutes(t time.Time) bool {
	if t.IsZero() {
		return false
	}
	ny := t.In(newYork())
	minutes := ny.Hour()*60 + ny.Minute()
	return minutes >= 9*60+25 && minutes < 9*60+30
}

func newYork() *time.Location {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return time.UTC
	}
	return loc
}
