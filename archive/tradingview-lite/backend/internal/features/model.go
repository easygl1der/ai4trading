package features

import "time"

type IntradayResponse struct {
	Symbol          string            `json:"symbol"`
	ReceivedAt      time.Time         `json:"receivedAt"`
	CurrentPrice    float64           `json:"currentPrice"`
	PreviousClose   float64           `json:"previousClose,omitempty"`
	RegularOpen     float64           `json:"regularOpen,omitempty"`
	ProviderTimes   map[string]string `json:"providerTimes,omitempty"`
	Realized        RealizedVol       `json:"realized"`
	PriceSpeed      []SpeedPoint      `json:"priceSpeed"`
	RWA             *RWAFeature       `json:"rwa,omitempty"`
	Options         *OptionsFeature   `json:"options,omitempty"`
	ReasonableRange ReasonableRange   `json:"reasonableRange"`
	Warnings        []string          `json:"warnings,omitempty"`
}

type RealizedVol struct {
	BarsUsed                   int     `json:"barsUsed"`
	RegularBarsUsed            int     `json:"regularBarsUsed"`
	RealizedMovePercent        float64 `json:"realizedMovePercent"`
	RegularRealizedMovePercent float64 `json:"regularRealizedMovePercent"`
	Last30MinMovePercent       float64 `json:"last30MinMovePercent"`
	ProjectedDailyMovePercent  float64 `json:"projectedDailyMovePercent"`
	Method                     string  `json:"method"`
}

type SpeedPoint struct {
	Horizon              string    `json:"horizon"`
	ElapsedSeconds       int64     `json:"elapsedSeconds"`
	FromTime             time.Time `json:"fromTime"`
	ToTime               time.Time `json:"toTime"`
	FromPrice            float64   `json:"fromPrice"`
	ToPrice              float64   `json:"toPrice"`
	ReturnPercent        float64   `json:"returnPercent"`
	VelocityPctPerSecond float64   `json:"velocityPctPerSecond"`
}

type RWAFeature struct {
	Symbol       string       `json:"symbol"`
	Ticker       string       `json:"ticker"`
	LatestPrice  float64      `json:"latestPrice"`
	MarketStatus string       `json:"marketStatus,omitempty"`
	DataSource   string       `json:"dataSource,omitempty"`
	Speeds       []SpeedPoint `json:"speeds,omitempty"`
	Warning      string       `json:"warning,omitempty"`
}

type OptionsFeature struct {
	ExpirationDate                   time.Time `json:"expirationDate"`
	ATMStrike                        float64   `json:"atmStrike"`
	StraddleExpectedMove             float64   `json:"straddleExpectedMove"`
	StraddleExpectedMovePercent      float64   `json:"straddleExpectedMovePercent"`
	AverageResolvedImpliedVolatility float64   `json:"averageResolvedImpliedVolatility,omitempty"`
	AverageRawImpliedVolatility      float64   `json:"averageRawImpliedVolatility,omitempty"`
	OneTradingDayExpectedMove        float64   `json:"oneTradingDayExpectedMove,omitempty"`
	OneTradingDayExpectedMovePercent float64   `json:"oneTradingDayExpectedMovePercent,omitempty"`
	TimeToExpirationYears            float64   `json:"timeToExpirationYears,omitempty"`
	UnderlyingDataAgeSeconds         int64     `json:"underlyingDataAgeSeconds,omitempty"`
	ContractTradeDataAgeSeconds      int64     `json:"contractTradeDataAgeSeconds,omitempty"`
	MethodWarning                    string    `json:"methodWarning,omitempty"`
}

type ReasonableRange struct {
	SelectedOneDayMove        float64    `json:"selectedOneDayMove"`
	SelectedOneDayMovePercent float64    `json:"selectedOneDayMovePercent"`
	SelectedSource            string     `json:"selectedSource"`
	PreviousCloseRange        *RangeBand `json:"previousCloseRange,omitempty"`
	OpenRange                 *RangeBand `json:"openRange,omitempty"`
	CurrentRemainingRange     *RangeBand `json:"currentRemainingRange,omitempty"`
	RegularMinutesRemaining   int        `json:"regularMinutesRemaining,omitempty"`
	Method                    string     `json:"method"`
	Warnings                  []string   `json:"warnings,omitempty"`
}

type RangeBand struct {
	Anchor      string  `json:"anchor"`
	AnchorPrice float64 `json:"anchorPrice"`
	Move1Sigma  float64 `json:"move1Sigma"`
	Low1Sigma   float64 `json:"low1Sigma"`
	High1Sigma  float64 `json:"high1Sigma"`
	Move2Sigma  float64 `json:"move2Sigma"`
	Low2Sigma   float64 `json:"low2Sigma"`
	High2Sigma  float64 `json:"high2Sigma"`
}
