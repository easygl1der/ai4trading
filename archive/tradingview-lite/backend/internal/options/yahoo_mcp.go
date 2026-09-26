package options

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
)

var validOptionSymbol = regexp.MustCompile(`^[A-Z0-9.\-=^]+$`)
var nowUTC = func() time.Time { return time.Now().UTC() }

type YahooMCPClient struct {
	endpoint     string
	token        string
	riskFreeRate float64
	httpClient   *http.Client
}

func NewYahooMCPClient(endpoint, token string, riskFreeRate float64, httpClient *http.Client) *YahooMCPClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	if riskFreeRate < 0 {
		riskFreeRate = 0
	}
	return &YahooMCPClient{
		endpoint:     strings.TrimSpace(endpoint),
		token:        strings.TrimSpace(token),
		riskFreeRate: riskFreeRate,
		httpClient:   httpClient,
	}
}

func (c *YahooMCPClient) Enabled() bool {
	return c != nil && c.endpoint != ""
}

func (c *YahooMCPClient) FetchChain(ctx context.Context, symbol string, nearStrikes int) (*ChainResponse, error) {
	return c.fetchChain(ctx, symbol, nearStrikes, nil)
}

// FetchExpiryBuckets returns explicitly labelled nearest-valid, 7D, and 30D chains.
// The Yahoo endpoint remains an unofficial source, so callers must retain quality warnings.
func (c *YahooMCPClient) FetchExpiryBuckets(ctx context.Context, symbol string, nearStrikes int) ([]*ChainResponse, error) {
	base, err := c.fetchChain(ctx, symbol, nearStrikes, nil)
	if err != nil {
		return nil, err
	}
	targets := selectExpiryBuckets(nowUTC(), base.ExpirationDates)
	chains := make([]*ChainResponse, 0, len(targets))
	for bucket, expiry := range targets {
		chain, err := c.fetchChain(ctx, symbol, nearStrikes, &expiry)
		if err != nil {
			return nil, fmt.Errorf("%s %s: %w", symbol, bucket, err)
		}
		chain.ExpiryBucket = bucket
		chains = append(chains, chain)
	}
	return chains, nil
}

func (c *YahooMCPClient) fetchChain(ctx context.Context, symbol string, nearStrikes int, expiry *time.Time) (*ChainResponse, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if !validOptionSymbol.MatchString(symbol) {
		return nil, fmt.Errorf("invalid symbol %q", symbol)
	}
	if nearStrikes <= 0 {
		nearStrikes = 20
	}
	if nearStrikes > 200 {
		nearStrikes = 200
	}

	var raw yahooOptionsData
	queryOptions := map[string]any{}
	if expiry != nil {
		queryOptions["date"] = expirationMarketClose(*expiry).Unix()
	}
	args := map[string]any{
		"symbol":        symbol,
		"queryOptions":  queryOptions,
		"moduleOptions": map[string]any{"validateResult": false},
	}
	if err := c.callTool(ctx, "options", args, &raw); err != nil {
		return nil, err
	}
	return normalizeYahooOptions(raw, nearStrikes, c.riskFreeRate)
}

func (c *YahooMCPClient) callTool(ctx context.Context, name string, args map[string]any, target *yahooOptionsData) error {
	if !c.Enabled() {
		return errors.New("yahoo options provider is not configured")
	}
	body, err := json.Marshal(jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      fmt.Sprintf("%d", time.Now().UTC().UnixNano()),
		Method:  "tools/call",
		Params: toolCallParams{
			Name:      name,
			Arguments: args,
		},
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("User-Agent", "tradingview-lite-risk-agent/0.1")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, 32<<20))
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("yahoo mcp returned HTTP %d: %s", res.StatusCode, string(raw))
	}
	jsonBody := extractStreamableJSON(raw)
	var rpc jsonRPCResponse
	if err := json.Unmarshal(jsonBody, &rpc); err != nil {
		return fmt.Errorf("decode yahoo mcp response: %w", err)
	}
	if rpc.Error != nil {
		return fmt.Errorf("yahoo mcp error %d: %s", rpc.Error.Code, rpc.Error.Message)
	}
	if rpc.Result == nil || len(rpc.Result.Content) == 0 {
		return errors.New("yahoo mcp returned no content")
	}

	text := []byte(rpc.Result.Content[0].Text)
	var wrapped yahooMCPToolOutput
	if err := json.Unmarshal(text, &wrapped); err == nil && wrapped.Result != nil {
		text = wrapped.Result
	}
	if err := json.Unmarshal(text, target); err != nil {
		return fmt.Errorf("decode yahoo options content: %w", err)
	}
	return nil
}

func normalizeYahooOptions(raw yahooOptionsData, nearStrikes int, riskFreeRate float64) (*ChainResponse, error) {
	if raw.UnderlyingSymbol == "" {
		return nil, errors.New("yahoo options response has no underlying symbol")
	}
	if len(raw.Options) == 0 {
		return nil, errors.New("yahoo options response has no option chains")
	}
	now := nowUTC()
	chain := raw.Options[0]
	spot := raw.Quote.RegularMarketPrice
	if spot == 0 {
		spot = raw.Quote.PostMarketPrice
	}
	if spot == 0 {
		spot = midpoint(raw.Quote.Bid, raw.Quote.Ask, 0)
	}

	expirationDate := parseYahooTime(chain.ExpirationDate)
	timeToExpiry := timeToExpirationYears(now, expirationDate)
	calls := normalizeContracts("call", chain.Calls, spot, timeToExpiry, riskFreeRate)
	puts := normalizeContracts("put", chain.Puts, spot, timeToExpiry, riskFreeRate)
	summary := expectedMoveSummary(spot, calls, puts, now, timeToExpiry, riskFreeRate)
	underlying := UnderlyingQuote{
		Symbol:                raw.Quote.Symbol,
		ShortName:             raw.Quote.ShortName,
		Currency:              raw.Quote.Currency,
		Exchange:              raw.Quote.Exchange,
		MarketState:           raw.Quote.MarketState,
		QuoteSourceName:       raw.Quote.QuoteSourceName,
		RegularMarketPrice:    raw.Quote.RegularMarketPrice,
		PostMarketPrice:       raw.Quote.PostMarketPrice,
		PreviousClose:         raw.Quote.RegularMarketPreviousClose,
		Bid:                   raw.Quote.Bid,
		Ask:                   raw.Quote.Ask,
		RegularMarketTime:     parseYahooTime(raw.Quote.RegularMarketTime),
		PostMarketTime:        parseYahooTime(raw.Quote.PostMarketTime),
		ExchangeDataDelayedBy: raw.Quote.ExchangeDataDelayedBy,
		SourceInterval:        raw.Quote.SourceInterval,
	}
	if !underlying.RegularMarketTime.IsZero() {
		summary.UnderlyingDataAgeSeconds = int64(now.Sub(underlying.RegularMarketTime).Seconds())
	}

	filteredCalls, filteredPuts := filterNearATM(calls, puts, summary.ATMStrike, nearStrikes)
	resp := &ChainResponse{
		Symbol:              raw.UnderlyingSymbol,
		Provider:            ProviderYahooMCP,
		Source:              "yahoo-finance2 options via Yahoo Finance MCP",
		ReceivedAt:          now,
		ExpirationDate:      expirationDate,
		ExpirationDates:     parseYahooTimes(raw.ExpirationDates),
		Strikes:             raw.Strikes,
		Underlying:          underlying,
		Calls:               filteredCalls,
		Puts:                filteredPuts,
		Summary:             summary,
		ReturnedNearStrikes: nearStrikes,
		RawCounts: map[string]int{
			"calls": len(calls),
			"puts":  len(puts),
		},
		ProviderWarning: "Yahoo options are a free/unofficial exploration source. Treat raw impliedVolatility as advisory; re-solve IV from bid/ask/mid before production risk decisions.",
	}
	return resp, nil
}

func normalizeContracts(kind string, rows []yahooContract, spot, timeToExpiry, riskFreeRate float64) []Contract {
	out := make([]Contract, 0, len(rows))
	for _, row := range rows {
		inputPrice, inputSource := ivPriceInput(row.Bid, row.Ask, row.LastPrice)
		resolvedIV, ivErr := solveImpliedVol(ivInput{
			Kind:          kind,
			Spot:          spot,
			Strike:        row.Strike,
			TimeYears:     timeToExpiry,
			RiskFreeRate:  riskFreeRate,
			DividendYield: 0,
			Price:         inputPrice,
		})
		ivErrText := ""
		if ivErr != nil {
			ivErrText = ivErr.Error()
		}
		mid := midpoint(row.Bid, row.Ask, row.LastPrice)
		spreadPct, quality, warning, eligible := optionQuality(row.Bid, row.Ask, mid, row.Volume, row.OpenInterest, parseYahooTime(row.LastTradeDate), nowUTC())
		out = append(out, Contract{
			ContractSymbol:          row.ContractSymbol,
			Type:                    kind,
			Strike:                  row.Strike,
			Currency:                row.Currency,
			LastPrice:               row.LastPrice,
			Bid:                     row.Bid,
			Ask:                     row.Ask,
			Mid:                     mid,
			IVInputPrice:            inputPrice,
			IVInputSource:           inputSource,
			Volume:                  row.Volume,
			OpenInterest:            row.OpenInterest,
			ImpliedVolatility:       row.ImpliedVolatility,
			ResolvedImpliedVol:      resolvedIV,
			ResolvedImpliedVolError: ivErrText,
			InTheMoney:              row.InTheMoney,
			LastTradeDate:           parseYahooTime(row.LastTradeDate),
			SpreadPercent:           spreadPct,
			Quality:                 quality,
			QualityWarning:          warning,
			EligibleForRisk:         eligible,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Strike < out[j].Strike })
	return out
}

func optionQuality(bid, ask, mid float64, volume, openInterest int, lastTradeAt, now time.Time) (float64, string, string, bool) {
	if bid <= 0 || ask <= 0 || mid <= 0 {
		return 0, "degraded", "missing_two_sided_quote", false
	}
	if bid > ask {
		return 0, "blocked", "crossed_quote", false
	}
	spreadPct := (ask - bid) / mid * 100
	if spreadPct > 25 {
		return spreadPct, "blocked", "bid_ask_spread_exceeds_25pct", false
	}
	if openInterest <= 0 && volume <= 0 {
		return spreadPct, "degraded", "no_volume_or_open_interest", false
	}
	if !lastTradeAt.IsZero() && now.Sub(lastTradeAt) > 15*time.Minute {
		return spreadPct, "degraded", "last_trade_older_than_15m", false
	}
	if spreadPct > 10 {
		return spreadPct, "degraded", "bid_ask_spread_exceeds_10pct", false
	}
	return spreadPct, "usable", "", true
}

func selectExpiryBuckets(now time.Time, dates []time.Time) map[string]time.Time {
	valid := make([]time.Time, 0, len(dates))
	for _, date := range dates {
		if expirationMarketClose(date).After(now) {
			valid = append(valid, date)
		}
	}
	sort.Slice(valid, func(i, j int) bool { return valid[i].Before(valid[j]) })
	result := map[string]time.Time{}
	if len(valid) == 0 {
		return result
	}
	result["0DTE_or_nearest_valid"] = valid[0]
	for _, target := range []struct {
		name string
		days int
	}{{"7D", 7}, {"30D", 30}} {
		want := now.AddDate(0, 0, target.days)
		best := valid[0]
		for _, date := range valid {
			if date.After(want) || date.Equal(want) {
				best = date
				break
			}
		}
		result[target.name] = best
	}
	return result
}

func expectedMoveSummary(spot float64, calls, puts []Contract, now time.Time, timeToExpiry, riskFreeRate float64) ExpectedMove {
	summary := ExpectedMove{
		Spot:                    spot,
		TimeToExpirationYears:   timeToExpiry,
		RiskFreeRate:            riskFreeRate,
		DividendYieldAssumption: 0,
		Method:                  "nearest ATM call mid plus nearest ATM put mid, plus Black-Scholes bisection IV from Yahoo bid/ask midpoint with last price fallback",
		MethodWarning:           "This is a European Black-Scholes approximation on free Yahoo option quotes. IV is annualized decimal; expected move is amplitude, not direction.",
	}
	if spot <= 0 {
		return summary
	}

	callByStrike := map[float64]Contract{}
	for _, call := range calls {
		if call.EligibleForRisk {
			callByStrike[call.Strike] = call
		}
	}
	var bestCall Contract
	var bestPut Contract
	bestDistance := math.MaxFloat64
	for _, put := range puts {
		if !put.EligibleForRisk {
			continue
		}
		call, ok := callByStrike[put.Strike]
		if !ok {
			continue
		}
		distance := math.Abs(put.Strike - spot)
		if distance < bestDistance {
			bestDistance = distance
			bestCall = call
			bestPut = put
		}
	}
	if bestDistance == math.MaxFloat64 {
		return summary
	}
	summary.ATMStrike = bestCall.Strike
	summary.CallMid = bestCall.Mid
	summary.PutMid = bestPut.Mid
	summary.StraddleMid = bestCall.Mid + bestPut.Mid
	summary.ExpectedMove = summary.StraddleMid
	summary.ExpectedMovePercent = summary.StraddleMid / spot * 100
	summary.RawCallImpliedVolatility = bestCall.ImpliedVolatility
	summary.RawPutImpliedVolatility = bestPut.ImpliedVolatility
	if bestCall.ImpliedVolatility > 0 && bestPut.ImpliedVolatility > 0 {
		summary.AverageRawImpliedVolatility = (bestCall.ImpliedVolatility + bestPut.ImpliedVolatility) / 2
	}
	summary.ResolvedCallImpliedVolatility = bestCall.ResolvedImpliedVol
	summary.ResolvedPutImpliedVolatility = bestPut.ResolvedImpliedVol
	if bestCall.ResolvedImpliedVol > 0 && bestPut.ResolvedImpliedVol > 0 {
		summary.AverageResolvedImpliedVolatility = (bestCall.ResolvedImpliedVol + bestPut.ResolvedImpliedVol) / 2
		summary.OneTradingDayExpectedMove, summary.OneTradingDayExpectedMovePercent = oneTradingDayMove(spot, summary.AverageResolvedImpliedVolatility)
	}
	if bestCall.LastTradeDate.After(bestPut.LastTradeDate) {
		summary.MostRecentContractTradeAt = bestCall.LastTradeDate
	} else {
		summary.MostRecentContractTradeAt = bestPut.LastTradeDate
	}
	if !summary.MostRecentContractTradeAt.IsZero() {
		summary.ContractTradeDataAgeSeconds = int64(now.Sub(summary.MostRecentContractTradeAt).Seconds())
	}
	return summary
}

func filterNearATM(calls, puts []Contract, atmStrike float64, nearStrikes int) ([]Contract, []Contract) {
	if atmStrike == 0 || nearStrikes >= len(calls)+len(puts) {
		return calls, puts
	}
	return filterContracts(calls, atmStrike, nearStrikes), filterContracts(puts, atmStrike, nearStrikes)
}

func filterContracts(rows []Contract, atmStrike float64, nearStrikes int) []Contract {
	if nearStrikes <= 0 || len(rows) <= nearStrikes {
		return rows
	}
	cp := append([]Contract(nil), rows...)
	sort.Slice(cp, func(i, j int) bool {
		di := math.Abs(cp[i].Strike - atmStrike)
		dj := math.Abs(cp[j].Strike - atmStrike)
		if di == dj {
			return cp[i].Strike < cp[j].Strike
		}
		return di < dj
	})
	cp = cp[:nearStrikes]
	sort.Slice(cp, func(i, j int) bool { return cp[i].Strike < cp[j].Strike })
	return cp
}

func midpoint(bid, ask, fallback float64) float64 {
	if bid > 0 && ask > 0 {
		return (bid + ask) / 2
	}
	if fallback > 0 {
		return fallback
	}
	if bid > 0 {
		return bid
	}
	return ask
}

func ivPriceInput(bid, ask, last float64) (float64, string) {
	if bid > 0 && ask > 0 {
		return (bid + ask) / 2, "bid_ask_mid"
	}
	if last > 0 {
		return last, "last_price_fallback"
	}
	if bid > 0 {
		return bid, "bid_only"
	}
	if ask > 0 {
		return ask, "ask_only"
	}
	return 0, "missing_price"
}

func timeToExpirationYears(now time.Time, expirationDate time.Time) float64 {
	expiry := expirationMarketClose(expirationDate)
	if expiry.IsZero() || !expiry.After(now) {
		return 0
	}
	return expiry.Sub(now).Hours() / (24 * 365)
}

func expirationMarketClose(expirationDate time.Time) time.Time {
	if expirationDate.IsZero() {
		return time.Time{}
	}
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return expirationDate
	}
	// Yahoo supplies expiration as an ISO calendar date at midnight UTC. It is
	// a date label, not a market-time instant, so converting it to New York
	// first can incorrectly shift it to the prior trading day.
	utc := expirationDate.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 16, 0, 0, 0, loc).UTC()
}

func parseYahooTimes(values []string) []time.Time {
	out := make([]time.Time, 0, len(values))
	for _, value := range values {
		parsed := parseYahooTime(value)
		if !parsed.IsZero() {
			out = append(out, parsed)
		}
	}
	return out
}

func parseYahooTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	if parsed, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return parsed.UTC()
	}
	return time.Time{}
}

func extractStreamableJSON(raw []byte) []byte {
	text := strings.TrimSpace(string(raw))
	if strings.HasPrefix(text, "{") {
		return raw
	}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if strings.HasPrefix(data, "{") {
				return []byte(data)
			}
		}
	}
	return raw
}

type jsonRPCRequest struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      string         `json:"id"`
	Method  string         `json:"method"`
	Params  toolCallParams `json:"params"`
}

type toolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type jsonRPCResponse struct {
	Result *struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	} `json:"result"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type yahooMCPToolOutput struct {
	Result json.RawMessage `json:"result"`
}

type yahooOptionsData struct {
	UnderlyingSymbol string             `json:"underlyingSymbol"`
	ExpirationDates  []string           `json:"expirationDates"`
	Strikes          []float64          `json:"strikes"`
	Quote            yahooQuote         `json:"quote"`
	Options          []yahooOptionChain `json:"options"`
}

type yahooQuote struct {
	Symbol                     string  `json:"symbol"`
	ShortName                  string  `json:"shortName"`
	Currency                   string  `json:"currency"`
	Exchange                   string  `json:"exchange"`
	MarketState                string  `json:"marketState"`
	QuoteSourceName            string  `json:"quoteSourceName"`
	RegularMarketPrice         float64 `json:"regularMarketPrice"`
	PostMarketPrice            float64 `json:"postMarketPrice"`
	RegularMarketPreviousClose float64 `json:"regularMarketPreviousClose"`
	Bid                        float64 `json:"bid"`
	Ask                        float64 `json:"ask"`
	RegularMarketTime          string  `json:"regularMarketTime"`
	PostMarketTime             string  `json:"postMarketTime"`
	ExchangeDataDelayedBy      int     `json:"exchangeDataDelayedBy"`
	SourceInterval             int     `json:"sourceInterval"`
}

type yahooOptionChain struct {
	ExpirationDate string          `json:"expirationDate"`
	Calls          []yahooContract `json:"calls"`
	Puts           []yahooContract `json:"puts"`
}

type yahooContract struct {
	ContractSymbol    string  `json:"contractSymbol"`
	Strike            float64 `json:"strike"`
	Currency          string  `json:"currency"`
	LastPrice         float64 `json:"lastPrice"`
	Bid               float64 `json:"bid"`
	Ask               float64 `json:"ask"`
	Volume            int     `json:"volume"`
	OpenInterest      int     `json:"openInterest"`
	ImpliedVolatility float64 `json:"impliedVolatility"`
	InTheMoney        bool    `json:"inTheMoney"`
	LastTradeDate     string  `json:"lastTradeDate"`
}
