package store

import (
	"strings"
	"testing"
	"time"
)

func TestSelectRangeUsesMostRecentRowsWhenOnlyAsOfUpperBoundExists(t *testing.T) {
	asOf := time.Date(2026, time.July, 13, 13, 55, 0, 0, time.UTC)
	query, args := selectRange("SELECT * FROM market_quote_snapshots", "received_at", "NVDA", "", TimeRange{
		To:    asOf,
		Limit: 20,
	})
	if !strings.Contains(query, "ORDER BY received_at DESC LIMIT") || !strings.Contains(query, "AS recent_snapshots ORDER BY received_at ASC") {
		t.Fatalf("query must select the latest samples at or before as_of, got %q", query)
	}
	if len(args) != 3 || args[1] != asOf || args[2] != 20 {
		t.Fatalf("args=%v", args)
	}
}

func TestSelectRangeKeepsAscendingOrderForExplicitRange(t *testing.T) {
	from := time.Date(2026, time.July, 13, 13, 50, 0, 0, time.UTC)
	to := from.Add(5 * time.Minute)
	query, _ := selectRange("SELECT * FROM market_quote_snapshots", "received_at", "NVDA", "", TimeRange{From: from, To: to, Limit: 20})
	if strings.Contains(query, "recent_snapshots") || !strings.Contains(query, "ORDER BY received_at ASC LIMIT") {
		t.Fatalf("explicit interval must remain chronological, got %q", query)
	}
}
