package options

import "time"

const ProviderYahooMCP = "yahoo_mcp_options"

type ChainResponse struct {
	Symbol              string          `json:"symbol"`
	Provider            string          `json:"provider"`
	Source              string          `json:"source"`
	ReceivedAt          time.Time       `json:"receivedAt"`
	ExpirationDate      time.Time       `json:"expirationDate"`
	ExpiryBucket        string          `json:"expiryBucket,omitempty"`
	ExpirationDates     []time.Time     `json:"expirationDates"`
	Strikes             []float64       `json:"strikes,omitempty"`
	Underlying          UnderlyingQuote `json:"underlying"`
	Calls               []Contract      `json:"calls"`
	Puts                []Contract      `json:"puts"`
	Summary             ExpectedMove    `json:"summary"`
	ProviderWarning     string          `json:"providerWarning"`
	ReturnedNearStrikes int             `json:"returnedNearStrikes"`
	RawCounts           map[string]int  `json:"rawCounts,omitempty"`
}

type UnderlyingQuote struct {
	Symbol                string    `json:"symbol"`
	ShortName             string    `json:"shortName,omitempty"`
	Currency              string    `json:"currency,omitempty"`
	Exchange              string    `json:"exchange,omitempty"`
	MarketState           string    `json:"marketState,omitempty"`
	QuoteSourceName       string    `json:"quoteSourceName,omitempty"`
	RegularMarketPrice    float64   `json:"regularMarketPrice"`
	PostMarketPrice       float64   `json:"postMarketPrice,omitempty"`
	PreviousClose         float64   `json:"previousClose,omitempty"`
	Bid                   float64   `json:"bid,omitempty"`
	Ask                   float64   `json:"ask,omitempty"`
	RegularMarketTime     time.Time `json:"regularMarketTime,omitempty"`
	PostMarketTime        time.Time `json:"postMarketTime,omitempty"`
	ExchangeDataDelayedBy int       `json:"exchangeDataDelayedBy"`
	SourceInterval        int       `json:"sourceInterval"`
}

type Contract struct {
	ContractSymbol          string    `json:"contractSymbol"`
	Type                    string    `json:"type"`
	Strike                  float64   `json:"strike"`
	Currency                string    `json:"currency,omitempty"`
	LastPrice               float64   `json:"lastPrice"`
	Bid                     float64   `json:"bid"`
	Ask                     float64   `json:"ask"`
	Mid                     float64   `json:"mid"`
	IVInputPrice            float64   `json:"ivInputPrice"`
	IVInputSource           string    `json:"ivInputSource,omitempty"`
	Volume                  int       `json:"volume,omitempty"`
	OpenInterest            int       `json:"openInterest,omitempty"`
	ImpliedVolatility       float64   `json:"impliedVolatility"`
	ResolvedImpliedVol      float64   `json:"resolvedImpliedVolatility,omitempty"`
	ResolvedImpliedVolError string    `json:"resolvedImpliedVolatilityError,omitempty"`
	InTheMoney              bool      `json:"inTheMoney"`
	LastTradeDate           time.Time `json:"lastTradeDate,omitempty"`
	SpreadPercent           float64   `json:"spreadPercent,omitempty"`
	Quality                 string    `json:"quality,omitempty"`
	QualityWarning          string    `json:"qualityWarning,omitempty"`
	EligibleForRisk         bool      `json:"eligibleForRisk"`
}

type ExpectedMove struct {
	Spot                             float64   `json:"spot"`
	ATMStrike                        float64   `json:"atmStrike"`
	CallMid                          float64   `json:"callMid"`
	PutMid                           float64   `json:"putMid"`
	StraddleMid                      float64   `json:"straddleMid"`
	ExpectedMove                     float64   `json:"expectedMove"`
	ExpectedMovePercent              float64   `json:"expectedMovePercent"`
	RawCallImpliedVolatility         float64   `json:"rawCallImpliedVolatility"`
	RawPutImpliedVolatility          float64   `json:"rawPutImpliedVolatility"`
	AverageRawImpliedVolatility      float64   `json:"averageRawImpliedVolatility"`
	ResolvedCallImpliedVolatility    float64   `json:"resolvedCallImpliedVolatility,omitempty"`
	ResolvedPutImpliedVolatility     float64   `json:"resolvedPutImpliedVolatility,omitempty"`
	AverageResolvedImpliedVolatility float64   `json:"averageResolvedImpliedVolatility,omitempty"`
	OneTradingDayExpectedMove        float64   `json:"oneTradingDayExpectedMove,omitempty"`
	OneTradingDayExpectedMovePercent float64   `json:"oneTradingDayExpectedMovePercent,omitempty"`
	TimeToExpirationYears            float64   `json:"timeToExpirationYears,omitempty"`
	RiskFreeRate                     float64   `json:"riskFreeRate"`
	DividendYieldAssumption          float64   `json:"dividendYieldAssumption"`
	MostRecentContractTradeAt        time.Time `json:"mostRecentContractTradeAt,omitempty"`
	ContractTradeDataAgeSeconds      int64     `json:"contractTradeDataAgeSeconds,omitempty"`
	UnderlyingDataAgeSeconds         int64     `json:"underlyingDataAgeSeconds,omitempty"`
	Method                           string    `json:"method"`
	MethodWarning                    string    `json:"methodWarning"`
}
