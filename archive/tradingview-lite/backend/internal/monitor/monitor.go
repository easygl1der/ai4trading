package monitor

import (
	"context"
	"log"
	"sync"
	"time"

	"tradingview-lite/backend/internal/alert"
	"tradingview-lite/backend/internal/market"
	"tradingview-lite/backend/internal/store"
)

type Result struct {
	StartedAt    time.Time     `json:"started_at"`
	FinishedAt   time.Time     `json:"finished_at"`
	Symbols      int           `json:"symbols"`
	Quotes       int           `json:"quotes"`
	BarsUpserted int           `json:"bars_upserted"`
	Events       []alert.Event `json:"events"`
	Errors       []string      `json:"errors,omitempty"`
}

type MarketClient interface {
	FetchQuote(ctx context.Context, symbol string) (*market.Quote, error)
	FetchBars(ctx context.Context, symbol, rangeValue, interval string, includePrePost bool) (*market.BarsResponse, error)
}

type Monitor struct {
	store     store.Store
	market    MarketClient
	evaluator *alert.Evaluator

	mu      sync.Mutex
	running bool
	last    *Result
}

func New(store store.Store, market MarketClient, evaluator *alert.Evaluator) *Monitor {
	return &Monitor{
		store:     store,
		market:    market,
		evaluator: evaluator,
	}
}

func (m *Monitor) PollOnce(ctx context.Context) (*Result, error) {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		result := &Result{
			StartedAt:  time.Now().UTC(),
			FinishedAt: time.Now().UTC(),
			Errors:     []string{"poll already running"},
		}
		return result, nil
	}
	m.running = true
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		m.running = false
		m.mu.Unlock()
	}()

	result := &Result{StartedAt: time.Now().UTC()}
	items, err := m.store.ListWatchlist(ctx)
	if err != nil {
		return result, err
	}
	result.Symbols = len(items)

	for _, item := range items {
		if err := m.pollSymbol(ctx, item.Symbol, result); err != nil {
			result.Errors = append(result.Errors, item.Symbol+": "+err.Error())
		}
	}

	events, err := m.evaluator.EvaluateAll(ctx)
	if err != nil {
		result.Errors = append(result.Errors, "alerts: "+err.Error())
	} else {
		result.Events = events
	}
	result.FinishedAt = time.Now().UTC()

	m.mu.Lock()
	m.last = result
	m.mu.Unlock()

	return result, nil
}

func (m *Monitor) LastResult() *Result {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.last == nil {
		return nil
	}
	cp := *m.last
	return &cp
}

func (m *Monitor) pollSymbol(ctx context.Context, symbol string, result *Result) error {
	quote, err := m.market.FetchQuote(ctx, symbol)
	if err != nil {
		return err
	}
	if err := m.store.UpsertQuote(ctx, *quote); err != nil {
		return err
	}
	result.Quotes++

	bars, err := m.market.FetchBars(ctx, symbol, "1d", "1m", false)
	if err != nil {
		return err
	}
	upserted, err := m.store.UpsertBars(ctx, bars.Symbol, bars.Interval, bars.Provider, bars.Bars)
	if err != nil {
		return err
	}
	result.BarsUpserted += upserted
	return nil
}

func (m *Monitor) Start(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = time.Minute
	}
	go func() {
		if result, err := m.PollOnce(ctx); err != nil {
			log.Printf("initial poll failed: %v", err)
		} else {
			log.Printf("initial poll complete: symbols=%d quotes=%d bars_upserted=%d events=%d errors=%d",
				result.Symbols, result.Quotes, result.BarsUpserted, len(result.Events), len(result.Errors))
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				result, err := m.PollOnce(ctx)
				if err != nil {
					log.Printf("poll failed: %v", err)
					continue
				}
				log.Printf("poll complete: symbols=%d quotes=%d bars_upserted=%d events=%d errors=%d",
					result.Symbols, result.Quotes, result.BarsUpserted, len(result.Events), len(result.Errors))
			}
		}
	}()
}
