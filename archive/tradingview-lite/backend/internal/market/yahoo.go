package market

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const providerName = "yahoo_chart"

var validSymbol = regexp.MustCompile(`^[A-Z0-9.\-=^]+$`)

type YahooClient struct {
	httpClient *http.Client
	baseURL    string
}

func NewYahooClient(httpClient *http.Client) *YahooClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}

	return &YahooClient{
		httpClient: httpClient,
		baseURL:    "https://query1.finance.yahoo.com/v8/finance/chart",
	}
}

func (c *YahooClient) FetchQuote(ctx context.Context, symbol string) (*Quote, error) {
	chart, err := c.fetchChart(ctx, symbol, "1d", "1m", true)
	if err != nil {
		return nil, err
	}

	bars, dropped := finalizedBars(chart, "1m")
	intrabar := lastCompletePoint(chart)
	now := time.Now().UTC()
	marketTime := unixTime(chart.Meta.RegularMarketTime)
	if marketTime.IsZero() && intrabar != nil {
		marketTime = intrabar.Time
	}
	if intrabar != nil && !intrabar.Time.Before(marketTime) {
		marketTime = intrabar.Time
	}

	price := chart.Meta.RegularMarketPrice
	if intrabar != nil && (price == 0 || !intrabar.Time.Before(marketTime)) {
		price = intrabar.Close
	}

	quote := &Quote{
		Symbol:              chart.Meta.Symbol,
		Provider:            providerName,
		Price:               price,
		RegularMarketTime:   marketTime,
		ReceivedAt:          now,
		DataAgeSeconds:      ageSeconds(now, marketTime),
		ExchangeName:        chart.Meta.ExchangeName,
		ExchangeTimezone:    chart.Meta.ExchangeTimezoneName,
		DataGranularity:     chart.Meta.DataGranularity,
		LastIntrabarPoint:   intrabar,
		DroppedNonFinalRows: dropped,
		ProviderMarketPrice: chart.Meta.RegularMarketPrice,
	}
	if len(bars) > 0 {
		quote.LastFinalizedBar = &bars[len(bars)-1]
	}
	if quote.DataAgeSeconds > 120 {
		quote.ProviderWarning = "provider data is older than 120 seconds"
	}

	return quote, nil
}

func (c *YahooClient) FetchBars(ctx context.Context, symbol, rangeValue, interval string, includePrePost bool) (*BarsResponse, error) {
	if rangeValue == "" {
		rangeValue = "1d"
	}
	if interval == "" {
		interval = "1m"
	}

	chart, err := c.fetchChart(ctx, symbol, rangeValue, interval, includePrePost)
	if err != nil {
		return nil, err
	}

	bars, dropped := finalizedBars(chart, interval)
	now := time.Now().UTC()
	marketTime := unixTime(chart.Meta.RegularMarketTime)
	if len(bars) > 0 && !bars[len(bars)-1].Time.Before(marketTime) {
		marketTime = bars[len(bars)-1].Time
	}

	return &BarsResponse{
		Symbol:              chart.Meta.Symbol,
		Provider:            providerName,
		Interval:            interval,
		Range:               rangeValue,
		ReceivedAt:          now,
		RegularMarketTime:   marketTime,
		DataAgeSeconds:      ageSeconds(now, marketTime),
		Count:               len(bars),
		DroppedNonFinalRows: dropped,
		Bars:                bars,
	}, nil
}

func (c *YahooClient) fetchChart(ctx context.Context, symbol, rangeValue, interval string, includePrePost bool) (*chartResult, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if !validSymbol.MatchString(symbol) {
		return nil, fmt.Errorf("invalid symbol %q", symbol)
	}

	endpoint := fmt.Sprintf("%s/%s", c.baseURL, url.PathEscape(symbol))
	query := url.Values{}
	query.Set("range", rangeValue)
	query.Set("interval", normalizeInterval(interval))
	query.Set("includePrePost", fmt.Sprintf("%t", includePrePost))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 tradingview-lite/0.1")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(io.LimitReader(res.Body, 10<<20))
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("yahoo chart returned HTTP %d: %s", res.StatusCode, string(body))
	}

	var payload yahooChartResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	if payload.Chart.Error != nil {
		return nil, fmt.Errorf("yahoo chart error %s: %s", payload.Chart.Error.Code, payload.Chart.Error.Description)
	}
	if len(payload.Chart.Result) == 0 {
		return nil, errors.New("yahoo chart returned no result")
	}

	return &payload.Chart.Result[0], nil
}

func finalizedBars(chart *chartResult, interval string) ([]Bar, int) {
	if len(chart.Indicators.Quote) == 0 {
		return nil, len(chart.Timestamp)
	}

	series := chart.Indicators.Quote[0]
	count := min(
		len(chart.Timestamp),
		len(series.Open),
		len(series.High),
		len(series.Low),
		len(series.Close),
		len(series.Volume),
	)

	bars := make([]Bar, 0, count)
	for i := 0; i < count; i++ {
		if series.Open[i] == nil || series.High[i] == nil || series.Low[i] == nil || series.Close[i] == nil || series.Volume[i] == nil {
			continue
		}
		if !isFinalizedTimestamp(chart.Timestamp[i], interval) {
			continue
		}
		if isQuoteOnlyPoint(*series.Open[i], *series.High[i], *series.Low[i], *series.Close[i], *series.Volume[i]) {
			continue
		}
		bars = append(bars, Bar{
			Time:   unixTime(chart.Timestamp[i]),
			Open:   *series.Open[i],
			High:   *series.High[i],
			Low:    *series.Low[i],
			Close:  *series.Close[i],
			Volume: *series.Volume[i],
		})
	}

	return bars, len(chart.Timestamp) - len(bars)
}

func lastCompletePoint(chart *chartResult) *Bar {
	if len(chart.Indicators.Quote) == 0 {
		return nil
	}

	series := chart.Indicators.Quote[0]
	count := min(
		len(chart.Timestamp),
		len(series.Open),
		len(series.High),
		len(series.Low),
		len(series.Close),
	)

	for i := count - 1; i >= 0; i-- {
		if series.Open[i] == nil || series.High[i] == nil || series.Low[i] == nil || series.Close[i] == nil {
			continue
		}
		volume := 0.0
		if i < len(series.Volume) && series.Volume[i] != nil {
			volume = *series.Volume[i]
		}
		return &Bar{
			Time:   unixTime(chart.Timestamp[i]),
			Open:   *series.Open[i],
			High:   *series.High[i],
			Low:    *series.Low[i],
			Close:  *series.Close[i],
			Volume: volume,
		}
	}

	return nil
}

func isQuoteOnlyPoint(open, high, low, close, volume float64) bool {
	return volume == 0 && open == high && high == low && low == close
}

func normalizeInterval(interval string) string {
	if interval == "1h" {
		return "60m"
	}
	return interval
}

func isFinalizedTimestamp(ts int64, interval string) bool {
	seconds := intervalSeconds(normalizeInterval(interval))
	if seconds == 0 {
		return true
	}
	return ts%seconds == 0
}

func intervalSeconds(interval string) int64 {
	switch interval {
	case "1m":
		return 60
	case "2m":
		return 120
	case "5m":
		return 300
	case "15m":
		return 900
	case "30m":
		return 1800
	case "60m":
		return 3600
	default:
		return 0
	}
}

func unixTime(ts int64) time.Time {
	if ts <= 0 {
		return time.Time{}
	}
	return time.Unix(ts, 0).UTC()
}

func ageSeconds(now, eventTime time.Time) int64 {
	if eventTime.IsZero() {
		return 0
	}
	return int64(now.Sub(eventTime).Seconds())
}

func min(values ...int) int {
	if len(values) == 0 {
		return 0
	}
	out := values[0]
	for _, value := range values[1:] {
		if value < out {
			out = value
		}
	}
	return out
}

type yahooChartResponse struct {
	Chart struct {
		Result []chartResult `json:"result"`
		Error  *yahooError   `json:"error"`
	} `json:"chart"`
}

type yahooError struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type chartResult struct {
	Meta       chartMeta       `json:"meta"`
	Timestamp  []int64         `json:"timestamp"`
	Indicators chartIndicators `json:"indicators"`
}

type chartMeta struct {
	Symbol               string  `json:"symbol"`
	ExchangeName         string  `json:"exchangeName"`
	ExchangeTimezoneName string  `json:"exchangeTimezoneName"`
	RegularMarketTime    int64   `json:"regularMarketTime"`
	RegularMarketPrice   float64 `json:"regularMarketPrice"`
	DataGranularity      string  `json:"dataGranularity"`
}

type chartIndicators struct {
	Quote []chartQuote `json:"quote"`
}

type chartQuote struct {
	Open   []*float64 `json:"open"`
	High   []*float64 `json:"high"`
	Low    []*float64 `json:"low"`
	Close  []*float64 `json:"close"`
	Volume []*float64 `json:"volume"`
}
