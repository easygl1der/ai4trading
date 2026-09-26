package features

import (
	"testing"
	"time"

	"tradingview-lite/backend/internal/market"
	"tradingview-lite/backend/internal/options"
)

func TestComputeFromDataExcludesStaleRangeInputs(t *testing.T) {
	now := time.Date(2026, time.July, 13, 3, 0, 0, 0, time.UTC)
	result, err := ComputeFromData(IntradayInput{
		ComputedAt: now,
		Quote: &market.Quote{
			Symbol: "NVDA", Price: 210, ReceivedAt: now, RegularMarketTime: now.Add(-10 * time.Minute), DataAgeSeconds: 600,
		},
		Bars: &market.BarsResponse{
			Symbol: "NVDA", ReceivedAt: now, RegularMarketTime: now.Add(-10 * time.Minute), DataAgeSeconds: 600,
			Bars: []market.Bar{
				{Time: now.Add(-2 * time.Minute), Open: 208, High: 209, Low: 207, Close: 208.5},
				{Time: now.Add(-time.Minute), Open: 208.5, High: 210, Low: 208, Close: 210},
			},
		},
		OptionChain: &options.ChainResponse{
			Underlying: options.UnderlyingQuote{PreviousClose: 205},
			Summary:    options.ExpectedMove{OneTradingDayExpectedMove: 5, OneTradingDayExpectedMovePercent: 2.4},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.ReasonableRange.SelectedSource != "unavailable" || result.ReasonableRange.SelectedOneDayMove != 0 {
		t.Fatalf("stale inputs produced range: %+v", result.ReasonableRange)
	}
	if len(result.Warnings) == 0 {
		t.Fatal("expected stale-input warning")
	}
}
