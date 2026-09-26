package riskcontext

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"tradingview-lite/backend/internal/store"
)

func TestBuildExcludesFutureReceivedEvidence(t *testing.T) {
	ctx := context.Background()
	asOf := time.Date(2026, time.July, 10, 14, 0, 0, 0, time.UTC)
	repo := store.NewMemoryStore([]string{"NVDA"})
	if err := repo.AppendQuoteSnapshot(ctx, store.QuoteSnapshot{
		Symbol: "NVDA", Provider: "yahoo_chart", Price: 100, ProviderTime: asOf.Add(-5 * time.Second), ReceivedAt: asOf.Add(-5 * time.Second), DataAgeSeconds: 5, Session: "regular",
	}); err != nil {
		t.Fatal(err)
	}
	run, err := repo.AddEventIngestionRun(ctx, store.EventIngestionRun{
		Provider: "firecrawl_news", Query: "NVDA latest news", Scope: "symbol", RequestedSymbols: json.RawMessage(`["NVDA"]`), StrictTicker: true,
		StartedAt: asOf.Add(-time.Minute), FinishedAt: asOf.Add(-30 * time.Second),
	})
	if err != nil {
		t.Fatal(err)
	}
	rows := []store.EventEvidence{
		{CanonicalURL: "https://www.sec.gov/one", HeadlineHash: "one", Title: "NVDA filing", Provider: "rss", SourceTier: "T1_OFFICIAL", ReceivedAt: asOf.Add(-20 * time.Second), MatchedSymbols: json.RawMessage(`["NVDA"]`), Scope: "symbol", EventType: "filing"},
		{CanonicalURL: "https://www.sec.gov/two", HeadlineHash: "two", Title: "future filing", Provider: "rss", SourceTier: "T1_OFFICIAL", ReceivedAt: asOf.Add(time.Second), MatchedSymbols: json.RawMessage(`["NVDA"]`), Scope: "symbol", EventType: "filing"},
	}
	if _, err := repo.SaveEventEvidence(ctx, run.ID, rows); err != nil {
		t.Fatal(err)
	}

	snapshot, err := (Builder{Store: repo, DataStaleAfter: time.Minute}).Build(ctx, "NVDA", "semis_memory", asOf)
	if err != nil {
		t.Fatal(err)
	}
	var eventIDs []int64
	if err := json.Unmarshal(snapshot.KnownEventIDs, &eventIDs); err != nil {
		t.Fatal(err)
	}
	if len(eventIDs) != 1 || snapshot.EvidenceCoverage != "supported" || snapshot.RiskState != "elevated" {
		t.Fatalf("snapshot=%+v eventIDs=%v", snapshot, eventIDs)
	}
}
