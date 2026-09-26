package shadow

import (
	"context"
	"encoding/json"
	"math"
	"sort"
	"time"

	"tradingview-lite/backend/internal/session"
	"tradingview-lite/backend/internal/store"
)

const (
	SignalImpulseRejection = "IMPULSE_REJECTION"
	StateIdle              = "IDLE"
	StateNoBaseline        = "NO_BASELINE"
	StateBlocked           = "BLOCKED"
	StateImpulseForming    = "IMPULSE_FORMING"
	StateContinuing        = "CONTINUING"
	StateRejected          = "REJECTED"
)

type Config struct {
	DataStaleAfter      time.Duration
	BaselineLookback    time.Duration
	RecentWindow        time.Duration
	RejectionWindow     time.Duration
	MinSessions         int
	MinBaselineSamples  int
	MinImpulseMovePct   float64
	MinImpulsePctMinute float64
	ThresholdVersion    string
}

type Evaluator struct {
	Store  store.Store
	Config Config
}

func (e Evaluator) Evaluate(ctx context.Context, symbol string, asOf time.Time, eventContext *store.EventContextSnapshot) (store.ShadowSignalObservation, error) {
	if asOf.IsZero() {
		asOf = time.Now().UTC()
	}
	cfg := normalizeConfig(e.Config)
	observation := store.ShadowSignalObservation{
		SignalType:       SignalImpulseRejection,
		Symbol:           symbol,
		AsOf:             asOf,
		ThresholdVersion: cfg.ThresholdVersion,
		State:            StateBlocked,
		CreatedAt:        time.Now().UTC(),
	}
	if eventContext == nil {
		return blocked(observation, "event_context_missing"), nil
	}
	observation.EventContextID = eventContext.ID
	if !eventContext.MarketDataFresh || eventContext.RiskState == "blocked" {
		return blocked(observation, "market_context_blocked"), nil
	}

	quotes, err := e.Store.ListQuoteSnapshots(ctx, symbol, "yahoo_chart", store.TimeRange{
		From:  asOf.Add(-cfg.BaselineLookback),
		To:    asOf,
		Limit: 10000,
	})
	if err != nil {
		return observation, err
	}
	if len(quotes) < 2 {
		return blocked(observation, "insufficient_quote_samples"), nil
	}
	latest := quotes[len(quotes)-1]
	latestAge := asOf.Sub(latest.ReceivedAt)
	if latestAge < 0 {
		latestAge = 0
	}
	if latest.DataAgeSeconds > int64(latestAge.Seconds()) {
		latestAge = time.Duration(latest.DataAgeSeconds) * time.Second
	}
	if latestAge > cfg.DataStaleAfter {
		return blocked(observation, "quote_stale"), nil
	}
	if session.ClassifyUS(latest.ProviderTime) != session.Regular {
		return blocked(observation, "outside_regular_us_equity_session"), nil
	}

	sessions, baselineSpeeds := baseline(quotes, cfg)
	if len(sessions) < cfg.MinSessions || len(baselineSpeeds) < cfg.MinBaselineSamples {
		observation.State = StateNoBaseline
		observation.Eligible = false
		observation.BlockedReason = "insufficient_regular_session_baseline"
		observation.ReasonCodes, _ = json.Marshal([]string{"baseline_history_insufficient"})
		observation.Metrics, _ = json.Marshal(map[string]any{
			"regularSessions": len(sessions),
			"baselineSamples": len(baselineSpeeds),
		})
		return observation, nil
	}

	anchor, found := earliestRecent(quotes, latest.ReceivedAt.Add(-cfg.RecentWindow))
	if !found || !latest.ReceivedAt.After(anchor.ReceivedAt) {
		return blocked(observation, "insufficient_recent_quote_window"), nil
	}
	delta := latest.ReceivedAt.Sub(anchor.ReceivedAt).Seconds()
	movePct := 100 * (latest.Price/anchor.Price - 1)
	speed := movePct / delta * 60
	threshold := math.Max(percentileAbs(baselineSpeeds, 0.95), cfg.MinImpulsePctMinute)
	metrics := map[string]any{
		"anchorPrice":       anchor.Price,
		"latestPrice":       latest.Price,
		"elapsedSeconds":    delta,
		"impulseMovePct":    movePct,
		"speedPctPerMinute": speed,
		"speedThreshold":    threshold,
		"regularSessions":   len(sessions),
		"baselineSamples":   len(baselineSpeeds),
		"selectedMovePct":   eventContext.SelectedMovePct,
		"evidenceCoverage":  eventContext.EvidenceCoverage,
	}
	observation.Eligible = true
	observation.State = StateIdle
	reasons := []string{"no_impulse_threshold_breach"}
	if math.Abs(movePct) >= cfg.MinImpulseMovePct && math.Abs(speed) >= threshold {
		reasons = []string{"impulse_speed_exceeds_symbol_baseline"}
		retracement, continuing := continuationState(quotes, latest, speed, cfg.RejectionWindow)
		metrics["retracementPct"] = retracement
		if math.Abs(retracement) >= math.Max(cfg.MinImpulseMovePct/2, threshold/2) {
			observation.State = StateRejected
			reasons = append(reasons, "failed_continuation_retracement")
		} else if continuing {
			observation.State = StateContinuing
			reasons = append(reasons, "new_extreme_holds")
		} else {
			observation.State = StateImpulseForming
			reasons = append(reasons, "continuation_pending")
		}
	}
	observation.ReasonCodes, _ = json.Marshal(reasons)
	observation.Metrics, _ = json.Marshal(metrics)
	return observation, nil
}

func normalizeConfig(cfg Config) Config {
	if cfg.DataStaleAfter <= 0 {
		cfg.DataStaleAfter = 2 * time.Minute
	}
	if cfg.BaselineLookback <= 0 {
		cfg.BaselineLookback = 14 * 24 * time.Hour
	}
	if cfg.RecentWindow <= 0 {
		cfg.RecentWindow = 2 * time.Minute
	}
	if cfg.RejectionWindow <= 0 {
		cfg.RejectionWindow = 3 * time.Minute
	}
	if cfg.MinSessions <= 0 {
		cfg.MinSessions = 5
	}
	if cfg.MinBaselineSamples <= 0 {
		cfg.MinBaselineSamples = 30
	}
	if cfg.MinImpulseMovePct <= 0 {
		cfg.MinImpulseMovePct = 0.15
	}
	if cfg.MinImpulsePctMinute <= 0 {
		cfg.MinImpulsePctMinute = 0.20
	}
	if cfg.ThresholdVersion == "" {
		cfg.ThresholdVersion = "impulse_shadow_v1"
	}
	return cfg
}

func blocked(observation store.ShadowSignalObservation, reason string) store.ShadowSignalObservation {
	observation.State = StateBlocked
	observation.Eligible = false
	observation.BlockedReason = reason
	observation.ReasonCodes, _ = json.Marshal([]string{reason})
	observation.Metrics, _ = json.Marshal(map[string]any{})
	return observation
}

func baseline(quotes []store.QuoteSnapshot, cfg Config) (map[string]bool, []float64) {
	sessions := map[string]bool{}
	speeds := make([]float64, 0)
	for index := 1; index < len(quotes); index++ {
		previous, current := quotes[index-1], quotes[index]
		if session.ClassifyUS(previous.ProviderTime) != session.Regular || session.ClassifyUS(current.ProviderTime) != session.Regular {
			continue
		}
		date := session.TradingDate(current.ProviderTime).Format("2006-01-02")
		sessions[date] = true
		delta := current.ReceivedAt.Sub(previous.ReceivedAt).Seconds()
		if delta <= 0 || delta > 120 || previous.Price <= 0 {
			continue
		}
		speeds = append(speeds, 100*(current.Price/previous.Price-1)/delta*60)
	}
	return sessions, speeds
}

func earliestRecent(quotes []store.QuoteSnapshot, cutoff time.Time) (store.QuoteSnapshot, bool) {
	for _, quote := range quotes {
		if !quote.ReceivedAt.Before(cutoff) {
			return quote, true
		}
	}
	return store.QuoteSnapshot{}, false
}

func continuationState(quotes []store.QuoteSnapshot, latest store.QuoteSnapshot, speed float64, window time.Duration) (float64, bool) {
	start := latest.ReceivedAt.Add(-window)
	extreme := latest.Price
	for _, quote := range quotes {
		if quote.ReceivedAt.Before(start) {
			continue
		}
		if speed >= 0 && quote.Price > extreme {
			extreme = quote.Price
		}
		if speed < 0 && quote.Price < extreme {
			extreme = quote.Price
		}
	}
	if speed >= 0 {
		return 100 * (latest.Price/extreme - 1), latest.Price >= extreme
	}
	return 100 * (latest.Price/extreme - 1), latest.Price <= extreme
}

func percentileAbs(values []float64, percentile float64) float64 {
	absolute := make([]float64, len(values))
	for index, value := range values {
		absolute[index] = math.Abs(value)
	}
	sort.Float64s(absolute)
	if len(absolute) == 0 {
		return 0
	}
	index := int(math.Ceil(percentile*float64(len(absolute)))) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(absolute) {
		index = len(absolute) - 1
	}
	return absolute[index]
}
