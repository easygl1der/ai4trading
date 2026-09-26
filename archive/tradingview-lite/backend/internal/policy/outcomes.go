package policy

import (
	"context"
	"fmt"
	"math"
	"time"

	"tradingview-lite/backend/internal/store"
)

type OutcomeEvaluator struct {
	Store           store.Store
	HorizonMinutes  []int
	CandidateWindow time.Duration
}

func (e OutcomeEvaluator) Evaluate(ctx context.Context, symbol string, asOf time.Time) (int, error) {
	if asOf.IsZero() {
		asOf = time.Now().UTC()
	}
	window := e.CandidateWindow
	if window <= 0 {
		window = 12 * time.Hour
	}
	horizons := e.HorizonMinutes
	if len(horizons) == 0 {
		horizons = []int{15, 30, 60, 180}
	}
	candidates, err := e.Store.ListPolicyCandidates(ctx, store.PolicyCandidateFilter{
		Symbol: symbol,
		Bounds: store.TimeRange{From: asOf.Add(-window), To: asOf, Limit: 10000},
	})
	if err != nil {
		return 0, err
	}
	written := 0
	for _, candidate := range candidates {
		if !candidate.Eligible || candidate.ReferencePrice <= 0 {
			continue
		}
		for _, horizon := range horizons {
			target := candidate.AsOf.Add(time.Duration(horizon) * time.Minute)
			if target.After(asOf) {
				continue
			}
			quotes, err := e.Store.ListQuoteSnapshots(ctx, candidate.Symbol, "yahoo_chart", store.TimeRange{
				From:  candidate.AsOf,
				To:    target,
				Limit: 10000,
			})
			if err != nil {
				return written, err
			}
			if len(quotes) == 0 {
				continue
			}
			observed := quotes[len(quotes)-1]
			maxUp := -math.MaxFloat64
			maxDown := math.MaxFloat64
			for _, quote := range quotes {
				if quote.Price <= 0 {
					continue
				}
				ret := 100 * (quote.Price/candidate.ReferencePrice - 1)
				maxUp = math.Max(maxUp, ret)
				maxDown = math.Min(maxDown, ret)
			}
			if maxUp == -math.MaxFloat64 || maxDown == math.MaxFloat64 || observed.Price <= 0 {
				continue
			}
			warning := ""
			if target.Sub(observed.ReceivedAt) > 2*time.Minute {
				warning = fmt.Sprintf("target horizon is under-sampled by %ds", int64(target.Sub(observed.ReceivedAt).Seconds()))
			}
			_, err = e.Store.UpsertPolicyCandidateOutcome(ctx, store.PolicyCandidateOutcome{
				CandidateAlertID: candidate.ID,
				HorizonMinutes:   horizon,
				TargetAt:         target,
				ObservedAt:       observed.ReceivedAt,
				ObservedPrice:    observed.Price,
				ReturnPercent:    100 * (observed.Price/candidate.ReferencePrice - 1),
				MaxUpPercent:     maxUp,
				MaxDownPercent:   maxDown,
				SampleCount:      len(quotes),
				Warning:          warning,
				EvaluatedAt:      asOf,
			})
			if err != nil {
				return written, err
			}
			written++
		}
	}
	return written, nil
}
