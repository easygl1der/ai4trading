package session

import (
	"testing"
	"time"
)

func TestClassifyUS(t *testing.T) {
	tests := []struct {
		name string
		at   time.Time
		want string
	}{
		{"premarket", nyTime(4, 0), Premarket},
		{"regular open", nyTime(9, 30), Regular},
		{"regular close excluded", nyTime(16, 0), Postmarket},
		{"postmarket", nyTime(19, 59), Postmarket},
		{"offhours", nyTime(20, 0), Offhours},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ClassifyUS(test.at); got != test.want {
				t.Fatalf("ClassifyUS(%s) = %q, want %q", test.at, got, test.want)
			}
		})
	}
}

func TestIsPreOpenFiveMinutes(t *testing.T) {
	if !IsPreOpenFiveMinutes(nyTime(9, 25)) {
		t.Fatal("09:25 New York should be in pre-open window")
	}
	if IsPreOpenFiveMinutes(nyTime(9, 30)) {
		t.Fatal("09:30 New York should be outside pre-open window")
	}
}

func nyTime(hour, minute int) time.Time {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		loc = time.UTC
	}
	return time.Date(2026, time.July, 10, hour, minute, 0, 0, loc).UTC()
}
