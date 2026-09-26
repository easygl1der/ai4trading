package collector

import (
	"context"
	"testing"
	"time"

	"tradingview-lite/backend/internal/evidence"
	"tradingview-lite/backend/internal/store"
)

func TestEventCollectorStoresStrictTickerEvidenceWithoutFalsePromotion(t *testing.T) {
	now := time.Date(2026, time.July, 10, 14, 0, 0, 0, time.UTC)
	repo := store.NewMemoryStore([]string{"NVDA"})
	client := fakeEvidence{response: &evidence.SearchResponse{
		Success:  true,
		Metadata: evidence.ResponseMetadata{ProviderRoute: "single_stock", StrictTicker: true},
		Data: []evidence.NewsItem{
			{Title: "NVIDIA filing", URL: "https://www.sec.gov/Archives/nvda", Provider: "rss_feeds", DiscoveredAt: now.Format(time.RFC3339), Tickers: []string{"NVDA"}, EventType: "filing"},
			{Title: "Generic market item", URL: "https://example.test/market", Provider: "rss_feeds", DiscoveredAt: now.Format(time.RFC3339), Tickers: []string{"QQQ"}, EventType: "macro"},
		},
	}}
	collector := New(repo, nil, nil, nil, Config{
		EventSymbols:           []string{"NVDA"},
		EventRequestsPerMinute: 10,
		OfficialDomains:        []string{"sec.gov"},
	}).WithEvidence(client)

	result, err := collector.CollectEventEvidence(context.Background())
	if err != nil || result.RecordsWritten != 2 || result.SymbolsSucceeded != 1 {
		t.Fatalf("collect result=%+v err=%v", result, err)
	}
	evidenceRows, err := repo.ListEventEvidence(context.Background(), store.EventFilter{Symbol: "NVDA", Bounds: store.TimeRange{Limit: 10}})
	if err != nil || len(evidenceRows) != 1 {
		t.Fatalf("NVDA evidence=%+v err=%v", evidenceRows, err)
	}
	if evidenceRows[0].SourceTier != "T1_OFFICIAL" {
		t.Fatalf("source tier = %q", evidenceRows[0].SourceTier)
	}
	events, err := repo.ListMarketEvents(context.Background(), store.EventFilter{Symbol: "NVDA", Bounds: store.TimeRange{Limit: 10}})
	if err != nil || len(events) != 1 || events[0].PrimarySymbol != "NVDA" || events[0].EvidenceStatus != "supported" {
		t.Fatalf("market events=%+v err=%v", events, err)
	}
}

type fakeEvidence struct {
	response *evidence.SearchResponse
	err      error
}

func (f fakeEvidence) Search(ctx context.Context, request evidence.SearchRequest) (*evidence.SearchResponse, error) {
	return f.response, f.err
}
