package collector

import (
	"context"
	"testing"
	"time"

	"tradingview-lite/backend/internal/market"
	"tradingview-lite/backend/internal/options"
	"tradingview-lite/backend/internal/rwa"
	"tradingview-lite/backend/internal/store"
)

func TestCollectorsPersistProviderAwareSnapshots(t *testing.T) {
	ctx := context.Background()
	repo := store.NewMemoryStore([]string{"NVDA"})
	now := time.Date(2026, time.July, 10, 13, 31, 0, 0, time.UTC)
	marketClient := fakeMarket{now: now}
	optionsClient := fakeOptions{now: now}
	rwaClient := fakeRWA{now: now}
	c := New(repo, marketClient, optionsClient, rwaClient, Config{
		OptionSymbols: []string{"NVDA"},
		RWASymbols:    []string{"NVDA"},
		Retention: store.RetentionPolicy{
			QuoteSnapshots:   14 * 24 * time.Hour,
			RWASnapshots:     14 * 24 * time.Hour,
			OptionSnapshots:  90 * 24 * time.Hour,
			FeatureSnapshots: 90 * 24 * time.Hour,
			CollectionRuns:   90 * 24 * time.Hour,
		},
	})

	if result, err := c.CollectMarket(ctx); err != nil || result.SymbolsSucceeded != 1 {
		t.Fatalf("CollectMarket() result=%+v err=%v", result, err)
	}
	if result, err := c.CollectRWA(ctx); err != nil || result.SymbolsSucceeded != 1 {
		t.Fatalf("CollectRWA() result=%+v err=%v", result, err)
	}
	if result, err := c.CollectOptions(ctx); err != nil || result.RecordsWritten != 3 {
		t.Fatalf("CollectOptions() result=%+v err=%v", result, err)
	}
	if result, err := c.CollectFeatures(ctx); err != nil || result.SymbolsSucceeded != 1 {
		t.Fatalf("CollectFeatures() result=%+v err=%v", result, err)
	}

	quotes, err := repo.ListQuoteSnapshots(ctx, "NVDA", "yahoo_chart", store.TimeRange{})
	if err != nil || len(quotes) != 1 || quotes[0].Session != "regular" {
		t.Fatalf("quote snapshots=%+v err=%v", quotes, err)
	}
	rwaSnapshots, err := repo.ListRWASnapshots(ctx, "NVDA", "bitget_rwa", store.TimeRange{})
	if err != nil || len(rwaSnapshots) != 1 || rwaSnapshots[0].Ticker != "NVDAon" {
		t.Fatalf("RWA snapshots=%+v err=%v", rwaSnapshots, err)
	}
	rwaBars, err := repo.GetRWABars(ctx, "NVDA", "1m", 10)
	if err != nil || len(rwaBars) != 2 {
		t.Fatalf("RWA bars=%+v err=%v", rwaBars, err)
	}
	contracts, err := repo.ListOptionContracts(ctx, "NVDA", store.TimeRange{})
	if err != nil || len(contracts) != 2 || contracts[0].IVInputSource != "bid_ask_mid" {
		t.Fatalf("option contracts=%+v err=%v", contracts, err)
	}
	ivs, err := repo.ListIVSnapshots(ctx, "NVDA", store.TimeRange{})
	if err != nil || len(ivs) != 1 || ivs[0].AverageResolvedIV == 0 {
		t.Fatalf("IV snapshots=%+v err=%v", ivs, err)
	}
	features, err := repo.ListFeatureSnapshots(ctx, "NVDA", store.TimeRange{})
	if err != nil || len(features) != 1 || len(features[0].Payload) == 0 {
		t.Fatalf("feature snapshots=%+v err=%v", features, err)
	}
	summary, err := repo.GetDailySummary(ctx, "NVDA", now)
	if err != nil || summary == nil || summary.RegularOpen == 0 {
		t.Fatalf("daily summary=%+v err=%v", summary, err)
	}
}

func TestMemoryRetentionPreservesRecentSamples(t *testing.T) {
	ctx := context.Background()
	repo := store.NewMemoryStore(nil)
	now := time.Date(2026, time.July, 10, 13, 31, 0, 0, time.UTC)
	for _, receivedAt := range []time.Time{now.Add(-48 * time.Hour), now.Add(-time.Hour)} {
		if err := repo.AppendQuoteSnapshot(ctx, store.QuoteSnapshot{
			Symbol: "NVDA", Provider: "yahoo_chart", Price: 100, ProviderTime: receivedAt,
			ReceivedAt: receivedAt, Session: "regular",
		}); err != nil {
			t.Fatal(err)
		}
	}
	result, err := repo.PruneSnapshots(ctx, store.RetentionPolicy{QuoteSnapshots: 24 * time.Hour}, now)
	if err != nil || result.QuoteSnapshots != 1 {
		t.Fatalf("PruneSnapshots() result=%+v err=%v", result, err)
	}
	quotes, err := repo.ListQuoteSnapshots(ctx, "NVDA", "", store.TimeRange{})
	if err != nil || len(quotes) != 1 || !quotes[0].ReceivedAt.Equal(now.Add(-time.Hour)) {
		t.Fatalf("remaining quotes=%+v err=%v", quotes, err)
	}
}

func TestRequestBudgetWaitsForNextWindow(t *testing.T) {
	budget := newRequestBudget(1, 30*time.Millisecond)
	if err := budget.Take(context.Background()); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()
	if err := budget.Take(ctx); err == nil {
		t.Fatal("expected context timeout while budget is exhausted")
	}
	time.Sleep(35 * time.Millisecond)
	if err := budget.Take(context.Background()); err != nil {
		t.Fatalf("budget did not reset: %v", err)
	}
}

type fakeMarket struct{ now time.Time }

func (f fakeMarket) FetchQuote(ctx context.Context, symbol string) (*market.Quote, error) {
	return &market.Quote{
		Symbol:            symbol,
		Provider:          "yahoo_chart",
		Price:             210,
		RegularMarketTime: f.now,
		ReceivedAt:        f.now,
		DataAgeSeconds:    2,
	}, nil
}

func (f fakeMarket) FetchBars(ctx context.Context, symbol, rangeValue, interval string, includePrePost bool) (*market.BarsResponse, error) {
	return &market.BarsResponse{
		Symbol:            symbol,
		Provider:          "yahoo_chart",
		Interval:          "1m",
		ReceivedAt:        f.now,
		RegularMarketTime: f.now,
		Bars: []market.Bar{
			{Time: f.now.Add(-time.Minute), Open: 209, High: 210, Low: 208.5, Close: 209.5, Volume: 1000},
			{Time: f.now, Open: 209.5, High: 211, Low: 209, Close: 210, Volume: 1200},
		},
	}, nil
}

type fakeOptions struct{ now time.Time }

func (f fakeOptions) FetchChain(ctx context.Context, symbol string, nearStrikes int) (*options.ChainResponse, error) {
	return &options.ChainResponse{
		Symbol:         symbol,
		Provider:       options.ProviderYahooMCP,
		ReceivedAt:     f.now,
		ExpirationDate: f.now.Add(24 * time.Hour),
		Underlying: options.UnderlyingQuote{
			Symbol: symbol, RegularMarketPrice: 210, PreviousClose: 205, RegularMarketTime: f.now,
		},
		Calls: []options.Contract{{
			ContractSymbol: "NVDA_TEST_CALL", Type: "call", Strike: 210, Bid: 2, Ask: 2.2, Mid: 2.1,
			LastPrice: 2.1, IVInputPrice: 2.1, IVInputSource: "bid_ask_mid", ResolvedImpliedVol: 0.3,
		}},
		Puts: []options.Contract{{
			ContractSymbol: "NVDA_TEST_PUT", Type: "put", Strike: 210, Bid: 1.8, Ask: 2, Mid: 1.9,
			LastPrice: 1.9, IVInputPrice: 1.9, IVInputSource: "bid_ask_mid", ResolvedImpliedVol: 0.32,
		}},
		Summary: options.ExpectedMove{
			Spot: 210, ATMStrike: 210, ResolvedCallImpliedVolatility: 0.3, ResolvedPutImpliedVolatility: 0.32,
			AverageResolvedImpliedVolatility: 0.31, OneTradingDayExpectedMove: 4.1, OneTradingDayExpectedMovePercent: 1.95,
			ExpectedMove: 4, ExpectedMovePercent: 1.9, RiskFreeRate: 0.045,
		},
	}, nil
}

type fakeRWA struct{ now time.Time }

func (f fakeRWA) StockInfo(ctx context.Context, symbol string) (*rwa.StockInfo, error) {
	return &rwa.StockInfo{
		Symbol: symbol, Ticker: "NVDAon", Provider: rwa.ProviderBitgetRWA, DataSource: "ondo",
		LatestPrice: 210.2, MarketStatus: "open", ReceivedAt: f.now,
	}, nil
}

func (f fakeRWA) Kline(ctx context.Context, symbol, period string, size int) (*rwa.KlineResponse, error) {
	return &rwa.KlineResponse{
		Symbol: symbol, Ticker: "NVDAon", Provider: rwa.ProviderBitgetRWA, Period: "1m", ReceivedAt: f.now,
		Bars: []rwa.Bar{
			{Time: f.now.Add(-time.Minute), Open: 209, High: 210, Low: 208.5, Close: 209.5},
			{Time: f.now, Open: 209.5, High: 210.5, Low: 209, Close: 210.2},
		},
	}, nil
}
