package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"tradingview-lite/backend/internal/alert"
	"tradingview-lite/backend/internal/evidence"
	"tradingview-lite/backend/internal/features"
	"tradingview-lite/backend/internal/market"
	"tradingview-lite/backend/internal/options"
	"tradingview-lite/backend/internal/policy"
	"tradingview-lite/backend/internal/riskcontext"
	"tradingview-lite/backend/internal/rwa"
	"tradingview-lite/backend/internal/session"
	"tradingview-lite/backend/internal/shadow"
	"tradingview-lite/backend/internal/store"
)

type MarketClient interface {
	FetchQuote(ctx context.Context, symbol string) (*market.Quote, error)
	FetchBars(ctx context.Context, symbol, rangeValue, interval string, includePrePost bool) (*market.BarsResponse, error)
}

type OptionsClient interface {
	FetchChain(ctx context.Context, symbol string, nearStrikes int) (*options.ChainResponse, error)
}

type OptionsBucketClient interface {
	FetchExpiryBuckets(ctx context.Context, symbol string, nearStrikes int) ([]*options.ChainResponse, error)
}

type RWAClient interface {
	StockInfo(ctx context.Context, symbol string) (*rwa.StockInfo, error)
	Kline(ctx context.Context, symbol, period string, size int) (*rwa.KlineResponse, error)
}

type EvidenceClient interface {
	Search(ctx context.Context, request evidence.SearchRequest) (*evidence.SearchResponse, error)
}

type AlertEvaluator interface {
	EvaluateAll(ctx context.Context) ([]alert.Event, error)
}

type Config struct {
	MarketInterval           time.Duration
	DataStaleAfter           time.Duration
	FocusQuoteInterval       time.Duration
	RWAInterval              time.Duration
	RWAFocusInterval         time.Duration
	OptionsInterval          time.Duration
	FeatureInterval          time.Duration
	AlertInterval            time.Duration
	RetentionInterval        time.Duration
	EventSymbolInterval      time.Duration
	EventMacroInterval       time.Duration
	EventContextInterval     time.Duration
	ShadowSignalInterval     time.Duration
	EventTriggeredCooldown   time.Duration
	MaxBackoff               time.Duration
	YahooRequestsPerMinute   int
	OptionsRequestsPerMinute int
	RWARequestsPerMinute     int
	EventRequestsPerMinute   int
	FocusSymbols             []string
	RWASymbols               []string
	RWAFocusSymbols          []string
	OptionSymbols            []string
	EventSymbols             []string
	ShadowSymbols            []string
	MacroQueries             []string
	GroupQueries             map[string]string
	OfficialDomains          []string
	EventMaxAgeMinutes       int
	PolicyEnabled            bool
	PolicyCandidateInterval  time.Duration
	PolicyOutcomeInterval    time.Duration
	PolicySymbols            []string
	PolicyConfig             policy.Config
	Retention                store.RetentionPolicy
}

type Result struct {
	Collector        string `json:"collector"`
	Provider         string `json:"provider"`
	SymbolsAttempted int    `json:"symbolsAttempted"`
	SymbolsSucceeded int    `json:"symbolsSucceeded"`
	RecordsWritten   int    `json:"recordsWritten"`
	RateLimited      bool   `json:"rateLimited"`
	ErrorSummary     string `json:"errorSummary,omitempty"`
}

type Collector struct {
	store         store.Store
	market        MarketClient
	options       OptionsClient
	rwa           RWAClient
	evidence      EvidenceClient
	alerts        AlertEvaluator
	config        Config
	yahooBudget   *requestBudget
	optionsBudget *requestBudget
	rwaBudget     *requestBudget
	eventBudget   *requestBudget

	mu               sync.RWMutex
	states           map[string]symbolState
	triggeredEventAt map[string]time.Time
}

type symbolState struct {
	quote       *market.Quote
	bars        *market.BarsResponse
	optionChain *options.ChainResponse
	rwaInfo     *rwa.StockInfo
	rwaBars     *rwa.KlineResponse
	feature     *features.IntradayResponse
}

func New(repo store.Store, marketClient MarketClient, optionsClient OptionsClient, rwaClient RWAClient, cfg Config) *Collector {
	config := normalizeConfig(cfg)
	return &Collector{
		store:            repo,
		market:           marketClient,
		options:          optionsClient,
		rwa:              rwaClient,
		config:           config,
		yahooBudget:      newRequestBudget(config.YahooRequestsPerMinute, time.Minute),
		optionsBudget:    newRequestBudget(config.OptionsRequestsPerMinute, time.Minute),
		rwaBudget:        newRequestBudget(config.RWARequestsPerMinute, time.Minute),
		eventBudget:      newRequestBudget(config.EventRequestsPerMinute, time.Minute),
		states:           map[string]symbolState{},
		triggeredEventAt: map[string]time.Time{},
	}
}

func (c *Collector) WithEvidence(client EvidenceClient) *Collector {
	c.evidence = client
	return c
}

func (c *Collector) WithAlertEvaluator(evaluator AlertEvaluator) *Collector {
	c.alerts = evaluator
	return c
}

func (c *Collector) Start(ctx context.Context) {
	c.startJob(ctx, "market", "yahoo_chart", c.config.MarketInterval, 0, c.CollectMarket)
	if c.market != nil && len(c.config.FocusSymbols) > 0 {
		c.startJob(ctx, "focus_quote", "yahoo_chart", c.config.FocusQuoteInterval, c.config.FocusQuoteInterval, c.CollectFocusQuotes)
	}
	if c.rwa != nil && len(c.config.RWASymbols) > 0 {
		c.startJob(ctx, "rwa", "bitget_rwa", c.config.RWAInterval, 0, c.CollectRWA)
	}
	if c.rwa != nil && len(c.config.RWAFocusSymbols) > 0 {
		c.startJob(ctx, "rwa_focus", "bitget_rwa", c.config.RWAFocusInterval, c.config.RWAFocusInterval, c.CollectFocusRWA)
	}
	if c.options != nil && len(c.config.OptionSymbols) > 0 {
		c.startJob(ctx, "options", "yahoo_mcp_options", c.config.OptionsInterval, 0, c.CollectOptions)
	}
	if c.evidence != nil && len(c.config.EventSymbols) > 0 {
		c.startJob(ctx, "event_symbol", evidence.ProviderFirecrawlNews, c.config.EventSymbolInterval, 0, c.CollectEventEvidence)
	}
	if c.evidence != nil && (len(c.config.MacroQueries) > 0 || len(c.config.GroupQueries) > 0) {
		c.startJob(ctx, "event_macro", evidence.ProviderFirecrawlNews, c.config.EventMacroInterval, c.config.EventMacroInterval, c.CollectMacroEvidence)
	}
	c.startJob(ctx, "features", "derived", c.config.FeatureInterval, c.config.FeatureInterval, c.CollectFeatures)
	if c.alerts != nil {
		c.startJob(ctx, "alerts", "rules", c.config.AlertInterval, c.config.AlertInterval, c.CollectAlerts)
	}
	if c.evidence != nil {
		c.startJob(ctx, "event_context", "derived", c.config.EventContextInterval, c.config.EventContextInterval, c.CollectEventContexts)
		c.startJob(ctx, "shadow_signal", "derived", c.config.ShadowSignalInterval, c.config.ShadowSignalInterval, c.CollectShadowSignals)
	}
	if c.config.PolicyEnabled {
		c.startJob(ctx, "behavior_policy", "derived", c.config.PolicyCandidateInterval, c.config.PolicyCandidateInterval, c.CollectPolicyCandidates)
		c.startJob(ctx, "policy_outcomes", "derived", c.config.PolicyOutcomeInterval, c.config.PolicyOutcomeInterval, c.CollectPolicyOutcomes)
	}
	c.startJob(ctx, "retention", "postgres_or_memory", c.config.RetentionInterval, c.config.RetentionInterval, c.CollectRetention)
}

func (c *Collector) CollectAlerts(ctx context.Context) (Result, error) {
	result := Result{Collector: "alerts", Provider: "rules"}
	if c.alerts == nil {
		return result, fmt.Errorf("alert evaluator is not configured")
	}
	events, err := c.alerts.EvaluateAll(ctx)
	if err != nil {
		return result, err
	}
	result.RecordsWritten = len(events)
	return result, nil
}

func (c *Collector) CollectMarket(ctx context.Context) (Result, error) {
	result := Result{Collector: "market", Provider: "yahoo_chart"}
	if c.market == nil {
		return result, fmt.Errorf("market collector is not configured")
	}
	items, err := c.store.ListWatchlist(ctx)
	if err != nil {
		return result, err
	}
	result.SymbolsAttempted = len(items)
	var errors []string
	for _, item := range items {
		if err := c.yahooBudget.Take(ctx); err != nil {
			errors = append(errors, item.Symbol+": quote budget: "+err.Error())
			continue
		}
		quote, err := c.market.FetchQuote(ctx, item.Symbol)
		if err != nil {
			errors = append(errors, item.Symbol+": quote: "+err.Error())
			result.RateLimited = result.RateLimited || rateLimited(err)
			continue
		}
		if err := c.store.UpsertQuote(ctx, *quote); err != nil {
			errors = append(errors, item.Symbol+": latest quote: "+err.Error())
			continue
		}
		if err := c.store.AppendQuoteSnapshot(ctx, quoteSnapshot(*quote)); err != nil {
			errors = append(errors, item.Symbol+": quote snapshot: "+err.Error())
			continue
		}
		result.RecordsWritten++

		if err := c.yahooBudget.Take(ctx); err != nil {
			errors = append(errors, item.Symbol+": bar budget: "+err.Error())
			c.updateMarketState(item.Symbol, quote, nil)
			result.SymbolsSucceeded++
			continue
		}
		bars, err := c.market.FetchBars(ctx, item.Symbol, "1d", "1m", true)
		if err != nil {
			errors = append(errors, item.Symbol+": bars: "+err.Error())
			result.RateLimited = result.RateLimited || rateLimited(err)
			c.updateMarketState(item.Symbol, quote, nil)
			result.SymbolsSucceeded++
			continue
		}
		count, err := c.store.UpsertBars(ctx, bars.Symbol, bars.Interval, bars.Provider, bars.Bars)
		if err != nil {
			errors = append(errors, item.Symbol+": store bars: "+err.Error())
			continue
		}
		result.RecordsWritten += count
		c.updateMarketState(item.Symbol, quote, bars)
		if err := c.updateDailySummary(ctx, item.Symbol); err != nil {
			errors = append(errors, item.Symbol+": daily summary: "+err.Error())
		}
		result.SymbolsSucceeded++
	}
	result.ErrorSummary = strings.Join(errors, "; ")
	return result, nil
}

func (c *Collector) CollectFocusQuotes(ctx context.Context) (Result, error) {
	result := Result{Collector: "focus_quote", Provider: "yahoo_chart", SymbolsAttempted: len(c.config.FocusSymbols)}
	if c.market == nil {
		return result, fmt.Errorf("market collector is not configured")
	}
	var errors []string
	for _, symbol := range c.config.FocusSymbols {
		if err := c.yahooBudget.Take(ctx); err != nil {
			errors = append(errors, symbol+": budget: "+err.Error())
			continue
		}
		quote, err := c.market.FetchQuote(ctx, symbol)
		if err != nil {
			errors = append(errors, symbol+": "+err.Error())
			result.RateLimited = result.RateLimited || rateLimited(err)
			continue
		}
		if err := c.store.UpsertQuote(ctx, *quote); err != nil {
			errors = append(errors, symbol+": latest quote: "+err.Error())
			continue
		}
		if err := c.store.AppendQuoteSnapshot(ctx, quoteSnapshot(*quote)); err != nil {
			errors = append(errors, symbol+": snapshot: "+err.Error())
			continue
		}
		c.updateMarketState(symbol, quote, nil)
		result.SymbolsSucceeded++
		result.RecordsWritten++
	}
	result.ErrorSummary = strings.Join(errors, "; ")
	return result, nil
}

func (c *Collector) CollectRWA(ctx context.Context) (Result, error) {
	return c.collectRWAForSymbols(ctx, "rwa", c.config.RWASymbols)
}

func (c *Collector) CollectFocusRWA(ctx context.Context) (Result, error) {
	return c.collectRWAForSymbols(ctx, "rwa_focus", c.config.RWAFocusSymbols)
}

func (c *Collector) collectRWAForSymbols(ctx context.Context, collectorName string, symbols []string) (Result, error) {
	result := Result{Collector: collectorName, Provider: "bitget_rwa", SymbolsAttempted: len(symbols)}
	if c.rwa == nil {
		return result, fmt.Errorf("RWA collector is not configured")
	}
	var errors []string
	for _, symbol := range symbols {
		if err := c.rwaBudget.Take(ctx); err != nil {
			errors = append(errors, symbol+": stock-info budget: "+err.Error())
			continue
		}
		info, err := c.rwa.StockInfo(ctx, symbol)
		if err != nil {
			errors = append(errors, symbol+": stock info: "+err.Error())
			result.RateLimited = result.RateLimited || rateLimited(err)
			continue
		}
		snapshot := store.RWASnapshot{
			Symbol:               info.Symbol,
			Ticker:               info.Ticker,
			Provider:             info.Provider,
			DataSource:           info.DataSource,
			Price:                info.LatestPrice,
			ReceivedAt:           info.ReceivedAt,
			Session:              session.ClassifyUS(info.ReceivedAt),
			ProviderMarketStatus: info.MarketStatus,
			Warning:              info.ProviderWarning,
		}
		if err := c.store.AppendRWASnapshot(ctx, snapshot); err != nil {
			errors = append(errors, symbol+": snapshot: "+err.Error())
			continue
		}
		result.RecordsWritten++
		var kline *rwa.KlineResponse
		if err := c.rwaBudget.Take(ctx); err != nil {
			errors = append(errors, symbol+": kline budget: "+err.Error())
			c.updateRWAState(symbol, info, nil)
			result.SymbolsSucceeded++
			continue
		}
		if fetched, err := c.rwa.Kline(ctx, symbol, "1m", 60); err == nil {
			kline = fetched
			count, err := c.store.UpsertRWABars(ctx, store.RWABars(symbol, kline, session.ClassifyUS))
			if err != nil {
				errors = append(errors, symbol+": bars: "+err.Error())
			} else {
				result.RecordsWritten += count
			}
		} else {
			errors = append(errors, symbol+": kline: "+err.Error())
			result.RateLimited = result.RateLimited || rateLimited(err)
		}
		c.updateRWAState(symbol, info, kline)
		result.SymbolsSucceeded++
	}
	result.ErrorSummary = strings.Join(errors, "; ")
	return result, nil
}

func (c *Collector) CollectOptions(ctx context.Context) (Result, error) {
	result := Result{Collector: "options", Provider: "yahoo_mcp_options", SymbolsAttempted: len(c.config.OptionSymbols)}
	if c.options == nil {
		return result, fmt.Errorf("options collector is not configured")
	}
	var errors []string
	for _, symbol := range c.config.OptionSymbols {
		if err := c.optionsBudget.Take(ctx); err != nil {
			errors = append(errors, symbol+": budget: "+err.Error())
			continue
		}
		chains := []*options.ChainResponse{}
		var err error
		if bucketClient, ok := c.options.(OptionsBucketClient); ok {
			chains, err = bucketClient.FetchExpiryBuckets(ctx, symbol, 6)
		} else {
			var chain *options.ChainResponse
			chain, err = c.options.FetchChain(ctx, symbol, 6)
			chains = []*options.ChainResponse{chain}
		}
		if err != nil {
			errors = append(errors, symbol+": "+err.Error())
			result.RateLimited = result.RateLimited || rateLimited(err)
			continue
		}
		written := 0
		for _, chain := range chains {
			count, saveErr := c.store.SaveOptionChain(ctx, chain)
			if saveErr != nil {
				errors = append(errors, symbol+": store: "+saveErr.Error())
				continue
			}
			written += count + 1
			c.updateOptionState(symbol, chain)
		}
		result.SymbolsSucceeded++
		result.RecordsWritten += written
	}
	result.ErrorSummary = strings.Join(errors, "; ")
	return result, nil
}

func (c *Collector) CollectEventEvidence(ctx context.Context) (Result, error) {
	result := Result{Collector: "event_symbol", Provider: evidence.ProviderFirecrawlNews, SymbolsAttempted: len(c.config.EventSymbols)}
	if c.evidence == nil {
		return result, fmt.Errorf("event evidence collector is not configured")
	}
	groups, err := c.watchlistGroups(ctx)
	if err != nil {
		return result, err
	}
	var errors []string
	for _, symbol := range c.config.EventSymbols {
		count, rateLimited, err := c.ingestNews(ctx, evidence.SearchRequest{
			Query:         symbol + " latest news",
			Tickers:       []string{symbol},
			Limit:         10,
			MaxAgeMinutes: c.config.EventMaxAgeMinutes,
			StrictTicker:  true,
		}, "symbol", groups[symbol])
		if err != nil {
			errors = append(errors, symbol+": "+err.Error())
			result.RateLimited = result.RateLimited || rateLimited
			continue
		}
		result.SymbolsSucceeded++
		result.RecordsWritten += count
	}
	result.ErrorSummary = strings.Join(errors, "; ")
	return result, nil
}

func (c *Collector) CollectMacroEvidence(ctx context.Context) (Result, error) {
	result := Result{Collector: "event_macro", Provider: evidence.ProviderFirecrawlNews, SymbolsAttempted: len(c.config.MacroQueries) + len(c.config.GroupQueries)}
	if c.evidence == nil {
		return result, fmt.Errorf("event evidence collector is not configured")
	}
	var errors []string
	for _, query := range c.config.MacroQueries {
		count, rateLimited, err := c.ingestNews(ctx, evidence.SearchRequest{
			Query:         query,
			Limit:         10,
			MaxAgeMinutes: c.config.EventMaxAgeMinutes,
			StrictTicker:  false,
		}, "macro", "")
		if err != nil {
			errors = append(errors, "macro: "+err.Error())
			result.RateLimited = result.RateLimited || rateLimited
			continue
		}
		result.SymbolsSucceeded++
		result.RecordsWritten += count
	}
	groupNames := make([]string, 0, len(c.config.GroupQueries))
	for group := range c.config.GroupQueries {
		groupNames = append(groupNames, group)
	}
	sort.Strings(groupNames)
	for _, group := range groupNames {
		count, rateLimited, err := c.ingestNews(ctx, evidence.SearchRequest{
			Query:         c.config.GroupQueries[group],
			Limit:         10,
			MaxAgeMinutes: c.config.EventMaxAgeMinutes,
			StrictTicker:  false,
		}, "group", group)
		if err != nil {
			errors = append(errors, group+": "+err.Error())
			result.RateLimited = result.RateLimited || rateLimited
			continue
		}
		result.SymbolsSucceeded++
		result.RecordsWritten += count
	}
	result.ErrorSummary = strings.Join(errors, "; ")
	return result, nil
}

func (c *Collector) CollectEventContexts(ctx context.Context) (Result, error) {
	result := Result{Collector: "event_context", Provider: "derived"}
	items, err := c.store.ListWatchlist(ctx)
	if err != nil {
		return result, err
	}
	result.SymbolsAttempted = len(items)
	builder := riskcontext.Builder{Store: c.store, DataStaleAfter: c.config.DataStaleAfter}
	asOf := time.Now().UTC()
	var errors []string
	for _, item := range items {
		snapshot, err := builder.Build(ctx, item.Symbol, item.Group, asOf)
		if err != nil {
			errors = append(errors, item.Symbol+": "+err.Error())
			continue
		}
		if _, err := c.store.AppendEventContext(ctx, snapshot); err != nil {
			errors = append(errors, item.Symbol+": store: "+err.Error())
			continue
		}
		result.SymbolsSucceeded++
		result.RecordsWritten++
	}
	result.ErrorSummary = strings.Join(errors, "; ")
	return result, nil
}

func (c *Collector) CollectShadowSignals(ctx context.Context) (Result, error) {
	result := Result{Collector: "shadow_signal", Provider: "derived", SymbolsAttempted: len(c.config.ShadowSymbols)}
	asOf := time.Now().UTC()
	evaluator := shadow.Evaluator{Store: c.store, Config: shadow.Config{DataStaleAfter: c.config.DataStaleAfter}}
	groups, err := c.watchlistGroups(ctx)
	if err != nil {
		return result, err
	}
	var errors []string
	for _, symbol := range c.config.ShadowSymbols {
		contextSnapshot, err := c.store.GetEventContext(ctx, symbol, asOf)
		if err != nil {
			errors = append(errors, symbol+": context: "+err.Error())
			continue
		}
		observation, err := evaluator.Evaluate(ctx, symbol, asOf, contextSnapshot)
		if err != nil {
			errors = append(errors, symbol+": "+err.Error())
			continue
		}
		if observation.State == shadow.StateImpulseForming && c.triggerAllowed(symbol, asOf) && c.evidence != nil {
			if _, rateLimited, err := c.ingestNews(ctx, evidence.SearchRequest{
				Query:         symbol + " latest news",
				Tickers:       []string{symbol},
				Limit:         10,
				MaxAgeMinutes: c.config.EventMaxAgeMinutes,
				StrictTicker:  true,
			}, "triggered", groups[symbol]); err != nil {
				errors = append(errors, symbol+": triggered evidence: "+err.Error())
				result.RateLimited = result.RateLimited || rateLimited
			} else {
				builder := riskcontext.Builder{Store: c.store, DataStaleAfter: c.config.DataStaleAfter}
				if refreshed, buildErr := builder.Build(ctx, symbol, groups[symbol], asOf); buildErr == nil {
					if saved, saveErr := c.store.AppendEventContext(ctx, refreshed); saveErr == nil {
						observation, _ = evaluator.Evaluate(ctx, symbol, asOf, &saved)
					}
				}
			}
		}
		features, err := c.store.ListFeatureSnapshots(ctx, symbol, store.TimeRange{To: asOf, Limit: 1000})
		if err == nil && len(features) > 0 {
			observation.FeatureSnapshotID = features[len(features)-1].ID
		}
		if _, err := c.store.AppendShadowSignal(ctx, observation); err != nil {
			errors = append(errors, symbol+": store: "+err.Error())
			continue
		}
		result.SymbolsSucceeded++
		result.RecordsWritten++
	}
	result.ErrorSummary = strings.Join(errors, "; ")
	return result, nil
}

func (c *Collector) CollectPolicyCandidates(ctx context.Context) (Result, error) {
	symbols := append([]string(nil), c.config.PolicySymbols...)
	if len(symbols) == 0 {
		items, err := c.store.ListWatchlist(ctx)
		if err != nil {
			return Result{Collector: "behavior_policy", Provider: "derived"}, err
		}
		for _, item := range items {
			symbols = append(symbols, item.Symbol)
		}
	}
	result := Result{Collector: "behavior_policy", Provider: "derived", SymbolsAttempted: len(symbols)}
	groups, err := c.watchlistGroups(ctx)
	if err != nil {
		return result, err
	}
	asOf := time.Now().UTC()
	evaluator := policy.Evaluator{Store: c.store, Config: c.config.PolicyConfig}
	var errors []string
	for _, symbol := range symbols {
		candidate, err := evaluator.Evaluate(ctx, symbol, groups[normalizeSymbol(symbol)], asOf)
		if err != nil {
			errors = append(errors, symbol+": "+err.Error())
			continue
		}
		if _, err := c.store.AppendPolicyCandidate(ctx, candidate); err != nil {
			errors = append(errors, symbol+": store: "+err.Error())
			continue
		}
		result.SymbolsSucceeded++
		result.RecordsWritten++
	}
	result.ErrorSummary = strings.Join(errors, "; ")
	return result, nil
}

func (c *Collector) CollectPolicyOutcomes(ctx context.Context) (Result, error) {
	symbols := append([]string(nil), c.config.PolicySymbols...)
	if len(symbols) == 0 {
		items, err := c.store.ListWatchlist(ctx)
		if err != nil {
			return Result{Collector: "policy_outcomes", Provider: "derived"}, err
		}
		for _, item := range items {
			symbols = append(symbols, item.Symbol)
		}
	}
	result := Result{Collector: "policy_outcomes", Provider: "derived", SymbolsAttempted: len(symbols)}
	evaluator := policy.OutcomeEvaluator{Store: c.store}
	var errors []string
	for _, symbol := range symbols {
		count, err := evaluator.Evaluate(ctx, symbol, time.Now().UTC())
		if err != nil {
			errors = append(errors, symbol+": "+err.Error())
			continue
		}
		result.SymbolsSucceeded++
		result.RecordsWritten += count
	}
	result.ErrorSummary = strings.Join(errors, "; ")
	return result, nil
}

func (c *Collector) ingestNews(ctx context.Context, request evidence.SearchRequest, scope, group string) (int, bool, error) {
	started := time.Now().UTC()
	requestedSymbols, _ := json.Marshal(evidence.NormalizeTickers(request.Tickers))
	if err := c.eventBudget.Take(ctx); err != nil {
		run, runErr := c.store.AddEventIngestionRun(ctx, store.EventIngestionRun{
			Provider: evidence.ProviderFirecrawlNews, Query: request.Query, Scope: scope, GroupName: group,
			RequestedSymbols: requestedSymbols, StrictTicker: request.StrictTicker, StartedAt: started,
			FinishedAt: time.Now().UTC(), ErrorSummary: "request budget: " + err.Error(),
		})
		_ = run
		if runErr != nil {
			return 0, false, runErr
		}
		return 0, false, err
	}
	response, err := c.evidence.Search(ctx, request)
	finished := time.Now().UTC()
	if err != nil {
		_, runErr := c.store.AddEventIngestionRun(ctx, store.EventIngestionRun{
			Provider: evidence.ProviderFirecrawlNews, Query: request.Query, Scope: scope, GroupName: group,
			RequestedSymbols: requestedSymbols, StrictTicker: request.StrictTicker, StartedAt: started,
			FinishedAt: finished, RateLimited: evidence.IsRateLimited(err), ErrorSummary: err.Error(),
		})
		if runErr != nil {
			return 0, evidence.IsRateLimited(err), runErr
		}
		return 0, evidence.IsRateLimited(err), err
	}
	metadata, _ := json.Marshal(response.Metadata)
	run, err := c.store.AddEventIngestionRun(ctx, store.EventIngestionRun{
		Provider: evidence.ProviderFirecrawlNews, Query: request.Query, Scope: scope, GroupName: group,
		RequestedSymbols: requestedSymbols, StrictTicker: response.Metadata.StrictTicker,
		ProviderRoute: response.Metadata.ProviderRoute, StartedAt: started, FinishedAt: finished,
		ResultCount: len(response.Data), RecordsWritten: len(response.Data), Metadata: metadata,
	})
	if err != nil {
		return 0, false, err
	}
	rows := makeEvidenceRows(response, request, scope, group, c.config.OfficialDomains, finished)
	written, err := c.store.SaveEventEvidence(ctx, run.ID, rows)
	return written, false, err
}

func (c *Collector) watchlistGroups(ctx context.Context) (map[string]string, error) {
	items, err := c.store.ListWatchlist(ctx)
	if err != nil {
		return nil, err
	}
	groups := map[string]string{}
	for _, item := range items {
		groups[item.Symbol] = item.Group
	}
	return groups, nil
}

func (c *Collector) triggerAllowed(symbol string, now time.Time) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	previous := c.triggeredEventAt[normalizeSymbol(symbol)]
	if !previous.IsZero() && now.Sub(previous) < c.config.EventTriggeredCooldown {
		return false
	}
	c.triggeredEventAt[normalizeSymbol(symbol)] = now
	return true
}

func makeEvidenceRows(response *evidence.SearchResponse, request evidence.SearchRequest, scope, group string, officialDomains []string, receivedAt time.Time) []store.EventEvidence {
	if response == nil {
		return nil
	}
	rows := make([]store.EventEvidence, 0, len(response.Data))
	requested := evidence.NormalizeTickers(request.Tickers)
	for _, item := range response.Data {
		matched := evidence.NormalizeTickers(item.Tickers)
		warning := ""
		if scope == "symbol" || scope == "triggered" {
			if len(requested) != 1 || !containsSymbol(matched, requested[0]) {
				warning = "strict_ticker_result_missing_explicit_requested_symbol"
			}
		}
		domain := evidence.SourceDomain(item.URL, item.Source)
		tier := evidence.SourceTier(domain, officialDomains)
		if warning != "" {
			tier = "T4_UNVERIFIED"
		}
		metadata, _ := json.Marshal(map[string]any{
			"itemMetadata":  json.RawMessage(item.Metadata),
			"providerRoute": response.Metadata.ProviderRoute,
			"strictTicker":  response.Metadata.StrictTicker,
			"rankScore":     item.RankScore,
		})
		matchedJSON, _ := json.Marshal(matched)
		rows = append(rows, store.EventEvidence{
			CanonicalURL:   evidence.CanonicalURL(item.URL),
			HeadlineHash:   evidence.HeadlineHash(item.Title),
			Title:          item.Title,
			Snippet:        item.Snippet,
			SourceDomain:   domain,
			Provider:       item.Provider,
			Publisher:      item.Publisher,
			SourceTier:     tier,
			PublishedAt:    evidence.ParseTime(item.PublishedAt),
			DiscoveredAt:   evidence.ParseTime(item.DiscoveredAt),
			ReceivedAt:     receivedAt,
			Freshness:      item.FreshnessConfidence,
			MatchedSymbols: matchedJSON,
			Scope:          scope,
			GroupName:      group,
			EventType:      normalizedEventType(item.EventType),
			Metadata:       metadata,
			Warning:        warning,
		})
	}
	return rows
}

func containsSymbol(symbols []string, target string) bool {
	for _, symbol := range symbols {
		if symbol == target {
			return true
		}
	}
	return false
}

func normalizedEventType(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "other"
	}
	return value
}

func (c *Collector) CollectFeatures(ctx context.Context) (Result, error) {
	result := Result{Collector: "features", Provider: "derived"}
	items, err := c.store.ListWatchlist(ctx)
	if err != nil {
		return result, err
	}
	result.SymbolsAttempted = len(items)
	var errors []string
	for _, item := range items {
		state, ok := c.state(item.Symbol)
		if !ok || state.quote == nil || state.bars == nil {
			continue
		}
		feature, err := features.ComputeFromData(features.IntradayInput{
			Quote:       state.quote,
			Bars:        state.bars,
			OptionChain: state.optionChain,
			RWAInfo:     state.rwaInfo,
			RWABars:     state.rwaBars,
			ComputedAt:  time.Now().UTC(),
		})
		if err != nil {
			errors = append(errors, item.Symbol+": "+err.Error())
			continue
		}
		payload, err := json.Marshal(feature)
		if err != nil {
			errors = append(errors, item.Symbol+": marshal: "+err.Error())
			continue
		}
		providerTimes, _ := json.Marshal(feature.ProviderTimes)
		warnings, _ := json.Marshal(feature.Warnings)
		freshness, _ := json.Marshal(inputFreshness(state))
		snapshot := store.FeatureSnapshot{
			Symbol:          feature.Symbol,
			ComputedAt:      feature.ReceivedAt,
			ProviderTimes:   providerTimes,
			InputFreshness:  freshness,
			SelectedSource:  feature.ReasonableRange.SelectedSource,
			SelectedMovePct: feature.ReasonableRange.SelectedOneDayMovePercent,
			Warnings:        warnings,
			Payload:         payload,
		}
		if err := c.store.AppendFeatureSnapshot(ctx, snapshot); err != nil {
			errors = append(errors, item.Symbol+": store: "+err.Error())
			continue
		}
		c.updateFeatureState(item.Symbol, feature)
		if err := c.updateDailySummary(ctx, item.Symbol); err != nil {
			errors = append(errors, item.Symbol+": daily summary: "+err.Error())
		}
		result.SymbolsSucceeded++
		result.RecordsWritten++
	}
	result.ErrorSummary = strings.Join(errors, "; ")
	return result, nil
}

func (c *Collector) CollectRetention(ctx context.Context) (Result, error) {
	result := Result{Collector: "retention", Provider: "postgres_or_memory", SymbolsAttempted: 1}
	pruned, err := c.store.PruneSnapshots(ctx, c.config.Retention, time.Now().UTC())
	if err != nil {
		return result, err
	}
	result.SymbolsSucceeded = 1
	_ = pruned
	result.RecordsWritten = 0
	return result, nil
}

func (c *Collector) startJob(ctx context.Context, collectorName, provider string, interval, initialDelay time.Duration, execute func(context.Context) (Result, error)) {
	go func() {
		if initialDelay > 0 {
			timer := time.NewTimer(initialDelay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
		}
		backoff := time.Duration(0)
		for {
			started := time.Now().UTC()
			result, err := execute(ctx)
			finished := time.Now().UTC()
			if result.Collector == "" {
				result.Collector = collectorName
			}
			if result.Provider == "" {
				result.Provider = provider
			}
			if err != nil {
				result.ErrorSummary = joinError(result.ErrorSummary, err.Error())
				result.RateLimited = result.RateLimited || rateLimited(err)
			}
			if result.ErrorSummary != "" || err != nil {
				backoff = nextBackoff(backoff, c.config.MaxBackoff)
			} else {
				backoff = 0
			}
			if saveErr := c.store.AddCollectionRun(ctx, store.CollectionRun{
				Collector:        result.Collector,
				Provider:         result.Provider,
				StartedAt:        started,
				FinishedAt:       finished,
				SymbolsAttempted: result.SymbolsAttempted,
				SymbolsSucceeded: result.SymbolsSucceeded,
				RecordsWritten:   result.RecordsWritten,
				RateLimited:      result.RateLimited,
				BackoffSeconds:   int64(backoff.Seconds()),
				ErrorSummary:     result.ErrorSummary,
			}); saveErr != nil {
				log.Printf("collector %s: record run: %v", collectorName, saveErr)
			}
			if result.ErrorSummary != "" {
				log.Printf("collector %s completed with errors: %s", collectorName, result.ErrorSummary)
			}
			wait := interval
			if backoff > wait {
				wait = backoff
			}
			timer := time.NewTimer(wait)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
		}
	}()
}

func (c *Collector) updateMarketState(symbol string, quote *market.Quote, bars *market.BarsResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()
	state := c.states[normalizeSymbol(symbol)]
	if quote != nil {
		state.quote = quote
	}
	if bars != nil {
		state.bars = bars
	}
	c.states[normalizeSymbol(symbol)] = state
}

func (c *Collector) updateRWAState(symbol string, info *rwa.StockInfo, bars *rwa.KlineResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()
	state := c.states[normalizeSymbol(symbol)]
	state.rwaInfo = info
	state.rwaBars = bars
	c.states[normalizeSymbol(symbol)] = state
}

func (c *Collector) updateOptionState(symbol string, chain *options.ChainResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()
	state := c.states[normalizeSymbol(symbol)]
	state.optionChain = chain
	c.states[normalizeSymbol(symbol)] = state
}

func (c *Collector) updateFeatureState(symbol string, feature *features.IntradayResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()
	state := c.states[normalizeSymbol(symbol)]
	state.feature = feature
	c.states[normalizeSymbol(symbol)] = state
}

func (c *Collector) state(symbol string) (symbolState, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	state, ok := c.states[normalizeSymbol(symbol)]
	return state, ok
}

func (c *Collector) updateDailySummary(ctx context.Context, symbol string) error {
	state, ok := c.state(symbol)
	if !ok || state.quote == nil {
		return nil
	}
	quote := state.quote
	tradingTime := quote.RegularMarketTime
	if state.bars != nil && len(state.bars.Bars) > 0 {
		tradingTime = state.bars.Bars[len(state.bars.Bars)-1].Time
	}
	if tradingTime.IsZero() {
		tradingTime = quote.ReceivedAt
	}
	tradingDate := session.TradingDate(tradingTime)
	if tradingDate.IsZero() {
		return nil
	}
	summary := store.DailyMarketSummary{
		Symbol:      quote.Symbol,
		TradingDate: tradingDate,
		Provider:    quote.Provider,
		UpdatedAt:   time.Now().UTC(),
	}
	if quote.DataAgeSeconds <= 120 && session.IsPreOpenFiveMinutes(quote.RegularMarketTime) {
		summary.PreOpenFiveMinute = quote.Price
	}
	if state.bars != nil {
		for _, bar := range state.bars.Bars {
			if !session.TradingDate(bar.Time).Equal(tradingDate) {
				continue
			}
			barSession := session.ClassifyUS(bar.Time)
			summary.DailyHigh = maxPrice(summary.DailyHigh, bar.High)
			summary.DailyLow = minPrice(summary.DailyLow, bar.Low)
			if barSession == session.Premarket {
				summary.PremarketHigh = maxPrice(summary.PremarketHigh, bar.High)
				summary.PremarketLow = minPrice(summary.PremarketLow, bar.Low)
			}
			if barSession == session.Regular && summary.RegularOpen == 0 {
				summary.RegularOpen = bar.Open
			}
		}
	}
	if state.optionChain != nil {
		summary.PreviousClose = state.optionChain.Underlying.PreviousClose
	}
	if state.feature != nil {
		summary.SelectedMove = state.feature.ReasonableRange.SelectedOneDayMove
		summary.SelectedMovePct = state.feature.ReasonableRange.SelectedOneDayMovePercent
		summary.SelectedSource = state.feature.ReasonableRange.SelectedSource
	}
	return c.store.UpsertDailySummary(ctx, summary)
}

func quoteSnapshot(quote market.Quote) store.QuoteSnapshot {
	return store.QuoteSnapshot{
		Symbol:         quote.Symbol,
		Provider:       quote.Provider,
		Price:          quote.Price,
		ProviderTime:   quote.RegularMarketTime,
		ReceivedAt:     quote.ReceivedAt,
		DataAgeSeconds: quote.DataAgeSeconds,
		Session:        session.ClassifyUS(quote.ReceivedAt),
		Warning:        quote.ProviderWarning,
	}
}

func inputFreshness(state symbolState) map[string]any {
	values := map[string]any{}
	if state.quote != nil {
		values["yahoo_quote"] = map[string]any{
			"providerTime":   state.quote.RegularMarketTime,
			"receivedAt":     state.quote.ReceivedAt,
			"dataAgeSeconds": state.quote.DataAgeSeconds,
		}
	}
	if state.bars != nil {
		values["yahoo_bars"] = map[string]any{
			"providerTime":   state.bars.RegularMarketTime,
			"receivedAt":     state.bars.ReceivedAt,
			"dataAgeSeconds": state.bars.DataAgeSeconds,
		}
	}
	if state.optionChain != nil {
		values["yahoo_options"] = map[string]any{
			"receivedAt":                  state.optionChain.ReceivedAt,
			"underlyingProviderTime":      state.optionChain.Underlying.RegularMarketTime,
			"underlyingDataAgeSeconds":    state.optionChain.Summary.UnderlyingDataAgeSeconds,
			"contractTradeDataAgeSeconds": state.optionChain.Summary.ContractTradeDataAgeSeconds,
		}
	}
	if state.rwaInfo != nil {
		values["bitget_rwa"] = map[string]any{"receivedAt": state.rwaInfo.ReceivedAt, "marketStatus": state.rwaInfo.MarketStatus}
	}
	return values
}

func normalizeConfig(cfg Config) Config {
	if cfg.DataStaleAfter <= 0 {
		cfg.DataStaleAfter = 2 * time.Minute
	}
	if cfg.MarketInterval <= 0 {
		cfg.MarketInterval = time.Minute
	}
	if cfg.FocusQuoteInterval <= 0 {
		cfg.FocusQuoteInterval = 30 * time.Second
	}
	if cfg.RWAInterval <= 0 {
		cfg.RWAInterval = 30 * time.Second
	}
	if cfg.RWAFocusInterval <= 0 {
		cfg.RWAFocusInterval = 15 * time.Second
	}
	if cfg.OptionsInterval <= 0 {
		cfg.OptionsInterval = 5 * time.Minute
	}
	if cfg.AlertInterval <= 0 {
		cfg.AlertInterval = time.Minute
	}
	if cfg.FeatureInterval <= 0 {
		cfg.FeatureInterval = time.Minute
	}
	if cfg.EventSymbolInterval <= 0 {
		cfg.EventSymbolInterval = 15 * time.Minute
	}
	if cfg.EventMacroInterval <= 0 {
		cfg.EventMacroInterval = 10 * time.Minute
	}
	if cfg.EventContextInterval <= 0 {
		cfg.EventContextInterval = time.Minute
	}
	if cfg.ShadowSignalInterval <= 0 {
		cfg.ShadowSignalInterval = 30 * time.Second
	}
	if cfg.EventTriggeredCooldown <= 0 {
		cfg.EventTriggeredCooldown = 2 * time.Minute
	}
	if cfg.RetentionInterval <= 0 {
		cfg.RetentionInterval = 24 * time.Hour
	}
	if cfg.MaxBackoff <= 0 {
		cfg.MaxBackoff = 10 * time.Minute
	}
	if cfg.YahooRequestsPerMinute <= 0 {
		cfg.YahooRequestsPerMinute = 50
	}
	if cfg.OptionsRequestsPerMinute <= 0 {
		cfg.OptionsRequestsPerMinute = 20
	}
	if cfg.RWARequestsPerMinute <= 0 {
		cfg.RWARequestsPerMinute = 40
	}
	if cfg.EventRequestsPerMinute <= 0 {
		cfg.EventRequestsPerMinute = 10
	}
	if cfg.EventMaxAgeMinutes <= 0 {
		cfg.EventMaxAgeMinutes = 1440
	}
	if cfg.PolicyCandidateInterval <= 0 {
		cfg.PolicyCandidateInterval = time.Minute
	}
	if cfg.PolicyOutcomeInterval <= 0 {
		cfg.PolicyOutcomeInterval = 5 * time.Minute
	}
	cfg.FocusSymbols = normalizeSymbols(cfg.FocusSymbols, 5)
	cfg.RWASymbols = normalizeSymbols(cfg.RWASymbols, 0)
	cfg.RWAFocusSymbols = normalizeSymbols(cfg.RWAFocusSymbols, 5)
	cfg.OptionSymbols = normalizeSymbols(cfg.OptionSymbols, 0)
	cfg.EventSymbols = normalizeSymbols(cfg.EventSymbols, 0)
	cfg.ShadowSymbols = normalizeSymbols(cfg.ShadowSymbols, 0)
	cfg.PolicySymbols = normalizeSymbols(cfg.PolicySymbols, 0)
	return cfg
}

func normalizeSymbols(symbols []string, max int) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		symbol = normalizeSymbol(symbol)
		if symbol == "" || seen[symbol] {
			continue
		}
		seen[symbol] = true
		out = append(out, symbol)
		if max > 0 && len(out) >= max {
			break
		}
	}
	return out
}

func normalizeSymbol(symbol string) string {
	return strings.ToUpper(strings.TrimSpace(symbol))
}

func rateLimited(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "429") || strings.Contains(text, "rate limit") || strings.Contains(text, "too many requests")
}

func nextBackoff(current, maximum time.Duration) time.Duration {
	if current <= 0 {
		return time.Second
	}
	next := current * 2
	if next > maximum {
		return maximum
	}
	return next
}

type requestBudget struct {
	maxPerWindow int
	window       time.Duration

	mu          sync.Mutex
	windowStart time.Time
	used        int
}

func newRequestBudget(maxPerWindow int, window time.Duration) *requestBudget {
	return &requestBudget{maxPerWindow: maxPerWindow, window: window}
}

func (b *requestBudget) Take(ctx context.Context) error {
	if b == nil || b.maxPerWindow <= 0 || b.window <= 0 {
		return nil
	}
	for {
		now := time.Now()
		b.mu.Lock()
		if b.windowStart.IsZero() || now.Sub(b.windowStart) >= b.window {
			b.windowStart = now
			b.used = 0
		}
		if b.used < b.maxPerWindow {
			b.used++
			b.mu.Unlock()
			return nil
		}
		wait := time.Until(b.windowStart.Add(b.window))
		b.mu.Unlock()
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func joinError(existing, next string) string {
	if existing == "" {
		return next
	}
	if next == "" {
		return existing
	}
	return existing + "; " + next
}

func maxPrice(current, next float64) float64 {
	return math.Max(current, next)
}

func minPrice(current, next float64) float64 {
	if current == 0 || (next > 0 && next < current) {
		return next
	}
	return current
}
