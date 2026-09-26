package market

import "time"

type Bar struct {
	Time   time.Time `json:"time"`
	Open   float64   `json:"open"`
	High   float64   `json:"high"`
	Low    float64   `json:"low"`
	Close  float64   `json:"close"`
	Volume float64   `json:"volume"`
}

type Quote struct {
	Symbol              string    `json:"symbol"`
	Provider            string    `json:"provider"`
	Price               float64   `json:"price"`
	RegularMarketTime   time.Time `json:"regularMarketTime"`
	ReceivedAt          time.Time `json:"receivedAt"`
	DataAgeSeconds      int64     `json:"dataAgeSeconds"`
	ExchangeName        string    `json:"exchangeName"`
	ExchangeTimezone    string    `json:"exchangeTimezone"`
	DataGranularity     string    `json:"dataGranularity"`
	LastFinalizedBar    *Bar      `json:"lastFinalizedBar,omitempty"`
	LastIntrabarPoint   *Bar      `json:"lastIntrabarPoint,omitempty"`
	DroppedNonFinalRows int       `json:"droppedNonFinalRows"`
	ProviderWarning     string    `json:"providerWarning,omitempty"`
	ProviderMarketPrice float64   `json:"providerMarketPrice"`
}

type BarsResponse struct {
	Symbol              string    `json:"symbol"`
	Provider            string    `json:"provider"`
	Interval            string    `json:"interval"`
	Range               string    `json:"range"`
	ReceivedAt          time.Time `json:"receivedAt"`
	RegularMarketTime   time.Time `json:"regularMarketTime"`
	DataAgeSeconds      int64     `json:"dataAgeSeconds"`
	Count               int       `json:"count"`
	DroppedNonFinalRows int       `json:"droppedNonFinalRows"`
	Bars                []Bar     `json:"bars"`
}
