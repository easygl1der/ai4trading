package alert

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"

	"tradingview-lite/backend/internal/market"
)

type Store interface {
	GetQuote(ctx context.Context, symbol string) (*market.Quote, error)
	GetBars(ctx context.Context, symbol, timeframe string, limit int) ([]market.Bar, error)
	ListAlerts(ctx context.Context) ([]AlertConfig, error)
	AddEvent(ctx context.Context, event Event) error
}

type Notifier interface {
	Send(ctx context.Context, event Event) error
}

type Evaluator struct {
	store          Store
	notifiers      []Notifier
	dataStaleAfter time.Duration
	cooldown       time.Duration

	mu        sync.Mutex
	lastFired map[string]time.Time
}

func NewEvaluator(store Store, dataStaleAfter time.Duration, notifiers ...Notifier) *Evaluator {
	return &Evaluator{
		store:          store,
		notifiers:      notifiers,
		dataStaleAfter: dataStaleAfter,
		cooldown:       10 * time.Minute,
		lastFired:      map[string]time.Time{},
	}
}

func (e *Evaluator) EvaluateAll(ctx context.Context) ([]Event, error) {
	configs, err := e.store.ListAlerts(ctx)
	if err != nil {
		return nil, err
	}

	var events []Event
	for _, cfg := range configs {
		if !cfg.Enabled {
			continue
		}
		evs, err := e.evaluateConfig(ctx, cfg)
		if err != nil {
			return nil, err
		}
		for _, event := range evs {
			if err := e.store.AddEvent(ctx, event); err != nil {
				return nil, err
			}
			e.notify(ctx, event)
			events = append(events, event)
		}
	}
	return events, nil
}

func (e *Evaluator) notify(ctx context.Context, event Event) {
	for _, notifier := range e.notifiers {
		if notifier == nil {
			continue
		}
		if err := notifier.Send(ctx, event); err != nil {
			log.Printf("notifier send failed: symbol=%s level=%s rule=%s err=%v", event.Symbol, event.Level, event.RuleType, err)
		}
	}
}

func (e *Evaluator) evaluateConfig(ctx context.Context, cfg AlertConfig) ([]Event, error) {
	quote, err := e.store.GetQuote(ctx, cfg.Symbol)
	if err != nil {
		return nil, err
	}
	if quote == nil {
		return nil, nil
	}

	now := time.Now().UTC()
	if e.isStale(*quote) {
		event := Event{
			ID:                newEventID(),
			AlertID:           cfg.ID,
			Symbol:            cfg.Symbol,
			Level:             LevelBlocked,
			RuleType:          RuleDataStale,
			Message:           fmt.Sprintf("%s data stale: age=%ds", cfg.Symbol, quote.DataAgeSeconds),
			DataAgeSeconds:    quote.DataAgeSeconds,
			ProviderTime:      quote.RegularMarketTime,
			ReceivedAt:        quote.ReceivedAt,
			TriggeredAt:       now,
			AnalysisOnTrigger: false,
		}
		if !e.allowFire(cfg.ID, Rule{Type: RuleDataStale}, event, now) {
			return nil, nil
		}
		return []Event{event}, nil
	}

	bars, err := e.store.GetBars(ctx, cfg.Symbol, "1m", 390)
	if err != nil {
		return nil, err
	}

	var events []Event
	for _, rule := range cfg.Rules {
		event, ok := e.evaluateRule(ctx, cfg, rule, *quote, bars, now)
		if !ok {
			continue
		}
		if !e.allowFire(cfg.ID, rule, event, now) {
			continue
		}
		events = append(events, event)
	}
	return events, nil
}

func (e *Evaluator) evaluateRule(ctx context.Context, cfg AlertConfig, rule Rule, quote market.Quote, bars []market.Bar, now time.Time) (Event, bool) {
	switch rule.Type {
	case RulePriceLevel:
		return e.evalPriceLevel(cfg, rule, quote, now)
	case RuleIntradayMove:
		return e.evalIntradayMove(cfg, rule, quote, bars, now)
	case RuleRelativeUnderperformance:
		return e.evalRelativeUnderperformance(ctx, cfg, rule, quote, bars, now)
	case RuleVolumeSpike:
		return e.evalVolumeSpike(cfg, rule, quote, bars, now)
	default:
		return Event{}, false
	}
}

func (e *Evaluator) evalPriceLevel(cfg AlertConfig, rule Rule, quote market.Quote, now time.Time) (Event, bool) {
	triggered := false
	switch strings.ToLower(rule.Operator) {
	case "below", "less_than", "<":
		triggered = quote.Price < rule.Value
	case "above", "greater_than", ">":
		triggered = quote.Price > rule.Value
	case "equal", "=":
		triggered = math.Abs(quote.Price-rule.Value) <= 0.0001
	}
	if !triggered {
		return Event{}, false
	}
	return e.event(cfg, rule, quote, now,
		fmt.Sprintf("%s price %.4f triggered %s %.4f", cfg.Symbol, quote.Price, rule.Operator, rule.Value),
		quote.Price,
		rule.Value,
		"",
	), true
}

func (e *Evaluator) evalIntradayMove(cfg AlertConfig, rule Rule, quote market.Quote, bars []market.Bar, now time.Time) (Event, bool) {
	if len(bars) == 0 || bars[0].Open == 0 {
		return Event{}, false
	}
	movePct := (quote.Price - bars[0].Open) / bars[0].Open * 100
	if math.Abs(movePct) < rule.ThresholdPct {
		return Event{}, false
	}
	return e.event(cfg, rule, quote, now,
		fmt.Sprintf("%s intraday move %.2f%% exceeded %.2f%%", cfg.Symbol, movePct, rule.ThresholdPct),
		movePct,
		rule.ThresholdPct,
		"",
	), true
}

func (e *Evaluator) evalRelativeUnderperformance(ctx context.Context, cfg AlertConfig, rule Rule, quote market.Quote, bars []market.Bar, now time.Time) (Event, bool) {
	if rule.Benchmark == "" || len(bars) == 0 || bars[0].Open == 0 {
		return Event{}, false
	}
	benchmarkQuote, err := e.store.GetQuote(ctx, rule.Benchmark)
	if err != nil || benchmarkQuote == nil {
		return Event{}, false
	}
	benchmarkBars, err := e.store.GetBars(ctx, rule.Benchmark, "1m", 390)
	if err != nil || len(benchmarkBars) == 0 || benchmarkBars[0].Open == 0 {
		return Event{}, false
	}

	symbolMove := (quote.Price - bars[0].Open) / bars[0].Open * 100
	benchmarkMove := (benchmarkQuote.Price - benchmarkBars[0].Open) / benchmarkBars[0].Open * 100
	underperformance := benchmarkMove - symbolMove
	if underperformance < rule.ThresholdPct {
		return Event{}, false
	}
	return e.event(cfg, rule, quote, now,
		fmt.Sprintf("%s underperformed %s by %.2f%%", cfg.Symbol, strings.ToUpper(rule.Benchmark), underperformance),
		underperformance,
		rule.ThresholdPct,
		strings.ToUpper(rule.Benchmark),
	), true
}

func (e *Evaluator) evalVolumeSpike(cfg AlertConfig, rule Rule, quote market.Quote, bars []market.Bar, now time.Time) (Event, bool) {
	if len(bars) < 21 {
		return Event{}, false
	}
	current := bars[len(bars)-1].Volume
	window := bars[len(bars)-21 : len(bars)-1]
	mean, std := meanStdVolume(window)
	if std == 0 {
		return Event{}, false
	}
	z := (current - mean) / std
	if z < rule.ZScore {
		return Event{}, false
	}
	return e.event(cfg, rule, quote, now,
		fmt.Sprintf("%s volume spike z=%.2f exceeded %.2f", cfg.Symbol, z, rule.ZScore),
		z,
		rule.ZScore,
		"",
	), true
}

func (e *Evaluator) event(cfg AlertConfig, rule Rule, quote market.Quote, now time.Time, message string, observed, threshold float64, benchmark string) Event {
	level := rule.EffectiveLevel()
	if cfg.AnalysisOnTrigger && level == LevelAlert {
		level = LevelAnalysisTrigger
	}
	return Event{
		ID:                newEventID(),
		AlertID:           cfg.ID,
		Symbol:            cfg.Symbol,
		Level:             level,
		RuleType:          rule.Type,
		Message:           message,
		ObservedValue:     observed,
		ThresholdValue:    threshold,
		Benchmark:         benchmark,
		DataAgeSeconds:    quote.DataAgeSeconds,
		ProviderTime:      quote.RegularMarketTime,
		ReceivedAt:        quote.ReceivedAt,
		TriggeredAt:       now,
		AnalysisOnTrigger: cfg.AnalysisOnTrigger,
	}
}

func (e *Evaluator) isStale(quote market.Quote) bool {
	if quote.RegularMarketTime.IsZero() {
		return true
	}
	return time.Since(quote.RegularMarketTime) > e.dataStaleAfter
}

func (e *Evaluator) allowFire(alertID string, rule Rule, event Event, now time.Time) bool {
	key := alertID + "|" + string(rule.Type) + "|" + event.Benchmark
	e.mu.Lock()
	defer e.mu.Unlock()

	last := e.lastFired[key]
	if !last.IsZero() && now.Sub(last) < e.cooldown {
		return false
	}
	e.lastFired[key] = now
	return true
}

func meanStdVolume(bars []market.Bar) (float64, float64) {
	if len(bars) == 0 {
		return 0, 0
	}
	sum := 0.0
	for _, bar := range bars {
		sum += bar.Volume
	}
	mean := sum / float64(len(bars))
	variance := 0.0
	for _, bar := range bars {
		delta := bar.Volume - mean
		variance += delta * delta
	}
	return mean, math.Sqrt(variance / float64(len(bars)))
}

func newEventID() string {
	return fmt.Sprintf("ev_%d", time.Now().UTC().UnixNano())
}
