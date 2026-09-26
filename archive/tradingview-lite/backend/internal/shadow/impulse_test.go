package shadow

import (
	"context"
	"testing"
	"time"

	"tradingview-lite/backend/internal/store"
)

func TestImpulseRejectionUsesSessionBaselineAndObservedElapsedTime(t *testing.T) {
	ctx := context.Background()
	repo := store.NewMemoryStore(nil)
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		loc = time.UTC
	}
	var asOf time.Time
	for day := 6; day <= 10; day++ {
		base := time.Date(2026, time.July, day, 9, 30, 0, 0, loc).UTC()
		for index := 0; index < 40; index++ {
			at := base.Add(time.Duration(index) * 30 * time.Second)
			price := 100 + float64(index)*0.002
			if day == 10 && index == 35 {
				price = 99.8
			}
			if day == 10 && index == 38 {
				price = 100.8
			}
			if day == 10 && index == 39 {
				price = 100.2
				asOf = at
			}
			if err := repo.AppendQuoteSnapshot(ctx, store.QuoteSnapshot{Symbol: "NVDA", Provider: "yahoo_chart", Price: price, ProviderTime: at, ReceivedAt: at, DataAgeSeconds: 1, Session: "regular"}); err != nil {
				t.Fatal(err)
			}
		}
	}

	observation, err := (Evaluator{Store: repo, Config: Config{MinSessions: 5, MinBaselineSamples: 30}}).Evaluate(ctx, "NVDA", asOf, &store.EventContextSnapshot{
		ID: 1, Symbol: "NVDA", AsOf: asOf, MarketDataFresh: true, RiskState: "normal", EvidenceCoverage: "current",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !observation.Eligible || observation.State != StateRejected {
		t.Fatalf("observation=%+v", observation)
	}
}

func TestImpulseRejectionRecordsNoBaselineInsteadOfGuessing(t *testing.T) {
	ctx := context.Background()
	repo := store.NewMemoryStore(nil)
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		loc = time.UTC
	}
	asOf := time.Date(2026, time.July, 10, 10, 0, 0, 0, loc).UTC()
	for _, point := range []struct {
		at    time.Time
		price float64
	}{
		{asOf.Add(-time.Minute), 100},
		{asOf, 101},
	} {
		if err := repo.AppendQuoteSnapshot(ctx, store.QuoteSnapshot{Symbol: "NVDA", Provider: "yahoo_chart", Price: point.price, ProviderTime: point.at, ReceivedAt: point.at, DataAgeSeconds: 1, Session: "regular"}); err != nil {
			t.Fatal(err)
		}
	}

	observation, err := (Evaluator{Store: repo}).Evaluate(ctx, "NVDA", asOf, &store.EventContextSnapshot{
		ID: 1, Symbol: "NVDA", AsOf: asOf, MarketDataFresh: true, RiskState: "normal", EvidenceCoverage: "current",
	})
	if err != nil {
		t.Fatal(err)
	}
	if observation.State != StateNoBaseline || observation.Eligible || observation.BlockedReason != "insufficient_regular_session_baseline" {
		t.Fatalf("observation=%+v", observation)
	}
}
