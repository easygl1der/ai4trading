package rwa

import "time"

const ProviderBitgetRWA = "bitget_rwa"

type StockInfo struct {
	Symbol              string    `json:"symbol"`
	Ticker              string    `json:"ticker"`
	Name                string    `json:"name,omitempty"`
	CNName              string    `json:"cnName,omitempty"`
	Provider            string    `json:"provider"`
	DataSource          string    `json:"dataSource,omitempty"`
	Status              string    `json:"status,omitempty"`
	MarketStatus        string    `json:"marketStatus,omitempty"`
	MarketStatusCode    string    `json:"marketStatusCode,omitempty"`
	MarketNextCode      string    `json:"marketNextCode,omitempty"`
	LatestPrice         float64   `json:"latestPrice"`
	LatestTickerPrice   float64   `json:"latestTickerPrice"`
	Price24hChange      float64   `json:"price24hChange"`
	Price24hChangeRatio float64   `json:"price24hChangeRatio"`
	HighPrice24h        float64   `json:"highPrice24h"`
	LowPrice24h         float64   `json:"lowPrice24h"`
	Volume24h           float64   `json:"volume24h"`
	Volume24hUSD        float64   `json:"volume24hUsd"`
	TradableSessions    []string  `json:"tradableSessions,omitempty"`
	OrderBookDepths     []int     `json:"orderBookDepths,omitempty"`
	ContractPairName    string    `json:"contractPairName,omitempty"`
	ReceivedAt          time.Time `json:"receivedAt"`
	ProviderWarning     string    `json:"providerWarning,omitempty"`
}

type KlineResponse struct {
	Symbol          string    `json:"symbol"`
	Ticker          string    `json:"ticker"`
	Provider        string    `json:"provider"`
	Period          string    `json:"period"`
	ReceivedAt      time.Time `json:"receivedAt"`
	Count           int       `json:"count"`
	HasVolume       bool      `json:"hasVolume"`
	SSEKey          string    `json:"sseKey,omitempty"`
	IsHaveMoreData  bool      `json:"isHaveMoreData"`
	LimitOrderCount int       `json:"limitOrderCount"`
	Bars            []Bar     `json:"bars"`
}

type Bar struct {
	Time             time.Time `json:"time"`
	Timestamp        int64     `json:"timestamp"`
	Open             float64   `json:"open"`
	High             float64   `json:"high"`
	Low              float64   `json:"low"`
	Close            float64   `json:"close"`
	Volume           float64   `json:"volume"`
	Amount           float64   `json:"amount"`
	TransactionCount float64   `json:"transactionCount"`
	BuyVolume        float64   `json:"buyVolume"`
	SellVolume       float64   `json:"sellVolume"`
	BuyAmount        float64   `json:"buyAmount"`
	SellAmount       float64   `json:"sellAmount"`
	UserBuyAmount    float64   `json:"userBuyAmount"`
	UserSellAmount   float64   `json:"userSellAmount"`
	UserBuyVolume    float64   `json:"userBuyVolume"`
	UserSellVolume   float64   `json:"userSellVolume"`
	UserAvgBuyPrice  float64   `json:"userAvgBuyPrice"`
	UserAvgSellPrice float64   `json:"userAvgSellPrice"`
}
