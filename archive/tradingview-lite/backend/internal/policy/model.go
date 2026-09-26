package policy

import "time"

const (
	StateBlocked                   = "BLOCKED"
	StateGapWithinRange            = "GAP_WITHIN_RANGE"
	StateGapExtended               = "GAP_EXTENDED"
	StateGapUnsupported            = "GAP_UNSUPPORTED"
	StatePriceDiscovery            = "PRICE_DISCOVERY"
	StateUnsupportedImpulse        = "UNSUPPORTED_IMPULSE"
	StateDislocationRisk           = "DISLOCATION_RISK"
	StateStabilizing               = "STABILIZING"
	StateReaccelerationRisk        = "RE_ACCELERATION_RISK"
	StateRWAUncalibratedDivergence = "RWA_DIVERGENCE_UNCALIBRATED"
	StateRWADivergence             = "RWA_DIVERGENCE"
	StateWaitConfirmation          = "WAIT_CONFIRMATION"

	ActionNoAction         = "NO_ACTION"
	ActionWaitConfirmation = "WAIT_CONFIRMATION"
	ActionNoChase          = "NO_CHASE"
	ActionHoldObserve      = "HOLD_OBSERVE"
	ActionReductionReview  = "REDUCTION_REVIEW"
	ActionStopReview       = "STOP_REVIEW"

	DeliveryModeShadowOnly = "shadow_only"
	ThresholdVersion       = "behavior_policy_shadow_v1"
)

type Config struct {
	DataStaleAfter       time.Duration
	CandidateCooldown    time.Duration
	DailyCandidateBudget int
	RWAAlignmentWindow   time.Duration
	RWAMinimumSamples    int
	RangeLookbackDays    int
	RangeMinimumDays     int
	ThresholdVersion     string
}

func normalizeConfig(cfg Config) Config {
	if cfg.DataStaleAfter <= 0 {
		cfg.DataStaleAfter = 2 * time.Minute
	}
	if cfg.CandidateCooldown <= 0 {
		cfg.CandidateCooldown = 30 * time.Minute
	}
	if cfg.DailyCandidateBudget <= 0 {
		cfg.DailyCandidateBudget = 5
	}
	if cfg.RWAAlignmentWindow <= 0 {
		cfg.RWAAlignmentWindow = 2 * time.Minute
	}
	if cfg.RWAMinimumSamples <= 0 {
		cfg.RWAMinimumSamples = 60
	}
	if cfg.RangeLookbackDays <= 0 {
		cfg.RangeLookbackDays = 90
	}
	if cfg.RangeMinimumDays <= 0 {
		cfg.RangeMinimumDays = 20
	}
	if cfg.ThresholdVersion == "" {
		cfg.ThresholdVersion = ThresholdVersion
	}
	return cfg
}

type RWAProxyCalibration struct {
	Symbol                 string    `json:"symbol"`
	AsOf                   time.Time `json:"asOf"`
	Status                 string    `json:"status"`
	MatchedSamples         int       `json:"matchedSamples"`
	ReturnPairs            int       `json:"returnPairs"`
	ContemporaneousCorr    float64   `json:"contemporaneousReturnCorrelation,omitempty"`
	MeanPremiumPercent     float64   `json:"meanPremiumPercent,omitempty"`
	LatestPremiumPercent   float64   `json:"latestPremiumPercent,omitempty"`
	MaxAlignmentAgeSeconds int64     `json:"maxAlignmentAgeSeconds"`
	Warning                string    `json:"warning,omitempty"`
}

type RangeCalibration struct {
	Symbol          string    `json:"symbol"`
	AsOf            time.Time `json:"asOf"`
	Status          string    `json:"status"`
	SampleDays      int       `json:"sampleDays"`
	CoveragePercent float64   `json:"coveragePercent,omitempty"`
	MeanRangeUse    float64   `json:"meanRangeUse,omitempty"`
	SelectedSources []string  `json:"selectedSources,omitempty"`
	Warning         string    `json:"warning,omitempty"`
}

type IVCalibration struct {
	Symbol                 string    `json:"symbol"`
	AsOf                   time.Time `json:"asOf"`
	Status                 string    `json:"status"`
	SnapshotCount          int       `json:"snapshotCount"`
	MatchedDays            int       `json:"matchedDays"`
	CoveragePercent        float64   `json:"coveragePercent,omitempty"`
	MeanImpliedMovePercent float64   `json:"meanImpliedMovePercent,omitempty"`
	MeanActualMovePercent  float64   `json:"meanActualMovePercent,omitempty"`
	MeanBiasPercent        float64   `json:"meanBiasPercent,omitempty"`
	Warning                string    `json:"warning,omitempty"`
}
