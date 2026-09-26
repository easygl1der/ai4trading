package alert

import "time"

type Level string

const (
	LevelInfo            Level = "INFO"
	LevelWatch           Level = "WATCH"
	LevelAlert           Level = "ALERT"
	LevelUrgent          Level = "URGENT"
	LevelAnalysisTrigger Level = "ANALYSIS_TRIGGER"
	LevelBlocked         Level = "BLOCKED"
)

type RuleType string

const (
	RulePriceLevel               RuleType = "price_level"
	RuleIntradayMove             RuleType = "intraday_move"
	RuleRelativeUnderperformance RuleType = "relative_underperformance"
	RuleVolumeSpike              RuleType = "volume_spike"
	RuleDataStale                RuleType = "data_stale"
)

type Rule struct {
	Type         RuleType `json:"type"`
	Operator     string   `json:"operator,omitempty"`
	Value        float64  `json:"value,omitempty"`
	ThresholdPct float64  `json:"threshold_pct,omitempty"`
	Benchmark    string   `json:"benchmark,omitempty"`
	Window       string   `json:"window,omitempty"`
	ZScore       float64  `json:"zscore,omitempty"`
	Level        Level    `json:"level,omitempty"`
}

type AlertConfig struct {
	ID                string    `json:"id"`
	Symbol            string    `json:"symbol"`
	Group             string    `json:"group,omitempty"`
	Rules             []Rule    `json:"rules"`
	Notify            []string  `json:"notify,omitempty"`
	AnalysisOnTrigger bool      `json:"analysis_on_trigger"`
	Enabled           bool      `json:"enabled"`
	CreatedAt         time.Time `json:"created_at"`
}

type Event struct {
	ID                string    `json:"id"`
	AlertID           string    `json:"alert_id,omitempty"`
	Symbol            string    `json:"symbol"`
	Level             Level     `json:"level"`
	RuleType          RuleType  `json:"rule_type,omitempty"`
	Message           string    `json:"message"`
	ObservedValue     float64   `json:"observed_value,omitempty"`
	ThresholdValue    float64   `json:"threshold_value,omitempty"`
	Benchmark         string    `json:"benchmark,omitempty"`
	DataAgeSeconds    int64     `json:"data_age_seconds"`
	ProviderTime      time.Time `json:"provider_time"`
	ReceivedAt        time.Time `json:"received_at"`
	TriggeredAt       time.Time `json:"triggered_at"`
	AnalysisOnTrigger bool      `json:"analysis_on_trigger"`
}

func (r Rule) EffectiveLevel() Level {
	if r.Level != "" {
		return r.Level
	}
	return LevelAlert
}
