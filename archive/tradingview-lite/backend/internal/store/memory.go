package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"tradingview-lite/backend/internal/alert"
	"tradingview-lite/backend/internal/market"
	"tradingview-lite/backend/internal/options"
)

type MemoryStore struct {
	mu               sync.RWMutex
	watchlist        map[string]WatchlistItem
	bars             map[string]map[time.Time]market.Bar
	quotes           map[string]market.Quote
	quoteSnapshots   []QuoteSnapshot
	rwaSnapshots     []RWASnapshot
	rwaBars          map[string]map[time.Time]RWABarSnapshot
	optionContracts  []OptionContractSnapshot
	ivSnapshots      []IVSnapshot
	featureSnapshots []FeatureSnapshot
	eventRuns        []EventIngestionRun
	eventEvidence    []EventEvidence
	marketEvents     []MarketEvent
	eventContexts    []EventContextSnapshot
	shadowSignals    []ShadowSignalObservation
	symbolProfiles   map[string]SymbolProfile
	peerMaps         []PeerMapVersion
	policyCandidates []PolicyCandidateAlert
	policyOutcomes   map[string]PolicyCandidateOutcome
	policyFeedback   map[int64]PolicyCandidateFeedback
	collectionRuns   []CollectionRun
	dailySummaries   map[string]DailyMarketSummary
	nextSnapshotID   int64
	evidenceByURL    map[string]int
	eventByKey       map[string]int
	alerts           map[string]alert.AlertConfig
	events           []alert.Event
	deliveries       []DiscordDelivery
}

func NewMemoryStore(defaultSymbols []string) *MemoryStore {
	s := &MemoryStore{
		watchlist:      map[string]WatchlistItem{},
		bars:           map[string]map[time.Time]market.Bar{},
		quotes:         map[string]market.Quote{},
		rwaBars:        map[string]map[time.Time]RWABarSnapshot{},
		dailySummaries: map[string]DailyMarketSummary{},
		evidenceByURL:  map[string]int{},
		eventByKey:     map[string]int{},
		symbolProfiles: map[string]SymbolProfile{},
		policyOutcomes: map[string]PolicyCandidateOutcome{},
		policyFeedback: map[int64]PolicyCandidateFeedback{},
		alerts:         map[string]alert.AlertConfig{},
		events:         []alert.Event{},
	}
	for _, symbol := range defaultSymbols {
		_ = s.AddWatchlist(context.Background(), WatchlistItem{
			Symbol: normalizeSymbol(symbol),
			Group:  defaultGroup(symbol),
		})
	}
	return s
}

func (s *MemoryStore) ListWatchlist(ctx context.Context) ([]WatchlistItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]WatchlistItem, 0, len(s.watchlist))
	for _, item := range s.watchlist {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Group == items[j].Group {
			return items[i].Symbol < items[j].Symbol
		}
		return items[i].Group < items[j].Group
	})
	return items, nil
}

func (s *MemoryStore) AddWatchlist(ctx context.Context, item WatchlistItem) error {
	symbol := normalizeSymbol(item.Symbol)
	if symbol == "" {
		return errors.New("symbol is required")
	}
	if item.Group == "" {
		item.Group = "custom"
	}
	item.Symbol = symbol

	s.mu.Lock()
	defer s.mu.Unlock()
	s.watchlist[symbol] = item
	return nil
}

func (s *MemoryStore) DeleteWatchlist(ctx context.Context, symbol string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.watchlist, normalizeSymbol(symbol))
	return nil
}

func (s *MemoryStore) UpsertBars(ctx context.Context, symbol, timeframe, provider string, bars []market.Bar) (int, error) {
	key := barKey(symbol, timeframe)
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.bars[key]; !ok {
		s.bars[key] = map[time.Time]market.Bar{}
	}
	changed := 0
	for _, bar := range bars {
		if bar.Time.IsZero() {
			continue
		}
		if _, exists := s.bars[key][bar.Time]; !exists {
			changed++
		}
		s.bars[key][bar.Time] = bar
	}
	return changed, nil
}

func (s *MemoryStore) GetBars(ctx context.Context, symbol, timeframe string, limit int) ([]market.Bar, error) {
	key := barKey(symbol, timeframe)
	s.mu.RLock()
	defer s.mu.RUnlock()

	byTime := s.bars[key]
	if len(byTime) == 0 {
		return []market.Bar{}, nil
	}
	bars := make([]market.Bar, 0, len(byTime))
	for _, bar := range byTime {
		bars = append(bars, bar)
	}
	sort.Slice(bars, func(i, j int) bool {
		return bars[i].Time.Before(bars[j].Time)
	})
	if limit > 0 && len(bars) > limit {
		return bars[len(bars)-limit:], nil
	}
	return bars, nil
}

func (s *MemoryStore) UpsertQuote(ctx context.Context, quote market.Quote) error {
	symbol := normalizeSymbol(quote.Symbol)
	if symbol == "" {
		return errors.New("quote symbol is required")
	}
	quote.Symbol = symbol

	s.mu.Lock()
	defer s.mu.Unlock()
	s.quotes[symbol] = quote
	return nil
}

func (s *MemoryStore) GetQuote(ctx context.Context, symbol string) (*market.Quote, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	quote, ok := s.quotes[normalizeSymbol(symbol)]
	if !ok {
		return nil, nil
	}
	return &quote, nil
}

func (s *MemoryStore) ListQuotes(ctx context.Context) ([]market.Quote, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	quotes := make([]market.Quote, 0, len(s.quotes))
	for _, quote := range s.quotes {
		quotes = append(quotes, quote)
	}
	sort.Slice(quotes, func(i, j int) bool {
		return quotes[i].Symbol < quotes[j].Symbol
	})
	return quotes, nil
}

func (s *MemoryStore) AppendQuoteSnapshot(ctx context.Context, snapshot QuoteSnapshot) error {
	snapshot.Symbol = normalizeSymbol(snapshot.Symbol)
	if snapshot.Symbol == "" || snapshot.Provider == "" || snapshot.Price <= 0 {
		return errors.New("quote snapshot requires symbol, provider, and positive price")
	}
	if snapshot.ReceivedAt.IsZero() {
		snapshot.ReceivedAt = time.Now().UTC()
	}
	if snapshot.ProviderTime.IsZero() {
		snapshot.ProviderTime = snapshot.ReceivedAt
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextSnapshotID++
	snapshot.ID = s.nextSnapshotID
	s.quoteSnapshots = append(s.quoteSnapshots, snapshot)
	return nil
}

func (s *MemoryStore) ListQuoteSnapshots(ctx context.Context, symbol, provider string, bounds TimeRange) ([]QuoteSnapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := make([]QuoteSnapshot, 0)
	for _, row := range s.quoteSnapshots {
		if !matchesRange(row.Symbol, row.Provider, row.ReceivedAt, symbol, provider, bounds) {
			continue
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].ReceivedAt.Before(rows[j].ReceivedAt) })
	return limitRows(rows, bounds.Limit), nil
}

func (s *MemoryStore) AppendRWASnapshot(ctx context.Context, snapshot RWASnapshot) error {
	snapshot.Symbol = normalizeSymbol(snapshot.Symbol)
	if snapshot.Symbol == "" || snapshot.Provider == "" || snapshot.Ticker == "" || snapshot.Price <= 0 {
		return errors.New("RWA snapshot requires symbol, ticker, provider, and positive price")
	}
	if snapshot.ReceivedAt.IsZero() {
		snapshot.ReceivedAt = time.Now().UTC()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextSnapshotID++
	snapshot.ID = s.nextSnapshotID
	s.rwaSnapshots = append(s.rwaSnapshots, snapshot)
	return nil
}

func (s *MemoryStore) ListRWASnapshots(ctx context.Context, symbol, provider string, bounds TimeRange) ([]RWASnapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := make([]RWASnapshot, 0)
	for _, row := range s.rwaSnapshots {
		if !matchesRange(row.Symbol, row.Provider, row.ReceivedAt, symbol, provider, bounds) {
			continue
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].ReceivedAt.Before(rows[j].ReceivedAt) })
	return limitRows(rows, bounds.Limit), nil
}

func (s *MemoryStore) UpsertRWABars(ctx context.Context, rows []RWABarSnapshot) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	changed := 0
	for _, row := range rows {
		row.Symbol = normalizeSymbol(row.Symbol)
		if row.Symbol == "" || row.Provider == "" || row.Timeframe == "" || row.Timestamp.IsZero() {
			continue
		}
		key := rwaBarKey(row.Symbol, row.Provider, row.Timeframe)
		if s.rwaBars[key] == nil {
			s.rwaBars[key] = map[time.Time]RWABarSnapshot{}
		}
		if _, exists := s.rwaBars[key][row.Timestamp]; !exists {
			changed++
		}
		s.rwaBars[key][row.Timestamp] = row
	}
	return changed, nil
}

func (s *MemoryStore) GetRWABars(ctx context.Context, symbol, timeframe string, limit int) ([]RWABarSnapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := make([]RWABarSnapshot, 0)
	for key, byTime := range s.rwaBars {
		prefix := normalizeSymbol(symbol) + "|"
		suffix := "|" + strings.ToLower(strings.TrimSpace(timeframe))
		if !strings.HasPrefix(key, prefix) || !strings.HasSuffix(key, suffix) {
			continue
		}
		for _, row := range byTime {
			rows = append(rows, row)
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Timestamp.Before(rows[j].Timestamp) })
	return limitRows(rows, limit), nil
}

func (s *MemoryStore) SaveOptionChain(ctx context.Context, chain *options.ChainResponse) (int, error) {
	if chain == nil {
		return 0, errors.New("option chain is required")
	}
	contracts := OptionContracts(chain)
	iv := IVFromChain(chain)
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range contracts {
		s.nextSnapshotID++
		contracts[i].ID = s.nextSnapshotID
		s.optionContracts = append(s.optionContracts, contracts[i])
	}
	if iv != nil {
		s.nextSnapshotID++
		iv.ID = s.nextSnapshotID
		s.ivSnapshots = append(s.ivSnapshots, *iv)
	}
	return len(contracts), nil
}

func (s *MemoryStore) ListOptionContracts(ctx context.Context, symbol string, bounds TimeRange) ([]OptionContractSnapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := make([]OptionContractSnapshot, 0)
	for _, row := range s.optionContracts {
		if matchesRange(row.Symbol, "", row.ReceivedAt, symbol, "", bounds) {
			rows = append(rows, row)
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].ReceivedAt.Before(rows[j].ReceivedAt) })
	return limitRows(rows, bounds.Limit), nil
}

func (s *MemoryStore) ListIVSnapshots(ctx context.Context, symbol string, bounds TimeRange) ([]IVSnapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := make([]IVSnapshot, 0)
	for _, row := range s.ivSnapshots {
		if matchesRange(row.Symbol, "", row.ReceivedAt, symbol, "", bounds) {
			rows = append(rows, row)
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].ReceivedAt.Before(rows[j].ReceivedAt) })
	return limitRows(rows, bounds.Limit), nil
}

func (s *MemoryStore) AppendFeatureSnapshot(ctx context.Context, snapshot FeatureSnapshot) error {
	snapshot.Symbol = normalizeSymbol(snapshot.Symbol)
	if snapshot.Symbol == "" {
		return errors.New("feature snapshot requires symbol")
	}
	if snapshot.ComputedAt.IsZero() {
		snapshot.ComputedAt = time.Now().UTC()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextSnapshotID++
	snapshot.ID = s.nextSnapshotID
	s.featureSnapshots = append(s.featureSnapshots, snapshot)
	return nil
}

func (s *MemoryStore) ListFeatureSnapshots(ctx context.Context, symbol string, bounds TimeRange) ([]FeatureSnapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := make([]FeatureSnapshot, 0)
	for _, row := range s.featureSnapshots {
		if matchesRange(row.Symbol, "", row.ComputedAt, symbol, "", bounds) {
			rows = append(rows, row)
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].ComputedAt.Before(rows[j].ComputedAt) })
	return limitRows(rows, bounds.Limit), nil
}

func (s *MemoryStore) UpsertDailySummary(ctx context.Context, summary DailyMarketSummary) error {
	summary.Symbol = normalizeSymbol(summary.Symbol)
	if summary.Symbol == "" || summary.Provider == "" || summary.TradingDate.IsZero() {
		return errors.New("daily summary requires symbol, provider, and trading date")
	}
	if summary.UpdatedAt.IsZero() {
		summary.UpdatedAt = time.Now().UTC()
	}
	key := dailySummaryKey(summary.Symbol, summary.Provider, summary.TradingDate)
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, ok := s.dailySummaries[key]; ok {
		summary = mergeDailySummary(existing, summary)
	}
	s.dailySummaries[key] = summary
	return nil
}

func (s *MemoryStore) GetDailySummary(ctx context.Context, symbol string, tradingDate time.Time) (*DailyMarketSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tradingDate = dateOnlyUTC(tradingDate)
	for _, summary := range s.dailySummaries {
		if summary.Symbol == normalizeSymbol(symbol) && summary.TradingDate.Equal(tradingDate) {
			copy := summary
			return &copy, nil
		}
	}
	return nil, nil
}

func (s *MemoryStore) ListDailySummaries(ctx context.Context, symbol string, bounds TimeRange) ([]DailyMarketSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := make([]DailyMarketSummary, 0)
	for _, summary := range s.dailySummaries {
		if symbol != "" && summary.Symbol != normalizeSymbol(symbol) {
			continue
		}
		if !matchesTime(summary.TradingDate, bounds) {
			continue
		}
		rows = append(rows, summary)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].TradingDate.Before(rows[j].TradingDate) })
	return limitRows(rows, bounds.Limit), nil
}

func (s *MemoryStore) AddCollectionRun(ctx context.Context, run CollectionRun) error {
	if run.Collector == "" || run.Provider == "" {
		return errors.New("collection run requires collector and provider")
	}
	if run.StartedAt.IsZero() {
		run.StartedAt = time.Now().UTC()
	}
	if run.FinishedAt.IsZero() {
		run.FinishedAt = run.StartedAt
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextSnapshotID++
	run.ID = s.nextSnapshotID
	s.collectionRuns = append(s.collectionRuns, run)
	return nil
}

func (s *MemoryStore) ListCollectionRuns(ctx context.Context, limit int) ([]CollectionRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := append([]CollectionRun(nil), s.collectionRuns...)
	sort.Slice(rows, func(i, j int) bool { return rows[i].FinishedAt.After(rows[j].FinishedAt) })
	if limit > 0 && len(rows) > limit {
		return rows[:limit], nil
	}
	return rows, nil
}

func (s *MemoryStore) AddEventIngestionRun(ctx context.Context, run EventIngestionRun) (EventIngestionRun, error) {
	if run.Provider == "" || run.Query == "" || run.Scope == "" {
		return run, errors.New("event ingestion run requires provider, query, and scope")
	}
	if run.StartedAt.IsZero() {
		run.StartedAt = time.Now().UTC()
	}
	if run.FinishedAt.IsZero() {
		run.FinishedAt = run.StartedAt
	}
	run.RequestedSymbols = defaultJSON(run.RequestedSymbols, []byte("[]"))
	run.Metadata = defaultJSON(run.Metadata, []byte("{}"))
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextSnapshotID++
	run.ID = s.nextSnapshotID
	s.eventRuns = append(s.eventRuns, run)
	return run, nil
}

func (s *MemoryStore) ListEventIngestionRuns(ctx context.Context, limit int) ([]EventIngestionRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := append([]EventIngestionRun(nil), s.eventRuns...)
	sort.Slice(rows, func(i, j int) bool { return rows[i].FinishedAt.After(rows[j].FinishedAt) })
	if limit > 0 && len(rows) > limit {
		rows = rows[:limit]
	}
	return rows, nil
}

func (s *MemoryStore) SaveEventEvidence(ctx context.Context, runID int64, rows []EventEvidence) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	written := 0
	for _, row := range rows {
		if row.CanonicalURL == "" || row.Title == "" || row.Provider == "" || row.Scope == "" || row.EventType == "" {
			return written, errors.New("event evidence requires URL, title, provider, scope, and event type")
		}
		if _, exists := s.evidenceByURL[row.CanonicalURL]; exists {
			continue
		}
		if row.ReceivedAt.IsZero() {
			row.ReceivedAt = time.Now().UTC()
		}
		row.IngestionRunID = runID
		row.MatchedSymbols = defaultJSON(row.MatchedSymbols, []byte("[]"))
		row.Metadata = defaultJSON(row.Metadata, []byte("{}"))
		s.nextSnapshotID++
		row.ID = s.nextSnapshotID

		key, primarySymbol := marketEventKey(row)
		index, exists := s.eventByKey[key]
		if !exists {
			s.nextSnapshotID++
			event := MarketEvent{
				ID:             s.nextSnapshotID,
				EventKey:       key,
				Scope:          row.Scope,
				PrimarySymbol:  primarySymbol,
				GroupName:      row.GroupName,
				EventType:      row.EventType,
				EvidenceStatus: evidenceStatus(row.SourceTier),
				Severity:       "unknown",
				State:          "active",
				FirstSeenAt:    row.ReceivedAt,
				LastSeenAt:     row.ReceivedAt,
			}
			s.marketEvents = append(s.marketEvents, event)
			index = len(s.marketEvents) - 1
			s.eventByKey[key] = index
		} else {
			event := &s.marketEvents[index]
			event.LastSeenAt = row.ReceivedAt
			if evidenceStatus(row.SourceTier) == "supported" {
				event.EvidenceStatus = "supported"
			}
		}
		row.EventID = s.marketEvents[index].ID
		s.evidenceByURL[row.CanonicalURL] = len(s.eventEvidence)
		s.eventEvidence = append(s.eventEvidence, row)
		written++
	}
	return written, nil
}

func (s *MemoryStore) ListEventEvidence(ctx context.Context, filter EventFilter) ([]EventEvidence, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := make([]EventEvidence, 0)
	for _, row := range s.eventEvidence {
		if !matchesTime(row.ReceivedAt, filter.Bounds) {
			continue
		}
		if filter.Symbol != "" && !jsonContainsString(row.MatchedSymbols, normalizeSymbol(filter.Symbol)) {
			continue
		}
		if filter.Group != "" && row.GroupName != filter.Group {
			continue
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].ReceivedAt.Before(rows[j].ReceivedAt) })
	return limitRows(rows, filter.Bounds.Limit), nil
}

func (s *MemoryStore) ListMarketEvents(ctx context.Context, filter EventFilter) ([]MarketEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := make([]MarketEvent, 0)
	for _, row := range s.marketEvents {
		if !matchesTime(row.LastSeenAt, filter.Bounds) {
			continue
		}
		if filter.Symbol != "" && row.PrimarySymbol != normalizeSymbol(filter.Symbol) && row.Scope != "macro" && row.Scope != "market" {
			continue
		}
		if filter.Group != "" && row.GroupName != "" && row.GroupName != filter.Group && row.Scope == "group" {
			continue
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].LastSeenAt.Before(rows[j].LastSeenAt) })
	return limitRows(rows, filter.Bounds.Limit), nil
}

func (s *MemoryStore) AppendEventContext(ctx context.Context, snapshot EventContextSnapshot) (EventContextSnapshot, error) {
	snapshot.Symbol = normalizeSymbol(snapshot.Symbol)
	if snapshot.Symbol == "" || snapshot.AsOf.IsZero() || snapshot.EvidenceCoverage == "" || snapshot.RiskState == "" {
		return snapshot, errors.New("event context requires symbol, asOf, coverage, and risk state")
	}
	if snapshot.CreatedAt.IsZero() {
		snapshot.CreatedAt = time.Now().UTC()
	}
	snapshot.KnownEventIDs = defaultJSON(snapshot.KnownEventIDs, []byte("[]"))
	snapshot.ReasonCodes = defaultJSON(snapshot.ReasonCodes, []byte("[]"))
	snapshot.Warnings = defaultJSON(snapshot.Warnings, []byte("[]"))
	snapshot.Payload = defaultJSON(snapshot.Payload, []byte("{}"))
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextSnapshotID++
	snapshot.ID = s.nextSnapshotID
	s.eventContexts = append(s.eventContexts, snapshot)
	return snapshot, nil
}

func (s *MemoryStore) GetEventContext(ctx context.Context, symbol string, asOf time.Time) (*EventContextSnapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var found *EventContextSnapshot
	for i := range s.eventContexts {
		row := s.eventContexts[i]
		if row.Symbol != normalizeSymbol(symbol) || (!asOf.IsZero() && row.AsOf.After(asOf)) {
			continue
		}
		if found == nil || row.AsOf.After(found.AsOf) {
			copy := row
			found = &copy
		}
	}
	return found, nil
}

func (s *MemoryStore) AppendShadowSignal(ctx context.Context, observation ShadowSignalObservation) (ShadowSignalObservation, error) {
	observation.Symbol = normalizeSymbol(observation.Symbol)
	if observation.SignalType == "" || observation.Symbol == "" || observation.AsOf.IsZero() || observation.State == "" || observation.ThresholdVersion == "" {
		return observation, errors.New("shadow signal requires type, symbol, asOf, state, and threshold version")
	}
	if observation.CreatedAt.IsZero() {
		observation.CreatedAt = time.Now().UTC()
	}
	observation.ReasonCodes = defaultJSON(observation.ReasonCodes, []byte("[]"))
	observation.Metrics = defaultJSON(observation.Metrics, []byte("{}"))
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextSnapshotID++
	observation.ID = s.nextSnapshotID
	s.shadowSignals = append(s.shadowSignals, observation)
	return observation, nil
}

func (s *MemoryStore) ListShadowSignals(ctx context.Context, symbol string, bounds TimeRange) ([]ShadowSignalObservation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := make([]ShadowSignalObservation, 0)
	for _, row := range s.shadowSignals {
		if row.Symbol == normalizeSymbol(symbol) && matchesTime(row.AsOf, bounds) {
			rows = append(rows, row)
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].AsOf.Before(rows[j].AsOf) })
	return limitRows(rows, bounds.Limit), nil
}

func (s *MemoryStore) ListSymbolProfiles(ctx context.Context) ([]SymbolProfile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := make([]SymbolProfile, 0, len(s.symbolProfiles))
	for _, profile := range s.symbolProfiles {
		rows = append(rows, cloneProfile(profile))
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Symbol < rows[j].Symbol })
	return rows, nil
}

func (s *MemoryStore) GetSymbolProfile(ctx context.Context, symbol string) (*SymbolProfile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	profile, ok := s.symbolProfiles[normalizeSymbol(symbol)]
	if !ok {
		return nil, nil
	}
	copy := cloneProfile(profile)
	return &copy, nil
}

func (s *MemoryStore) UpsertSymbolProfile(ctx context.Context, profile SymbolProfile) (SymbolProfile, error) {
	profile, err := normalizeSymbolProfile(profile)
	if err != nil {
		return profile, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.symbolProfiles[profile.Symbol] = cloneProfile(profile)
	return profile, nil
}

func (s *MemoryStore) CreatePeerMapVersion(ctx context.Context, peerMap PeerMapVersion) (PeerMapVersion, error) {
	peerMap, err := normalizePeerMapVersion(peerMap)
	if err != nil {
		return peerMap, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.peerMaps {
		if existing.MapKey == peerMap.MapKey && existing.Version >= peerMap.Version {
			peerMap.Version = existing.Version + 1
		}
	}
	s.nextSnapshotID++
	peerMap.ID = s.nextSnapshotID
	s.peerMaps = append(s.peerMaps, clonePeerMap(peerMap))
	return peerMap, nil
}

func (s *MemoryStore) ListPeerMapVersions(ctx context.Context, mapKey string) ([]PeerMapVersion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := make([]PeerMapVersion, 0)
	for _, peerMap := range s.peerMaps {
		if mapKey != "" && peerMap.MapKey != normalizeMapKey(mapKey) {
			continue
		}
		rows = append(rows, clonePeerMap(peerMap))
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].MapKey == rows[j].MapKey {
			return rows[i].Version < rows[j].Version
		}
		return rows[i].MapKey < rows[j].MapKey
	})
	return rows, nil
}

func (s *MemoryStore) GetActivePeerMap(ctx context.Context, symbol, group string, asOf time.Time) (*PeerMapVersion, error) {
	if asOf.IsZero() {
		asOf = time.Now().UTC()
	}
	symbol = normalizeSymbol(symbol)
	group = strings.TrimSpace(group)
	s.mu.RLock()
	defer s.mu.RUnlock()
	var selected *PeerMapVersion
	selectedSpecific := false
	for _, peerMap := range s.peerMaps {
		if peerMap.ReviewStatus != "approved" || peerMap.EffectiveFrom.After(asOf) {
			continue
		}
		specific := peerMapIncludes(peerMap, symbol)
		if !specific && peerMap.GroupName != group {
			continue
		}
		if selected == nil || (specific && !selectedSpecific) || (specific == selectedSpecific && peerMap.EffectiveFrom.After(selected.EffectiveFrom)) || (specific == selectedSpecific && peerMap.EffectiveFrom.Equal(selected.EffectiveFrom) && peerMap.Version > selected.Version) {
			copy := clonePeerMap(peerMap)
			selected = &copy
			selectedSpecific = specific
		}
	}
	return selected, nil
}

func (s *MemoryStore) AppendPolicyCandidate(ctx context.Context, candidate PolicyCandidateAlert) (PolicyCandidateAlert, error) {
	candidate, err := normalizePolicyCandidate(candidate)
	if err != nil {
		return candidate, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextSnapshotID++
	candidate.ID = s.nextSnapshotID
	s.policyCandidates = append(s.policyCandidates, candidate)
	return candidate, nil
}

func (s *MemoryStore) GetPolicyCandidate(ctx context.Context, id int64) (*PolicyCandidateAlert, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, candidate := range s.policyCandidates {
		if candidate.ID == id {
			copy := clonePolicyCandidate(candidate)
			return &copy, nil
		}
	}
	return nil, nil
}

func (s *MemoryStore) ListPolicyCandidates(ctx context.Context, filter PolicyCandidateFilter) ([]PolicyCandidateAlert, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := make([]PolicyCandidateAlert, 0)
	for _, candidate := range s.policyCandidates {
		if filter.Symbol != "" && candidate.Symbol != normalizeSymbol(filter.Symbol) {
			continue
		}
		if !filter.TradingDate.IsZero() && !candidate.TradingDate.Equal(dateOnlyUTC(filter.TradingDate)) {
			continue
		}
		if filter.OnlyWouldNotify && !candidate.WouldNotify {
			continue
		}
		if !matchesTime(candidate.AsOf, filter.Bounds) {
			continue
		}
		rows = append(rows, clonePolicyCandidate(candidate))
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].AsOf.Before(rows[j].AsOf) })
	return limitRows(rows, filter.Bounds.Limit), nil
}

func (s *MemoryStore) UpsertPolicyCandidateOutcome(ctx context.Context, outcome PolicyCandidateOutcome) (PolicyCandidateOutcome, error) {
	if outcome.CandidateAlertID == 0 || outcome.HorizonMinutes <= 0 || outcome.TargetAt.IsZero() || outcome.ObservedAt.IsZero() || outcome.SampleCount <= 0 {
		return outcome, errors.New("policy outcome requires candidate, horizon, target, observation, and samples")
	}
	if outcome.EvaluatedAt.IsZero() {
		outcome.EvaluatedAt = time.Now().UTC()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.hasPolicyCandidate(outcome.CandidateAlertID) {
		return outcome, errors.New("policy outcome candidate does not exist")
	}
	key := policyOutcomeKey(outcome.CandidateAlertID, outcome.HorizonMinutes)
	if existing, ok := s.policyOutcomes[key]; ok {
		outcome.ID = existing.ID
	} else {
		s.nextSnapshotID++
		outcome.ID = s.nextSnapshotID
	}
	s.policyOutcomes[key] = outcome
	return outcome, nil
}

func (s *MemoryStore) ListPolicyCandidateOutcomes(ctx context.Context, candidateID int64) ([]PolicyCandidateOutcome, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := make([]PolicyCandidateOutcome, 0)
	for _, outcome := range s.policyOutcomes {
		if outcome.CandidateAlertID == candidateID {
			rows = append(rows, outcome)
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].HorizonMinutes < rows[j].HorizonMinutes })
	return rows, nil
}

func (s *MemoryStore) UpsertPolicyCandidateFeedback(ctx context.Context, feedback PolicyCandidateFeedback) (PolicyCandidateFeedback, error) {
	if feedback.CandidateAlertID == 0 {
		return feedback, errors.New("policy feedback requires a candidate")
	}
	if feedback.EmotionIntensity != nil && (*feedback.EmotionIntensity < 0 || *feedback.EmotionIntensity > 5) {
		return feedback, errors.New("emotion intensity must be between 0 and 5")
	}
	if feedback.UpdatedAt.IsZero() {
		feedback.UpdatedAt = time.Now().UTC()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.hasPolicyCandidate(feedback.CandidateAlertID) {
		return feedback, errors.New("policy feedback candidate does not exist")
	}
	s.policyFeedback[feedback.CandidateAlertID] = feedback
	return feedback, nil
}

func (s *MemoryStore) GetPolicyCandidateFeedback(ctx context.Context, candidateID int64) (*PolicyCandidateFeedback, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	feedback, ok := s.policyFeedback[candidateID]
	if !ok {
		return nil, nil
	}
	copy := feedback
	return &copy, nil
}

func (s *MemoryStore) PruneSnapshots(ctx context.Context, policy RetentionPolicy, now time.Time) (RetentionResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := RetentionResult{}
	s.quoteSnapshots, result.QuoteSnapshots = pruneByTime(s.quoteSnapshots, policy.QuoteSnapshots, now, func(row QuoteSnapshot) time.Time { return row.ReceivedAt })
	s.rwaSnapshots, result.RWASnapshots = pruneByTime(s.rwaSnapshots, policy.RWASnapshots, now, func(row RWASnapshot) time.Time { return row.ReceivedAt })
	s.optionContracts, result.OptionSnapshots = pruneByTime(s.optionContracts, policy.OptionSnapshots, now, func(row OptionContractSnapshot) time.Time { return row.ReceivedAt })
	s.ivSnapshots, result.IVSnapshots = pruneByTime(s.ivSnapshots, policy.OptionSnapshots, now, func(row IVSnapshot) time.Time { return row.ReceivedAt })
	s.featureSnapshots, result.FeatureSnapshots = pruneByTime(s.featureSnapshots, policy.FeatureSnapshots, now, func(row FeatureSnapshot) time.Time { return row.ComputedAt })
	s.collectionRuns, result.CollectionRuns = pruneByTime(s.collectionRuns, policy.CollectionRuns, now, func(row CollectionRun) time.Time { return row.FinishedAt })
	s.eventEvidence, result.EventEvidence = pruneByTime(s.eventEvidence, policy.EventEvidence, now, func(row EventEvidence) time.Time { return row.ReceivedAt })
	s.evidenceByURL = map[string]int{}
	for index, evidence := range s.eventEvidence {
		s.evidenceByURL[evidence.CanonicalURL] = index
	}
	if policy.EventEvidence > 0 {
		s.marketEvents, _ = pruneByTime(s.marketEvents, policy.EventEvidence, now, func(row MarketEvent) time.Time { return row.LastSeenAt })
		s.eventByKey = map[string]int{}
		for index, event := range s.marketEvents {
			s.eventByKey[event.EventKey] = index
		}
	}
	s.eventContexts, result.EventContexts = pruneByTime(s.eventContexts, policy.EventContexts, now, func(row EventContextSnapshot) time.Time { return row.AsOf })
	s.shadowSignals, result.ShadowSignals = pruneByTime(s.shadowSignals, policy.ShadowSignals, now, func(row ShadowSignalObservation) time.Time { return row.AsOf })
	s.eventRuns, result.EventRuns = pruneByTime(s.eventRuns, policy.EventRuns, now, func(row EventIngestionRun) time.Time { return row.FinishedAt })
	s.policyCandidates, result.PolicyCandidates = pruneByTime(s.policyCandidates, policy.PolicyCandidates, now, func(row PolicyCandidateAlert) time.Time { return row.AsOf })
	remainingCandidates := map[int64]bool{}
	for _, candidate := range s.policyCandidates {
		remainingCandidates[candidate.ID] = true
	}
	for key, outcome := range s.policyOutcomes {
		if !remainingCandidates[outcome.CandidateAlertID] || (policy.PolicyOutcomes > 0 && outcome.EvaluatedAt.Before(now.Add(-policy.PolicyOutcomes))) {
			delete(s.policyOutcomes, key)
			result.PolicyOutcomes++
		}
	}
	for candidateID := range s.policyFeedback {
		if !remainingCandidates[candidateID] {
			delete(s.policyFeedback, candidateID)
		}
	}
	return result, nil
}

func (s *MemoryStore) StorageDiagnostics(ctx context.Context) (StorageDiagnostic, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return StorageDiagnostic{
		Store:      "memory",
		TableBytes: map[string]int64{},
		RowEstimates: map[string]int64{
			"market_quote_snapshots":     int64(len(s.quoteSnapshots)),
			"rwa_quote_snapshots":        int64(len(s.rwaSnapshots)),
			"option_contract_snapshots":  int64(len(s.optionContracts)),
			"iv_snapshots":               int64(len(s.ivSnapshots)),
			"intraday_feature_snapshots": int64(len(s.featureSnapshots)),
			"collection_runs":            int64(len(s.collectionRuns)),
			"event_ingestion_runs":       int64(len(s.eventRuns)),
			"event_evidence":             int64(len(s.eventEvidence)),
			"market_events":              int64(len(s.marketEvents)),
			"event_context_snapshots":    int64(len(s.eventContexts)),
			"shadow_signal_observations": int64(len(s.shadowSignals)),
			"symbol_profiles":            int64(len(s.symbolProfiles)),
			"peer_map_versions":          int64(len(s.peerMaps)),
			"policy_candidate_alerts":    int64(len(s.policyCandidates)),
			"policy_candidate_outcomes":  int64(len(s.policyOutcomes)),
			"policy_candidate_feedback":  int64(len(s.policyFeedback)),
		},
		CollectedAt: time.Now().UTC(),
	}, nil
}

func (s *MemoryStore) CreateAlert(ctx context.Context, cfg alert.AlertConfig) (alert.AlertConfig, error) {
	cfg.Symbol = normalizeSymbol(cfg.Symbol)
	if cfg.Symbol == "" {
		return cfg, errors.New("alert symbol is required")
	}
	if len(cfg.Rules) == 0 {
		return cfg, errors.New("alert requires at least one rule")
	}
	if cfg.ID == "" {
		cfg.ID = fmt.Sprintf("al_%d", time.Now().UTC().UnixNano())
	}
	if cfg.CreatedAt.IsZero() {
		cfg.CreatedAt = time.Now().UTC()
	}
	cfg.Enabled = true

	s.mu.Lock()
	defer s.mu.Unlock()
	s.alerts[cfg.ID] = cfg
	return cfg, nil
}

func (s *MemoryStore) ListAlerts(ctx context.Context) ([]alert.AlertConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	alerts := make([]alert.AlertConfig, 0, len(s.alerts))
	for _, cfg := range s.alerts {
		alerts = append(alerts, cfg)
	}
	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].CreatedAt.Before(alerts[j].CreatedAt)
	})
	return alerts, nil
}

func (s *MemoryStore) AddEvent(ctx context.Context, event alert.Event) error {
	if event.ID == "" {
		event.ID = fmt.Sprintf("ev_%d", time.Now().UTC().UnixNano())
	}
	if event.TriggeredAt.IsZero() {
		event.TriggeredAt = time.Now().UTC()
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, event)
	return nil
}

func (s *MemoryStore) ListEvents(ctx context.Context, limit int) ([]alert.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events := make([]alert.Event, len(s.events))
	copy(events, s.events)
	sort.Slice(events, func(i, j int) bool {
		return events[i].TriggeredAt.After(events[j].TriggeredAt)
	})
	if limit > 0 && len(events) > limit {
		return events[:limit], nil
	}
	return events, nil
}

func (s *MemoryStore) EnqueueDiscordDelivery(ctx context.Context, delivery DiscordDelivery) (bool, error) {
	if delivery.IdempotencyKey == "" {
		return false, errors.New("discord delivery idempotency key is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, row := range s.deliveries {
		if row.IdempotencyKey == delivery.IdempotencyKey {
			return false, nil
		}
	}
	s.nextSnapshotID++
	delivery.ID = s.nextSnapshotID
	if delivery.Status == "" {
		delivery.Status = "pending"
	}
	if delivery.NextAttemptAt.IsZero() {
		delivery.NextAttemptAt = time.Now().UTC()
	}
	s.deliveries = append(s.deliveries, delivery)
	return true, nil
}

func (s *MemoryStore) ClaimDiscordDeliveries(ctx context.Context, now time.Time, limit int) ([]DiscordDelivery, error) {
	if limit <= 0 {
		limit = 20
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	claimed := []DiscordDelivery{}
	for i := range s.deliveries {
		row := &s.deliveries[i]
		if len(claimed) >= limit {
			break
		}
		if row.Status == "pending" && !row.NextAttemptAt.After(now) {
			row.Status = "processing"
			row.Attempts++
			claimed = append(claimed, *row)
		}
	}
	return claimed, nil
}

func (s *MemoryStore) CompleteDiscordDelivery(ctx context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.deliveries {
		if s.deliveries[i].ID == id {
			s.deliveries[i].Status = "delivered"
			return nil
		}
	}
	return errors.New("discord delivery not found")
}

func (s *MemoryStore) RetryDiscordDelivery(ctx context.Context, id int64, next time.Time, message string, dead bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.deliveries {
		if s.deliveries[i].ID == id {
			s.deliveries[i].Status = "pending"
			if dead {
				s.deliveries[i].Status = "dead_letter"
			}
			s.deliveries[i].NextAttemptAt = next
			s.deliveries[i].LastError = message
			return nil
		}
	}
	return errors.New("discord delivery not found")
}

func barKey(symbol, timeframe string) string {
	return normalizeSymbol(symbol) + "|" + strings.ToLower(strings.TrimSpace(timeframe))
}

func normalizeSymbol(symbol string) string {
	return strings.ToUpper(strings.TrimSpace(symbol))
}

func defaultGroup(symbol string) string {
	switch normalizeSymbol(symbol) {
	case "MU", "NVDA", "AMD", "SMH":
		return "semis_memory"
	case "RKLB", "SPCX", "ARKX", "UFO", "LUNR":
		return "space"
	case "TE", "FCEL", "XLE", "TAN", "ICLN":
		return "energy_clean"
	case "IONQ", "RGTI", "QBTS", "QUBT":
		return "quantum"
	case "SPY", "QQQ", "IWM", "TLT", "^VIX":
		return "market_context"
	default:
		return "custom"
	}
}

func matchesRange(rowSymbol, rowProvider string, rowTime time.Time, symbol, provider string, bounds TimeRange) bool {
	if symbol != "" && rowSymbol != normalizeSymbol(symbol) {
		return false
	}
	if provider != "" && rowProvider != provider {
		return false
	}
	if !bounds.From.IsZero() && rowTime.Before(bounds.From) {
		return false
	}
	if !bounds.To.IsZero() && rowTime.After(bounds.To) {
		return false
	}
	return true
}

func limitRows[T any](rows []T, limit int) []T {
	if limit <= 0 || len(rows) <= limit {
		return rows
	}
	return rows[len(rows)-limit:]
}

func defaultJSON(value json.RawMessage, fallback []byte) json.RawMessage {
	if len(value) == 0 {
		return append(json.RawMessage(nil), fallback...)
	}
	return value
}

func jsonContainsString(raw json.RawMessage, target string) bool {
	var values []string
	if json.Unmarshal(raw, &values) != nil {
		return false
	}
	for _, value := range values {
		if normalizeSymbol(value) == target {
			return true
		}
	}
	return false
}

func matchesTime(value time.Time, bounds TimeRange) bool {
	return (bounds.From.IsZero() || !value.Before(bounds.From)) && (bounds.To.IsZero() || !value.After(bounds.To))
}

func marketEventKey(row EventEvidence) (string, string) {
	var symbols []string
	_ = json.Unmarshal(row.MatchedSymbols, &symbols)
	primary := ""
	if len(symbols) > 0 {
		primary = normalizeSymbol(symbols[0])
	}
	return strings.Join([]string{row.Scope, row.GroupName, primary, row.EventType, row.HeadlineHash}, "|"), primary
}

func evidenceStatus(tier string) string {
	if tier == "T1_OFFICIAL" || tier == "T2_PRIMARY" {
		return "supported"
	}
	return "single_source"
}

func pruneByTime[T any](rows []T, retention time.Duration, now time.Time, timestamp func(T) time.Time) ([]T, int64) {
	if retention <= 0 {
		return rows, 0
	}
	cutoff := now.Add(-retention)
	kept := make([]T, 0, len(rows))
	var removed int64
	for _, row := range rows {
		if timestamp(row).Before(cutoff) {
			removed++
			continue
		}
		kept = append(kept, row)
	}
	return kept, removed
}

func rwaBarKey(symbol, provider, timeframe string) string {
	return normalizeSymbol(symbol) + "|" + provider + "|" + strings.ToLower(strings.TrimSpace(timeframe))
}

func dailySummaryKey(symbol, provider string, tradingDate time.Time) string {
	return normalizeSymbol(symbol) + "|" + provider + "|" + dateOnlyUTC(tradingDate).Format("2006-01-02")
}

func dateOnlyUTC(value time.Time) time.Time {
	utc := value.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

func mergeDailySummary(existing, update DailyMarketSummary) DailyMarketSummary {
	if update.PreviousClose == 0 {
		update.PreviousClose = existing.PreviousClose
	}
	update.PremarketHigh = maxNonZero(existing.PremarketHigh, update.PremarketHigh)
	update.PremarketLow = minNonZero(existing.PremarketLow, update.PremarketLow)
	if update.PreOpenFiveMinute == 0 {
		update.PreOpenFiveMinute = existing.PreOpenFiveMinute
	}
	if update.RegularOpen == 0 {
		update.RegularOpen = existing.RegularOpen
	}
	update.DailyHigh = maxNonZero(existing.DailyHigh, update.DailyHigh)
	update.DailyLow = minNonZero(existing.DailyLow, update.DailyLow)
	if update.SelectedMove == 0 {
		update.SelectedMove = existing.SelectedMove
	}
	if update.SelectedMovePct == 0 {
		update.SelectedMovePct = existing.SelectedMovePct
	}
	if update.SelectedSource == "" {
		update.SelectedSource = existing.SelectedSource
	}
	return update
}

func maxNonZero(a, b float64) float64 {
	if a == 0 || b > a {
		return b
	}
	return a
}

func minNonZero(a, b float64) float64 {
	if a == 0 || (b > 0 && b < a) {
		return b
	}
	return a
}

func normalizeSymbolProfile(profile SymbolProfile) (SymbolProfile, error) {
	profile.Symbol = normalizeSymbol(profile.Symbol)
	if profile.Symbol == "" {
		return profile, errors.New("symbol profile requires a symbol")
	}
	if profile.Role == "" {
		profile.Role = "research_watch"
	}
	if profile.Horizon == "" {
		profile.Horizon = "unspecified"
	}
	if !validProfileRole(profile.Role) {
		return profile, errors.New("profile role must be position, tactical_watch, research_watch, or benchmark")
	}
	if !validProfileHorizon(profile.Horizon) {
		return profile, errors.New("profile horizon must be intraday, swing, medium_term, or unspecified")
	}
	if len(profile.ActionPermissions) > 0 && !json.Valid(profile.ActionPermissions) {
		return profile, errors.New("profile action permissions must be valid JSON")
	}
	profile.ActionPermissions = defaultJSON(profile.ActionPermissions, []byte("[]"))
	if profile.UpdatedAt.IsZero() {
		profile.UpdatedAt = time.Now().UTC()
	}
	return profile, nil
}

func normalizePeerMapVersion(peerMap PeerMapVersion) (PeerMapVersion, error) {
	peerMap.MapKey = normalizeMapKey(peerMap.MapKey)
	peerMap.GroupName = strings.TrimSpace(peerMap.GroupName)
	peerMap.BenchmarkSymbol = normalizeSymbol(peerMap.BenchmarkSymbol)
	peerMap.Methodology = strings.TrimSpace(peerMap.Methodology)
	if peerMap.MapKey == "" || peerMap.Methodology == "" {
		return peerMap, errors.New("peer map requires map key and methodology")
	}
	if peerMap.ReviewStatus == "" {
		peerMap.ReviewStatus = "proposed"
	}
	if peerMap.Version <= 0 {
		peerMap.Version = 1
	}
	if peerMap.ReviewStatus != "proposed" && peerMap.ReviewStatus != "approved" && peerMap.ReviewStatus != "retired" {
		return peerMap, errors.New("peer map review status must be proposed, approved, or retired")
	}
	if peerMap.EffectiveFrom.IsZero() {
		peerMap.EffectiveFrom = time.Now().UTC()
	}
	if peerMap.CreatedAt.IsZero() {
		peerMap.CreatedAt = time.Now().UTC()
	}
	seen := map[string]bool{}
	for index := range peerMap.Members {
		member := &peerMap.Members[index]
		member.Symbol = normalizeSymbol(member.Symbol)
		member.Relation = strings.TrimSpace(member.Relation)
		if member.Symbol == "" || member.Relation == "" {
			return peerMap, errors.New("peer map members require symbol and relation")
		}
		if seen[member.Symbol] {
			return peerMap, errors.New("peer map contains a duplicate symbol")
		}
		seen[member.Symbol] = true
		if member.Weight <= 0 {
			member.Weight = 1
		}
	}
	return peerMap, nil
}

func normalizePolicyCandidate(candidate PolicyCandidateAlert) (PolicyCandidateAlert, error) {
	candidate.Symbol = normalizeSymbol(candidate.Symbol)
	if candidate.Symbol == "" || candidate.AsOf.IsZero() || candidate.ProfileRole == "" || candidate.Horizon == "" || candidate.SessionProduct == "" || candidate.State == "" || candidate.ActionCandidate == "" || candidate.ThresholdVersion == "" {
		return candidate, errors.New("policy candidate requires symbol, asOf, profile, session, state, action, and threshold version")
	}
	if !validProfileRole(candidate.ProfileRole) || !validProfileHorizon(candidate.Horizon) {
		return candidate, errors.New("policy candidate contains an invalid profile role or horizon")
	}
	if candidate.TradingDate.IsZero() {
		candidate.TradingDate = dateOnlyUTC(candidate.AsOf)
	} else {
		candidate.TradingDate = dateOnlyUTC(candidate.TradingDate)
	}
	if candidate.DeliveryMode == "" {
		candidate.DeliveryMode = "shadow_only"
	}
	if candidate.DeliveryMode != "shadow_only" {
		return candidate, errors.New("policy candidate delivery mode must remain shadow_only")
	}
	if candidate.WouldNotify && (!candidate.Eligible || candidate.BudgetSlot <= 0) {
		return candidate, errors.New("would-notify policy candidates require eligibility and a budget slot")
	}
	candidate.ReasonCodes = defaultJSON(candidate.ReasonCodes, []byte("[]"))
	candidate.Metrics = defaultJSON(candidate.Metrics, []byte("{}"))
	candidate.Warnings = defaultJSON(candidate.Warnings, []byte("[]"))
	if candidate.CreatedAt.IsZero() {
		candidate.CreatedAt = time.Now().UTC()
	}
	return candidate, nil
}

func cloneProfile(profile SymbolProfile) SymbolProfile {
	profile.ActionPermissions = append(json.RawMessage(nil), profile.ActionPermissions...)
	if profile.CostBasis != nil {
		value := *profile.CostBasis
		profile.CostBasis = &value
	}
	return profile
}

func clonePeerMap(peerMap PeerMapVersion) PeerMapVersion {
	peerMap.Members = append([]PeerMapMember(nil), peerMap.Members...)
	return peerMap
}

func clonePolicyCandidate(candidate PolicyCandidateAlert) PolicyCandidateAlert {
	candidate.ReasonCodes = append(json.RawMessage(nil), candidate.ReasonCodes...)
	candidate.Metrics = append(json.RawMessage(nil), candidate.Metrics...)
	candidate.Warnings = append(json.RawMessage(nil), candidate.Warnings...)
	return candidate
}

func peerMapIncludes(peerMap PeerMapVersion, symbol string) bool {
	for _, member := range peerMap.Members {
		if member.Symbol == symbol {
			return true
		}
	}
	return false
}

func normalizeMapKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "_")
	return value
}

func validProfileRole(value string) bool {
	return value == "position" || value == "tactical_watch" || value == "research_watch" || value == "benchmark"
}

func validProfileHorizon(value string) bool {
	return value == "intraday" || value == "swing" || value == "medium_term" || value == "unspecified"
}

func policyOutcomeKey(candidateID int64, horizonMinutes int) string {
	return fmt.Sprintf("%d|%d", candidateID, horizonMinutes)
}

func (s *MemoryStore) hasPolicyCandidate(id int64) bool {
	for _, candidate := range s.policyCandidates {
		if candidate.ID == id {
			return true
		}
	}
	return false
}
