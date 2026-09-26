package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Host                    string
	Port                    string
	DatabaseURL             string
	PollInterval            time.Duration
	DataStaleAfter          time.Duration
	DefaultWatchlistSymbols []string
	EnableScheduler         bool
	Collection              CollectionConfig
	Discord                 DiscordConfig
	RWA                     RWAConfig
	Options                 OptionsConfig
	Events                  EventConfig
	Policy                  PolicyConfig
}

type CollectionConfig struct {
	Enabled                  bool
	MarketInterval           time.Duration
	FocusQuoteInterval       time.Duration
	RWAInterval              time.Duration
	RWAFocusInterval         time.Duration
	OptionsInterval          time.Duration
	FeatureInterval          time.Duration
	AlertInterval            time.Duration
	RetentionInterval        time.Duration
	MaxBackoff               time.Duration
	YahooRequestsPerMinute   int
	OptionsRequestsPerMinute int
	RWARequestsPerMinute     int
	FocusSymbols             []string
	RWASymbols               []string
	RWAFocusSymbols          []string
	OptionSymbols            []string
	QuoteRetention           time.Duration
	RWARetention             time.Duration
	OptionRetention          time.Duration
	FeatureRetention         time.Duration
	RunRetention             time.Duration
}

type DiscordConfig struct {
	Watch            string
	Alerts           string
	Urgent           string
	Daily            string
	System           string
	DeliveryInterval time.Duration
}

type RWAConfig struct {
	BitgetMCPURL   string
	BitgetMCPToken string
	SymbolMap      map[string]string
}

type OptionsConfig struct {
	YahooMCPURL   string
	YahooMCPToken string
	RiskFreeRate  float64
}

type EventConfig struct {
	Enabled           bool
	FirecrawlNewsURL  string
	FirecrawlAPIKey   string
	SymbolInterval    time.Duration
	MacroInterval     time.Duration
	ContextInterval   time.Duration
	ShadowInterval    time.Duration
	TriggeredCooldown time.Duration
	RequestsPerMinute int
	MaxAgeMinutes     int
	Symbols           []string
	ShadowSymbols     []string
	MacroQueries      []string
	GroupQueries      map[string]string
	OfficialDomains   []string
	EvidenceRetention time.Duration
	ContextRetention  time.Duration
	ShadowRetention   time.Duration
	RunRetention      time.Duration
}

type PolicyConfig struct {
	Enabled              bool
	CandidateInterval    time.Duration
	OutcomeInterval      time.Duration
	CandidateCooldown    time.Duration
	DailyCandidateBudget int
	Symbols              []string
	CandidateRetention   time.Duration
	OutcomeRetention     time.Duration
}

func Load() Config {
	return Config{
		Host:                    env("APP_HOST", "127.0.0.1"),
		Port:                    env("APP_PORT", "8088"),
		DatabaseURL:             env("DATABASE_URL", ""),
		PollInterval:            envDuration("POLL_INTERVAL", 60*time.Second),
		DataStaleAfter:          envDuration("DATA_STALE_AFTER", 120*time.Second),
		DefaultWatchlistSymbols: envList("WATCHLIST_SYMBOLS", "MU,NVDA,AMD,SMH,RKLB,SPCX,ARKX,TE,FCEL,IONQ,RGTI,QBTS,QUBT,SPY,QQQ,IWM,TLT,^VIX"),
		EnableScheduler:         envBool("ENABLE_SCHEDULER", true),
		Collection: CollectionConfig{
			Enabled:                  envBool("ENABLE_COLLECTORS", true),
			MarketInterval:           envDuration("COLLECTOR_MARKET_INTERVAL", envDuration("POLL_INTERVAL", 60*time.Second)),
			FocusQuoteInterval:       envDuration("FOCUS_QUOTE_INTERVAL", 30*time.Second),
			RWAInterval:              envDuration("RWA_COLLECT_INTERVAL", 30*time.Second),
			RWAFocusInterval:         envDuration("RWA_FOCUS_INTERVAL", 15*time.Second),
			OptionsInterval:          envDuration("OPTIONS_COLLECT_INTERVAL", 5*time.Minute),
			FeatureInterval:          envDuration("FEATURE_COLLECT_INTERVAL", 60*time.Second),
			AlertInterval:            envDuration("ALERT_EVALUATION_INTERVAL", time.Minute),
			RetentionInterval:        envDuration("RETENTION_INTERVAL", 24*time.Hour),
			MaxBackoff:               envDuration("COLLECTOR_MAX_BACKOFF", 10*time.Minute),
			YahooRequestsPerMinute:   envInt("YAHOO_REQUESTS_PER_MINUTE", 50),
			OptionsRequestsPerMinute: envInt("OPTIONS_REQUESTS_PER_MINUTE", 20),
			RWARequestsPerMinute:     envInt("RWA_REQUESTS_PER_MINUTE", 40),
			FocusSymbols:             envList("FOCUS_SYMBOLS", ""),
			RWASymbols:               envList("RWA_SYMBOLS", "NVDA,AMD"),
			RWAFocusSymbols:          envList("RWA_FOCUS_SYMBOLS", "NVDA"),
			OptionSymbols:            envList("OPTIONS_SYMBOLS", "MU,NVDA,AMD,SMH,SPY,QQQ"),
			QuoteRetention:           envDuration("QUOTE_SNAPSHOT_RETENTION", 14*24*time.Hour),
			RWARetention:             envDuration("RWA_SNAPSHOT_RETENTION", 14*24*time.Hour),
			OptionRetention:          envDuration("OPTION_SNAPSHOT_RETENTION", 90*24*time.Hour),
			FeatureRetention:         envDuration("FEATURE_SNAPSHOT_RETENTION", 90*24*time.Hour),
			RunRetention:             envDuration("COLLECTION_RUN_RETENTION", 90*24*time.Hour),
		},
		Discord: DiscordConfig{
			Watch:            env("DISCORD_WEBHOOK_WATCH", ""),
			Alerts:           env("DISCORD_WEBHOOK_ALERTS", ""),
			Urgent:           env("DISCORD_WEBHOOK_URGENT", ""),
			Daily:            env("DISCORD_WEBHOOK_DAILY", ""),
			System:           env("DISCORD_WEBHOOK_SYSTEM", ""),
			DeliveryInterval: envDuration("DISCORD_DELIVERY_INTERVAL", 10*time.Second),
		},
		RWA: RWAConfig{
			BitgetMCPURL:   env("BITGET_MCP_URL", ""),
			BitgetMCPToken: env("BITGET_MCP_TOKEN", ""),
			SymbolMap:      envMap("RWA_SYMBOL_MAP", defaultRWASymbolMap()),
		},
		Options: OptionsConfig{
			YahooMCPURL:   env("YAHOO_MCP_URL", ""),
			YahooMCPToken: env("YAHOO_MCP_TOKEN", ""),
			RiskFreeRate:  envFloat("OPTIONS_RISK_FREE_RATE", 0.045),
		},
		Events: EventConfig{
			Enabled:           envBool("ENABLE_EVENT_EVIDENCE", true),
			FirecrawlNewsURL:  env("FIRECRAWL_NEWS_URL", ""),
			FirecrawlAPIKey:   env("FIRECRAWL_API_KEY", ""),
			SymbolInterval:    envDuration("EVENT_SYMBOL_INTERVAL", 15*time.Minute),
			MacroInterval:     envDuration("EVENT_MACRO_INTERVAL", 10*time.Minute),
			ContextInterval:   envDuration("EVENT_CONTEXT_INTERVAL", time.Minute),
			ShadowInterval:    envDuration("SHADOW_SIGNAL_INTERVAL", 30*time.Second),
			TriggeredCooldown: envDuration("EVENT_TRIGGERED_COOLDOWN", 2*time.Minute),
			RequestsPerMinute: envInt("FIRECRAWL_REQUESTS_PER_MINUTE", 10),
			MaxAgeMinutes:     envInt("FIRECRAWL_NEWS_MAX_AGE_MINUTES", 1440),
			Symbols:           envList("EVENT_SYMBOLS", "MU,NVDA,AMD,SMH,RKLB,SPCX,ARKX,TE,FCEL,IONQ,RGTI,QBTS,QUBT,SPY,QQQ,IWM,TLT,^VIX"),
			ShadowSymbols:     envList("SHADOW_SYMBOLS", "MU,NVDA,RKLB,IONQ,SPY,QQQ"),
			MacroQueries:      envList("EVENT_MACRO_QUERIES", "SPY QQQ market moving news,Federal Reserve CPI jobs report"),
			GroupQueries: envNamedQueries("EVENT_GROUP_QUERIES", map[string]string{
				"semis_memory": "semiconductor stocks market news",
				"space":        "space stocks market news",
				"energy_clean": "clean energy stocks market news",
				"quantum":      "quantum computing stocks market news",
			}),
			OfficialDomains:   envList("EVENT_OFFICIAL_DOMAINS", "sec.gov,federalreserve.gov,bls.gov,bea.gov"),
			EvidenceRetention: envDuration("EVENT_EVIDENCE_RETENTION", 90*24*time.Hour),
			ContextRetention:  envDuration("EVENT_CONTEXT_RETENTION", 90*24*time.Hour),
			ShadowRetention:   envDuration("SHADOW_SIGNAL_RETENTION", 90*24*time.Hour),
			RunRetention:      envDuration("EVENT_RUN_RETENTION", 90*24*time.Hour),
		},
		Policy: PolicyConfig{
			Enabled:              envBool("ENABLE_BEHAVIOR_POLICY_SHADOW", true),
			CandidateInterval:    envDuration("POLICY_CANDIDATE_INTERVAL", time.Minute),
			OutcomeInterval:      envDuration("POLICY_OUTCOME_INTERVAL", 5*time.Minute),
			CandidateCooldown:    envDuration("POLICY_CANDIDATE_COOLDOWN", 30*time.Minute),
			DailyCandidateBudget: envInt("POLICY_DAILY_CANDIDATE_BUDGET", 5),
			Symbols:              envList("POLICY_SHADOW_SYMBOLS", "MU,NVDA,RKLB,IONQ,SPY,QQQ"),
			CandidateRetention:   envDuration("POLICY_CANDIDATE_RETENTION", 90*24*time.Hour),
			OutcomeRetention:     envDuration("POLICY_OUTCOME_RETENTION", 90*24*time.Hour),
		},
	}
}

func (c DiscordConfig) Enabled() bool {
	return c.Watch != "" || c.Alerts != "" || c.Urgent != "" || c.Daily != "" || c.System != ""
}

func (c RWAConfig) Enabled() bool {
	return c.BitgetMCPURL != ""
}

func (c OptionsConfig) Enabled() bool {
	return c.YahooMCPURL != ""
}

func (c EventConfig) EnabledForCollection() bool {
	return c.Enabled && c.FirecrawlNewsURL != "" && c.FirecrawlAPIKey != ""
}

func env(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err == nil {
		return parsed
	}
	seconds, err := strconv.Atoi(value)
	if err == nil {
		return time.Duration(seconds) * time.Second
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	return value == "1" || value == "true" || value == "yes" || value == "on"
}

func envFloat(key string, fallback float64) float64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func envList(key, fallback string) []string {
	raw := env(key, fallback)
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		symbol := strings.ToUpper(strings.TrimSpace(part))
		if symbol == "" || seen[symbol] {
			continue
		}
		seen[symbol] = true
		out = append(out, symbol)
	}
	return out
}

func envMap(key string, fallback map[string]string) map[string]string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	out := map[string]string{}
	for _, part := range strings.Split(raw, ",") {
		pair := strings.SplitN(part, "=", 2)
		if len(pair) != 2 {
			continue
		}
		symbol := strings.ToUpper(strings.TrimSpace(pair[0]))
		ticker := strings.TrimSpace(pair[1])
		if symbol == "" || ticker == "" {
			continue
		}
		out[symbol] = ticker
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}

func envNamedQueries(key string, fallback map[string]string) map[string]string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		result := make(map[string]string, len(fallback))
		for name, query := range fallback {
			result[name] = query
		}
		return result
	}
	result := map[string]string{}
	for _, pair := range strings.Split(raw, "|") {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}
		name := strings.ToLower(strings.TrimSpace(parts[0]))
		query := strings.TrimSpace(parts[1])
		if name != "" && query != "" {
			result[name] = query
		}
	}
	if len(result) == 0 {
		return fallback
	}
	return result
}

func defaultRWASymbolMap() map[string]string {
	return map[string]string{
		"AAPL":  "AAPLon",
		"AMD":   "AMDon",
		"GOOGL": "GOOGLon",
		"META":  "METAon",
		"MSFT":  "MSFTon",
		"NVDA":  "NVDAon",
		"TSLA":  "TSLAon",
	}
}
