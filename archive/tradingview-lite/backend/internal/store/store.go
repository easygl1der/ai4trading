package store

import (
	"context"
	"time"

	"tradingview-lite/backend/internal/alert"
	"tradingview-lite/backend/internal/market"
	"tradingview-lite/backend/internal/options"
)

type WatchlistItem struct {
	Symbol string `json:"symbol"`
	Group  string `json:"group,omitempty"`
}

type Store interface {
	ListWatchlist(ctx context.Context) ([]WatchlistItem, error)
	AddWatchlist(ctx context.Context, item WatchlistItem) error
	DeleteWatchlist(ctx context.Context, symbol string) error

	UpsertBars(ctx context.Context, symbol, timeframe, provider string, bars []market.Bar) (int, error)
	GetBars(ctx context.Context, symbol, timeframe string, limit int) ([]market.Bar, error)

	UpsertQuote(ctx context.Context, quote market.Quote) error
	GetQuote(ctx context.Context, symbol string) (*market.Quote, error)
	ListQuotes(ctx context.Context) ([]market.Quote, error)
	AppendQuoteSnapshot(ctx context.Context, snapshot QuoteSnapshot) error
	ListQuoteSnapshots(ctx context.Context, symbol, provider string, bounds TimeRange) ([]QuoteSnapshot, error)

	AppendRWASnapshot(ctx context.Context, snapshot RWASnapshot) error
	ListRWASnapshots(ctx context.Context, symbol, provider string, bounds TimeRange) ([]RWASnapshot, error)
	UpsertRWABars(ctx context.Context, rows []RWABarSnapshot) (int, error)
	GetRWABars(ctx context.Context, symbol, timeframe string, limit int) ([]RWABarSnapshot, error)

	SaveOptionChain(ctx context.Context, chain *options.ChainResponse) (int, error)
	ListOptionContracts(ctx context.Context, symbol string, bounds TimeRange) ([]OptionContractSnapshot, error)
	ListIVSnapshots(ctx context.Context, symbol string, bounds TimeRange) ([]IVSnapshot, error)

	AppendFeatureSnapshot(ctx context.Context, snapshot FeatureSnapshot) error
	ListFeatureSnapshots(ctx context.Context, symbol string, bounds TimeRange) ([]FeatureSnapshot, error)
	UpsertDailySummary(ctx context.Context, summary DailyMarketSummary) error
	GetDailySummary(ctx context.Context, symbol string, tradingDate time.Time) (*DailyMarketSummary, error)
	ListDailySummaries(ctx context.Context, symbol string, bounds TimeRange) ([]DailyMarketSummary, error)

	AddCollectionRun(ctx context.Context, run CollectionRun) error
	ListCollectionRuns(ctx context.Context, limit int) ([]CollectionRun, error)
	AddEventIngestionRun(ctx context.Context, run EventIngestionRun) (EventIngestionRun, error)
	ListEventIngestionRuns(ctx context.Context, limit int) ([]EventIngestionRun, error)
	SaveEventEvidence(ctx context.Context, runID int64, rows []EventEvidence) (int, error)
	ListEventEvidence(ctx context.Context, filter EventFilter) ([]EventEvidence, error)
	ListMarketEvents(ctx context.Context, filter EventFilter) ([]MarketEvent, error)
	AppendEventContext(ctx context.Context, snapshot EventContextSnapshot) (EventContextSnapshot, error)
	GetEventContext(ctx context.Context, symbol string, asOf time.Time) (*EventContextSnapshot, error)
	AppendShadowSignal(ctx context.Context, observation ShadowSignalObservation) (ShadowSignalObservation, error)
	ListShadowSignals(ctx context.Context, symbol string, bounds TimeRange) ([]ShadowSignalObservation, error)

	ListSymbolProfiles(ctx context.Context) ([]SymbolProfile, error)
	GetSymbolProfile(ctx context.Context, symbol string) (*SymbolProfile, error)
	UpsertSymbolProfile(ctx context.Context, profile SymbolProfile) (SymbolProfile, error)
	CreatePeerMapVersion(ctx context.Context, peerMap PeerMapVersion) (PeerMapVersion, error)
	ListPeerMapVersions(ctx context.Context, mapKey string) ([]PeerMapVersion, error)
	GetActivePeerMap(ctx context.Context, symbol, group string, asOf time.Time) (*PeerMapVersion, error)
	AppendPolicyCandidate(ctx context.Context, candidate PolicyCandidateAlert) (PolicyCandidateAlert, error)
	GetPolicyCandidate(ctx context.Context, id int64) (*PolicyCandidateAlert, error)
	ListPolicyCandidates(ctx context.Context, filter PolicyCandidateFilter) ([]PolicyCandidateAlert, error)
	UpsertPolicyCandidateOutcome(ctx context.Context, outcome PolicyCandidateOutcome) (PolicyCandidateOutcome, error)
	ListPolicyCandidateOutcomes(ctx context.Context, candidateID int64) ([]PolicyCandidateOutcome, error)
	UpsertPolicyCandidateFeedback(ctx context.Context, feedback PolicyCandidateFeedback) (PolicyCandidateFeedback, error)
	GetPolicyCandidateFeedback(ctx context.Context, candidateID int64) (*PolicyCandidateFeedback, error)
	PruneSnapshots(ctx context.Context, policy RetentionPolicy, now time.Time) (RetentionResult, error)
	StorageDiagnostics(ctx context.Context) (StorageDiagnostic, error)

	CreateAlert(ctx context.Context, cfg alert.AlertConfig) (alert.AlertConfig, error)
	ListAlerts(ctx context.Context) ([]alert.AlertConfig, error)
	AddEvent(ctx context.Context, event alert.Event) error
	ListEvents(ctx context.Context, limit int) ([]alert.Event, error)
	EnqueueDiscordDelivery(ctx context.Context, delivery DiscordDelivery) (bool, error)
	ClaimDiscordDeliveries(ctx context.Context, now time.Time, limit int) ([]DiscordDelivery, error)
	CompleteDiscordDelivery(ctx context.Context, id int64) error
	RetryDiscordDelivery(ctx context.Context, id int64, nextAttemptAt time.Time, message string, deadLetter bool) error
}
