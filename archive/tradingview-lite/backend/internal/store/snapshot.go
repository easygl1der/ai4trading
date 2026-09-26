package store

import (
	"encoding/json"
	"time"

	"tradingview-lite/backend/internal/alert"
	"tradingview-lite/backend/internal/options"
	"tradingview-lite/backend/internal/rwa"
)

type TimeRange struct {
	From  time.Time
	To    time.Time
	Limit int
}

type QuoteSnapshot struct {
	ID             int64     `json:"id"`
	Symbol         string    `json:"symbol"`
	Provider       string    `json:"provider"`
	Price          float64   `json:"price"`
	ProviderTime   time.Time `json:"providerTime"`
	ReceivedAt     time.Time `json:"receivedAt"`
	DataAgeSeconds int64     `json:"dataAgeSeconds"`
	Session        string    `json:"session"`
	MarketState    string    `json:"marketState,omitempty"`
	Warning        string    `json:"warning,omitempty"`
}

type RWASnapshot struct {
	ID                   int64     `json:"id"`
	Symbol               string    `json:"symbol"`
	Ticker               string    `json:"ticker"`
	Provider             string    `json:"provider"`
	DataSource           string    `json:"dataSource,omitempty"`
	Price                float64   `json:"price"`
	ReceivedAt           time.Time `json:"receivedAt"`
	Session              string    `json:"session"`
	ProviderMarketStatus string    `json:"providerMarketStatus,omitempty"`
	Warning              string    `json:"warning,omitempty"`
}

type RWABarSnapshot struct {
	Symbol     string    `json:"symbol"`
	Ticker     string    `json:"ticker"`
	Provider   string    `json:"provider"`
	Timeframe  string    `json:"timeframe"`
	Timestamp  time.Time `json:"timestamp"`
	Open       float64   `json:"open"`
	High       float64   `json:"high"`
	Low        float64   `json:"low"`
	Close      float64   `json:"close"`
	Volume     float64   `json:"volume"`
	Amount     float64   `json:"amount"`
	ReceivedAt time.Time `json:"receivedAt"`
	Session    string    `json:"session"`
}

type OptionContractSnapshot struct {
	ID                        int64     `json:"id"`
	Symbol                    string    `json:"symbol"`
	Provider                  string    `json:"provider"`
	ContractSymbol            string    `json:"contractSymbol"`
	ExpirationDate            time.Time `json:"expirationDate"`
	Strike                    float64   `json:"strike"`
	Right                     string    `json:"right"`
	Bid                       float64   `json:"bid"`
	Ask                       float64   `json:"ask"`
	Mid                       float64   `json:"mid"`
	Last                      float64   `json:"last"`
	Volume                    int       `json:"volume"`
	OpenInterest              int       `json:"openInterest"`
	RawImpliedVolatility      float64   `json:"rawImpliedVolatility"`
	ResolvedImpliedVolatility float64   `json:"resolvedImpliedVolatility"`
	ResolvedIVError           string    `json:"resolvedIVError,omitempty"`
	IVInputPrice              float64   `json:"ivInputPrice"`
	IVInputSource             string    `json:"ivInputSource,omitempty"`
	ContractTradeAt           time.Time `json:"contractTradeAt,omitempty"`
	UnderlyingSpot            float64   `json:"underlyingSpot"`
	UnderlyingProviderTime    time.Time `json:"underlyingProviderTime,omitempty"`
	ReceivedAt                time.Time `json:"receivedAt"`
	Warning                   string    `json:"warning,omitempty"`
}

type IVSnapshot struct {
	ID                          int64     `json:"id"`
	Symbol                      string    `json:"symbol"`
	Provider                    string    `json:"provider"`
	ExpirationDate              time.Time `json:"expirationDate"`
	ATMStrike                   float64   `json:"atmStrike"`
	Spot                        float64   `json:"spot"`
	CallResolvedIV              float64   `json:"callResolvedIV"`
	PutResolvedIV               float64   `json:"putResolvedIV"`
	AverageResolvedIV           float64   `json:"averageResolvedIV"`
	StraddleMove                float64   `json:"straddleMove"`
	StraddleMovePercent         float64   `json:"straddleMovePercent"`
	OneDayMove                  float64   `json:"oneDayMove"`
	OneDayMovePercent           float64   `json:"oneDayMovePercent"`
	RiskFreeRate                float64   `json:"riskFreeRate"`
	DividendYieldAssumption     float64   `json:"dividendYieldAssumption"`
	UnderlyingProviderTime      time.Time `json:"underlyingProviderTime,omitempty"`
	UnderlyingDataAgeSeconds    int64     `json:"underlyingDataAgeSeconds"`
	ContractTradeDataAgeSeconds int64     `json:"contractTradeDataAgeSeconds"`
	ReceivedAt                  time.Time `json:"receivedAt"`
	Warning                     string    `json:"warning,omitempty"`
}

type FeatureSnapshot struct {
	ID              int64           `json:"id"`
	Symbol          string          `json:"symbol"`
	ComputedAt      time.Time       `json:"computedAt"`
	ProviderTimes   json.RawMessage `json:"providerTimes"`
	InputFreshness  json.RawMessage `json:"inputFreshness"`
	SelectedSource  string          `json:"selectedSource,omitempty"`
	SelectedMovePct float64         `json:"selectedMovePct"`
	Warnings        json.RawMessage `json:"warnings"`
	Payload         json.RawMessage `json:"payload"`
}

type CollectionRun struct {
	ID               int64     `json:"id"`
	Collector        string    `json:"collector"`
	Provider         string    `json:"provider"`
	StartedAt        time.Time `json:"startedAt"`
	FinishedAt       time.Time `json:"finishedAt"`
	SymbolsAttempted int       `json:"symbolsAttempted"`
	SymbolsSucceeded int       `json:"symbolsSucceeded"`
	RecordsWritten   int       `json:"recordsWritten"`
	RateLimited      bool      `json:"rateLimited"`
	BackoffSeconds   int64     `json:"backoffSeconds"`
	ErrorSummary     string    `json:"errorSummary,omitempty"`
}

type DailyMarketSummary struct {
	Symbol            string    `json:"symbol"`
	TradingDate       time.Time `json:"tradingDate"`
	Provider          string    `json:"provider"`
	PreviousClose     float64   `json:"previousClose"`
	PremarketHigh     float64   `json:"premarketHigh"`
	PremarketLow      float64   `json:"premarketLow"`
	PreOpenFiveMinute float64   `json:"preOpenFiveMinute"`
	RegularOpen       float64   `json:"regularOpen"`
	DailyHigh         float64   `json:"dailyHigh"`
	DailyLow          float64   `json:"dailyLow"`
	SelectedMove      float64   `json:"selectedMove"`
	SelectedMovePct   float64   `json:"selectedMovePct"`
	SelectedSource    string    `json:"selectedSource,omitempty"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type EventIngestionRun struct {
	ID               int64           `json:"id"`
	Provider         string          `json:"provider"`
	Query            string          `json:"query"`
	Scope            string          `json:"scope"`
	GroupName        string          `json:"groupName,omitempty"`
	RequestedSymbols json.RawMessage `json:"requestedSymbols"`
	StrictTicker     bool            `json:"strictTicker"`
	ProviderRoute    string          `json:"providerRoute,omitempty"`
	StartedAt        time.Time       `json:"startedAt"`
	FinishedAt       time.Time       `json:"finishedAt"`
	ResultCount      int             `json:"resultCount"`
	RecordsWritten   int             `json:"recordsWritten"`
	RateLimited      bool            `json:"rateLimited"`
	BackoffSeconds   int64           `json:"backoffSeconds"`
	ErrorSummary     string          `json:"errorSummary,omitempty"`
	Metadata         json.RawMessage `json:"metadata,omitempty"`
}

type EventEvidence struct {
	ID             int64           `json:"id"`
	IngestionRunID int64           `json:"ingestionRunId,omitempty"`
	CanonicalURL   string          `json:"canonicalUrl"`
	HeadlineHash   string          `json:"headlineHash"`
	Title          string          `json:"title"`
	Snippet        string          `json:"snippet,omitempty"`
	SourceDomain   string          `json:"sourceDomain,omitempty"`
	Provider       string          `json:"provider"`
	Publisher      string          `json:"publisher,omitempty"`
	SourceTier     string          `json:"sourceTier"`
	PublishedAt    time.Time       `json:"publishedAt,omitempty"`
	DiscoveredAt   time.Time       `json:"discoveredAt,omitempty"`
	ReceivedAt     time.Time       `json:"receivedAt"`
	Freshness      string          `json:"freshness,omitempty"`
	MatchedSymbols json.RawMessage `json:"matchedSymbols"`
	Scope          string          `json:"scope"`
	GroupName      string          `json:"groupName,omitempty"`
	EventType      string          `json:"eventType"`
	Metadata       json.RawMessage `json:"metadata,omitempty"`
	Warning        string          `json:"warning,omitempty"`
	EventID        int64           `json:"eventId,omitempty"`
}

type MarketEvent struct {
	ID             int64     `json:"id"`
	EventKey       string    `json:"eventKey"`
	Scope          string    `json:"scope"`
	PrimarySymbol  string    `json:"primarySymbol,omitempty"`
	GroupName      string    `json:"groupName,omitempty"`
	EventType      string    `json:"eventType"`
	EvidenceStatus string    `json:"evidenceStatus"`
	Severity       string    `json:"severity"`
	State          string    `json:"state"`
	FirstSeenAt    time.Time `json:"firstSeenAt"`
	LastSeenAt     time.Time `json:"lastSeenAt"`
}

type DiscordDelivery struct {
	ID             int64       `json:"id"`
	IdempotencyKey string      `json:"idempotencyKey"`
	Event          alert.Event `json:"event"`
	Status         string      `json:"status"`
	Attempts       int         `json:"attempts"`
	NextAttemptAt  time.Time   `json:"nextAttemptAt"`
	LastError      string      `json:"lastError,omitempty"`
}

type EventFilter struct {
	Symbol string
	Group  string
	Bounds TimeRange
}

type EventContextSnapshot struct {
	ID               int64           `json:"id"`
	Symbol           string          `json:"symbol"`
	AsOf             time.Time       `json:"asOf"`
	MarketDataFresh  bool            `json:"marketDataFresh"`
	QuoteAgeSeconds  int64           `json:"quoteAgeSeconds"`
	Session          string          `json:"session"`
	SelectedMovePct  float64         `json:"selectedMovePct"`
	EvidenceCoverage string          `json:"evidenceCoverage"`
	LastIngestionAt  time.Time       `json:"lastIngestionAt,omitempty"`
	KnownEventIDs    json.RawMessage `json:"knownEventIds"`
	RiskState        string          `json:"riskState"`
	ReasonCodes      json.RawMessage `json:"reasonCodes"`
	Warnings         json.RawMessage `json:"warnings"`
	Payload          json.RawMessage `json:"payload"`
	CreatedAt        time.Time       `json:"createdAt"`
}

type ShadowSignalObservation struct {
	ID                int64           `json:"id"`
	SignalType        string          `json:"signalType"`
	Symbol            string          `json:"symbol"`
	AsOf              time.Time       `json:"asOf"`
	State             string          `json:"state"`
	Eligible          bool            `json:"eligible"`
	BlockedReason     string          `json:"blockedReason,omitempty"`
	ThresholdVersion  string          `json:"thresholdVersion"`
	EventContextID    int64           `json:"eventContextId,omitempty"`
	FeatureSnapshotID int64           `json:"featureSnapshotId,omitempty"`
	ReasonCodes       json.RawMessage `json:"reasonCodes"`
	Metrics           json.RawMessage `json:"metrics"`
	CreatedAt         time.Time       `json:"createdAt"`
}

type SymbolProfile struct {
	Symbol            string          `json:"symbol"`
	Role              string          `json:"role"`
	Horizon           string          `json:"horizon"`
	ActionPermissions json.RawMessage `json:"actionPermissions"`
	CostBasis         *float64        `json:"costBasis,omitempty"`
	Notes             string          `json:"notes,omitempty"`
	UpdatedAt         time.Time       `json:"updatedAt"`
}

type PeerMapMember struct {
	Symbol   string  `json:"symbol"`
	Relation string  `json:"relation"`
	Weight   float64 `json:"weight"`
}

type PeerMapVersion struct {
	ID              int64           `json:"id"`
	MapKey          string          `json:"mapKey"`
	Version         int             `json:"version"`
	GroupName       string          `json:"groupName,omitempty"`
	BenchmarkSymbol string          `json:"benchmarkSymbol,omitempty"`
	Methodology     string          `json:"methodology"`
	ReviewStatus    string          `json:"reviewStatus"`
	EffectiveFrom   time.Time       `json:"effectiveFrom"`
	CreatedAt       time.Time       `json:"createdAt"`
	Members         []PeerMapMember `json:"members"`
}

type PolicyCandidateFilter struct {
	Symbol          string
	TradingDate     time.Time
	OnlyWouldNotify bool
	Bounds          TimeRange
}

type PolicyCandidateAlert struct {
	ID               int64           `json:"id"`
	Symbol           string          `json:"symbol"`
	TradingDate      time.Time       `json:"tradingDate"`
	AsOf             time.Time       `json:"asOf"`
	ProfileRole      string          `json:"profileRole"`
	Horizon          string          `json:"horizon"`
	SessionProduct   string          `json:"sessionProduct"`
	State            string          `json:"state"`
	ActionCandidate  string          `json:"actionCandidate"`
	Direction        string          `json:"direction,omitempty"`
	ReferencePrice   float64         `json:"referencePrice"`
	Eligible         bool            `json:"eligible"`
	WouldNotify      bool            `json:"wouldNotify"`
	BudgetSlot       int             `json:"budgetSlot,omitempty"`
	DeliveryMode     string          `json:"deliveryMode"`
	SuppressedReason string          `json:"suppressedReason,omitempty"`
	ThresholdVersion string          `json:"thresholdVersion"`
	EventContextID   int64           `json:"eventContextId,omitempty"`
	PeerMapVersionID int64           `json:"peerMapVersionId,omitempty"`
	ReasonCodes      json.RawMessage `json:"reasonCodes"`
	Metrics          json.RawMessage `json:"metrics"`
	Warnings         json.RawMessage `json:"warnings"`
	CreatedAt        time.Time       `json:"createdAt"`
}

type PolicyCandidateOutcome struct {
	ID               int64     `json:"id"`
	CandidateAlertID int64     `json:"candidateAlertId"`
	HorizonMinutes   int       `json:"horizonMinutes"`
	TargetAt         time.Time `json:"targetAt"`
	ObservedAt       time.Time `json:"observedAt"`
	ObservedPrice    float64   `json:"observedPrice"`
	ReturnPercent    float64   `json:"returnPercent"`
	MaxUpPercent     float64   `json:"maxUpPercent"`
	MaxDownPercent   float64   `json:"maxDownPercent"`
	SampleCount      int       `json:"sampleCount"`
	Warning          string    `json:"warning,omitempty"`
	EvaluatedAt      time.Time `json:"evaluatedAt"`
}

type PolicyCandidateFeedback struct {
	CandidateAlertID int64     `json:"candidateAlertId"`
	Helpful          *bool     `json:"helpful,omitempty"`
	Acted            *bool     `json:"acted,omitempty"`
	EmotionIntensity *int      `json:"emotionIntensity,omitempty"`
	Notes            string    `json:"notes,omitempty"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type RetentionPolicy struct {
	QuoteSnapshots   time.Duration
	RWASnapshots     time.Duration
	OptionSnapshots  time.Duration
	FeatureSnapshots time.Duration
	CollectionRuns   time.Duration
	EventEvidence    time.Duration
	EventContexts    time.Duration
	ShadowSignals    time.Duration
	EventRuns        time.Duration
	PolicyCandidates time.Duration
	PolicyOutcomes   time.Duration
}

type RetentionResult struct {
	QuoteSnapshots   int64 `json:"quoteSnapshots"`
	RWASnapshots     int64 `json:"rwaSnapshots"`
	OptionSnapshots  int64 `json:"optionSnapshots"`
	IVSnapshots      int64 `json:"ivSnapshots"`
	FeatureSnapshots int64 `json:"featureSnapshots"`
	CollectionRuns   int64 `json:"collectionRuns"`
	EventEvidence    int64 `json:"eventEvidence"`
	EventContexts    int64 `json:"eventContexts"`
	ShadowSignals    int64 `json:"shadowSignals"`
	EventRuns        int64 `json:"eventRuns"`
	PolicyCandidates int64 `json:"policyCandidates"`
	PolicyOutcomes   int64 `json:"policyOutcomes"`
}

type StorageDiagnostic struct {
	Store        string           `json:"store"`
	TableBytes   map[string]int64 `json:"tableBytes"`
	RowEstimates map[string]int64 `json:"rowEstimates"`
	CollectedAt  time.Time        `json:"collectedAt"`
}

func OptionContracts(chain *options.ChainResponse) []OptionContractSnapshot {
	if chain == nil {
		return nil
	}
	contracts := make([]OptionContractSnapshot, 0, len(chain.Calls)+len(chain.Puts))
	appendContracts := func(rows []options.Contract) {
		for _, row := range rows {
			contracts = append(contracts, OptionContractSnapshot{
				Symbol:                    chain.Symbol,
				Provider:                  chain.Provider,
				ContractSymbol:            row.ContractSymbol,
				ExpirationDate:            chain.ExpirationDate,
				Strike:                    row.Strike,
				Right:                     row.Type,
				Bid:                       row.Bid,
				Ask:                       row.Ask,
				Mid:                       row.Mid,
				Last:                      row.LastPrice,
				Volume:                    row.Volume,
				OpenInterest:              row.OpenInterest,
				RawImpliedVolatility:      row.ImpliedVolatility,
				ResolvedImpliedVolatility: row.ResolvedImpliedVol,
				ResolvedIVError:           row.ResolvedImpliedVolError,
				IVInputPrice:              row.IVInputPrice,
				IVInputSource:             row.IVInputSource,
				ContractTradeAt:           row.LastTradeDate,
				UnderlyingSpot:            chain.Summary.Spot,
				UnderlyingProviderTime:    chain.Underlying.RegularMarketTime,
				ReceivedAt:                chain.ReceivedAt,
				Warning:                   chain.ProviderWarning,
			})
		}
	}
	appendContracts(chain.Calls)
	appendContracts(chain.Puts)
	return contracts
}

func IVFromChain(chain *options.ChainResponse) *IVSnapshot {
	if chain == nil {
		return nil
	}
	summary := chain.Summary
	return &IVSnapshot{
		Symbol:                      chain.Symbol,
		Provider:                    chain.Provider,
		ExpirationDate:              chain.ExpirationDate,
		ATMStrike:                   summary.ATMStrike,
		Spot:                        summary.Spot,
		CallResolvedIV:              summary.ResolvedCallImpliedVolatility,
		PutResolvedIV:               summary.ResolvedPutImpliedVolatility,
		AverageResolvedIV:           summary.AverageResolvedImpliedVolatility,
		StraddleMove:                summary.ExpectedMove,
		StraddleMovePercent:         summary.ExpectedMovePercent,
		OneDayMove:                  summary.OneTradingDayExpectedMove,
		OneDayMovePercent:           summary.OneTradingDayExpectedMovePercent,
		RiskFreeRate:                summary.RiskFreeRate,
		DividendYieldAssumption:     summary.DividendYieldAssumption,
		UnderlyingProviderTime:      chain.Underlying.RegularMarketTime,
		UnderlyingDataAgeSeconds:    summary.UnderlyingDataAgeSeconds,
		ContractTradeDataAgeSeconds: summary.ContractTradeDataAgeSeconds,
		ReceivedAt:                  chain.ReceivedAt,
		Warning:                     chain.ProviderWarning + "; " + summary.MethodWarning,
	}
}

func RWABars(symbol string, response *rwa.KlineResponse, sessionFor func(time.Time) string) []RWABarSnapshot {
	if response == nil {
		return nil
	}
	rows := make([]RWABarSnapshot, 0, len(response.Bars))
	for _, bar := range response.Bars {
		rows = append(rows, RWABarSnapshot{
			Symbol:     symbol,
			Ticker:     response.Ticker,
			Provider:   response.Provider,
			Timeframe:  response.Period,
			Timestamp:  bar.Time,
			Open:       bar.Open,
			High:       bar.High,
			Low:        bar.Low,
			Close:      bar.Close,
			Volume:     bar.Volume,
			Amount:     bar.Amount,
			ReceivedAt: response.ReceivedAt,
			Session:    sessionFor(bar.Time),
		})
	}
	return rows
}
