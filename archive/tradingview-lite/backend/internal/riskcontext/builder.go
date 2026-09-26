package riskcontext

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"tradingview-lite/backend/internal/session"
	"tradingview-lite/backend/internal/store"
)

type Builder struct {
	Store                 store.Store
	DataStaleAfter        time.Duration
	EvidenceFreshAfter    time.Duration
	IngestionHistoryLimit int
}

func (b Builder) Build(ctx context.Context, symbol, group string, asOf time.Time) (store.EventContextSnapshot, error) {
	if asOf.IsZero() {
		asOf = time.Now().UTC()
	}
	if b.DataStaleAfter <= 0 {
		b.DataStaleAfter = 2 * time.Minute
	}
	if b.EvidenceFreshAfter <= 0 {
		b.EvidenceFreshAfter = 20 * time.Minute
	}
	if b.IngestionHistoryLimit <= 0 {
		b.IngestionHistoryLimit = 500
	}

	quotes, err := b.Store.ListQuoteSnapshots(ctx, symbol, "", store.TimeRange{To: asOf, Limit: 20})
	if err != nil {
		return store.EventContextSnapshot{}, err
	}
	warnings := make([]string, 0, 3)
	reasons := make([]string, 0, 3)
	contextSnapshot := store.EventContextSnapshot{
		Symbol:           symbol,
		AsOf:             asOf,
		Session:          session.Offhours,
		EvidenceCoverage: "unavailable",
		RiskState:        "normal",
	}
	if len(quotes) == 0 {
		contextSnapshot.RiskState = "blocked"
		reasons = append(reasons, "market_quote_missing")
		warnings = append(warnings, "no_quote_snapshot_before_as_of")
	} else {
		quote := quotes[len(quotes)-1]
		age := int64(asOf.Sub(quote.ReceivedAt).Seconds())
		if quote.DataAgeSeconds > age {
			age = quote.DataAgeSeconds
		}
		if age < 0 {
			age = 0
		}
		contextSnapshot.QuoteAgeSeconds = age
		contextSnapshot.Session = quote.Session
		contextSnapshot.MarketDataFresh = age <= int64(b.DataStaleAfter.Seconds())
		if !contextSnapshot.MarketDataFresh {
			contextSnapshot.RiskState = "blocked"
			reasons = append(reasons, "market_data_stale")
			warnings = append(warnings, "latest_quote_exceeds_freshness_limit")
		}
	}

	features, err := b.Store.ListFeatureSnapshots(ctx, symbol, store.TimeRange{To: asOf, Limit: 1000})
	if err != nil {
		return store.EventContextSnapshot{}, err
	}
	if len(features) > 0 {
		contextSnapshot.SelectedMovePct = features[len(features)-1].SelectedMovePct
	}

	runs, err := b.Store.ListEventIngestionRuns(ctx, b.IngestionHistoryLimit)
	if err != nil {
		return store.EventContextSnapshot{}, err
	}
	for _, run := range runs {
		if run.FinishedAt.After(asOf) || run.ErrorSummary != "" || run.RateLimited || !requestedSymbol(run.RequestedSymbols, symbol) {
			continue
		}
		if contextSnapshot.LastIngestionAt.IsZero() || run.FinishedAt.After(contextSnapshot.LastIngestionAt) {
			contextSnapshot.LastIngestionAt = run.FinishedAt
		}
	}
	if !contextSnapshot.LastIngestionAt.IsZero() {
		if asOf.Sub(contextSnapshot.LastIngestionAt) <= b.EvidenceFreshAfter {
			contextSnapshot.EvidenceCoverage = "current"
		} else {
			contextSnapshot.EvidenceCoverage = "stale"
			warnings = append(warnings, "event_evidence_ingestion_stale")
		}
	} else {
		warnings = append(warnings, "no_successful_symbol_event_ingestion_before_as_of")
	}

	events, err := b.Store.ListMarketEvents(ctx, store.EventFilter{
		Symbol: symbol,
		Group:  group,
		Bounds: store.TimeRange{To: asOf, Limit: 100},
	})
	if err != nil {
		return store.EventContextSnapshot{}, err
	}
	eventIDs := make([]int64, 0, len(events))
	for _, event := range events {
		eventIDs = append(eventIDs, event.ID)
		if event.EvidenceStatus == "supported" && contextSnapshot.RiskState != "blocked" {
			contextSnapshot.RiskState = "elevated"
			reasons = append(reasons, "supported_event_evidence")
		}
	}
	if len(events) > 0 && contextSnapshot.EvidenceCoverage == "current" {
		contextSnapshot.EvidenceCoverage = "partial"
		for _, event := range events {
			if event.EvidenceStatus == "supported" {
				contextSnapshot.EvidenceCoverage = "supported"
				break
			}
		}
	}
	if contextSnapshot.RiskState == "normal" && contextSnapshot.MarketDataFresh {
		reasons = append(reasons, "market_data_fresh")
	}
	sort.Slice(eventIDs, func(i, j int) bool { return eventIDs[i] < eventIDs[j] })
	contextSnapshot.KnownEventIDs, _ = json.Marshal(eventIDs)
	contextSnapshot.ReasonCodes, _ = json.Marshal(reasons)
	contextSnapshot.Warnings, _ = json.Marshal(warnings)
	contextSnapshot.Payload, _ = json.Marshal(map[string]any{
		"schemaVersion":      "event_context_v1",
		"eventCount":         len(events),
		"marketDataFresh":    contextSnapshot.MarketDataFresh,
		"evidenceCoverage":   contextSnapshot.EvidenceCoverage,
		"riskState":          contextSnapshot.RiskState,
		"lookAheadProtected": true,
	})
	contextSnapshot.CreatedAt = time.Now().UTC()
	return contextSnapshot, nil
}

func requestedSymbol(raw json.RawMessage, symbol string) bool {
	var symbols []string
	if json.Unmarshal(raw, &symbols) != nil {
		return false
	}
	for _, candidate := range symbols {
		if candidate == symbol {
			return true
		}
	}
	return false
}
