package rwa

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type BitgetClient struct {
	endpoint   string
	token      string
	symbolMap  map[string]string
	httpClient *http.Client
}

func NewBitgetClient(endpoint, token string, symbolMap map[string]string, httpClient *http.Client) *BitgetClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	if symbolMap == nil {
		symbolMap = map[string]string{}
	}
	return &BitgetClient{
		endpoint:   strings.TrimSpace(endpoint),
		token:      strings.TrimSpace(token),
		symbolMap:  symbolMap,
		httpClient: httpClient,
	}
}

func (c *BitgetClient) Enabled() bool {
	return c != nil && c.endpoint != ""
}

func (c *BitgetClient) ResolveTicker(symbol string) string {
	clean := strings.TrimSpace(symbol)
	upper := strings.ToUpper(clean)
	if mapped := c.symbolMap[upper]; mapped != "" {
		return mapped
	}
	if strings.HasSuffix(strings.ToLower(clean), "on") {
		return upper[:len(upper)-2] + "on"
	}
	return upper + "on"
}

func (c *BitgetClient) StockInfo(ctx context.Context, symbol string) (*StockInfo, error) {
	ticker := c.ResolveTicker(symbol)
	var payload stockInfoPayload
	if err := c.callTool(ctx, "rwa_stock_info", map[string]any{"ticker": ticker}, &payload); err != nil {
		return nil, err
	}
	if payload.Status != 0 {
		return nil, fmt.Errorf("bitget rwa_stock_info returned status %d", payload.Status)
	}
	now := time.Now().UTC()
	out := &StockInfo{
		Symbol:              normalizeSymbol(symbol, ticker),
		Ticker:              payload.Data.Ticker,
		Name:                payload.Data.Name,
		CNName:              payload.Data.CNName,
		Provider:            ProviderBitgetRWA,
		DataSource:          payload.Data.DataSource,
		Status:              payload.Data.Status,
		MarketStatus:        payload.Data.MarketStatus,
		MarketStatusCode:    payload.Data.MarketStatusCode,
		MarketNextCode:      payload.Data.MarketNextCode,
		LatestPrice:         payload.Data.LatestPrice,
		LatestTickerPrice:   payload.Data.LatestTickerPrice,
		Price24hChange:      payload.Data.Price24hChange,
		Price24hChangeRatio: payload.Data.Price24hChangeRatio,
		HighPrice24h:        payload.Data.HighPrice24h,
		LowPrice24h:         payload.Data.LowPrice24h,
		Volume24h:           payload.Data.Volume24h,
		Volume24hUSD:        payload.Data.Volume24hUSD,
		TradableSessions:    payload.Data.TradableSessions,
		OrderBookDepths:     payload.Data.OrderBookDepths,
		ContractPairName:    payload.Data.ContractPairName,
		ReceivedAt:          now,
	}
	if len(out.OrderBookDepths) > 0 {
		out.ProviderWarning = "orderBookDepths describes RWA provider depth options; it is not Nasdaq/NYSE official Level 2 book data"
	}
	return out, nil
}

func (c *BitgetClient) Kline(ctx context.Context, symbol, period string, size int) (*KlineResponse, error) {
	ticker := c.ResolveTicker(symbol)
	if period == "" {
		period = "1m"
	}
	if size <= 0 {
		size = 60
	}
	if size > 1440 {
		size = 1440
	}
	var payload klinePayload
	args := map[string]any{
		"chain":    "rwa",
		"contract": ticker,
		"period":   period,
		"size":     size,
	}
	if err := c.callTool(ctx, "rwa_kline", args, &payload); err != nil {
		return nil, err
	}
	if payload.Status != 0 {
		return nil, fmt.Errorf("bitget rwa_kline returned status %d", payload.Status)
	}
	bars := make([]Bar, 0, len(payload.Data.List))
	for _, row := range payload.Data.List {
		bars = append(bars, Bar{
			Time:             time.Unix(row.Timestamp, 0).UTC(),
			Timestamp:        row.Timestamp,
			Open:             row.Open,
			High:             row.High,
			Low:              row.Low,
			Close:            row.Close,
			Volume:           row.Volume,
			Amount:           row.Amount,
			TransactionCount: row.TransactionCount,
			BuyVolume:        row.BuyVolume,
			SellVolume:       row.SellVolume,
			BuyAmount:        row.BuyAmount,
			SellAmount:       row.SellAmount,
			UserBuyAmount:    row.UserBuyAmount,
			UserSellAmount:   row.UserSellAmount,
			UserBuyVolume:    row.UserBuyVolume,
			UserSellVolume:   row.UserSellVolume,
			UserAvgBuyPrice:  row.UserAvgBuyPrice,
			UserAvgSellPrice: row.UserAvgSellPrice,
		})
	}
	return &KlineResponse{
		Symbol:          normalizeSymbol(symbol, ticker),
		Ticker:          ticker,
		Provider:        ProviderBitgetRWA,
		Period:          period,
		ReceivedAt:      time.Now().UTC(),
		Count:           len(bars),
		HasVolume:       payload.Data.HasVolume,
		SSEKey:          payload.Data.SSEKey,
		IsHaveMoreData:  payload.Data.IsHaveMoreData,
		LimitOrderCount: len(payload.Data.LimitOrders),
		Bars:            bars,
	}, nil
}

func (c *BitgetClient) callTool(ctx context.Context, name string, args map[string]any, target any) error {
	if !c.Enabled() {
		return errors.New("bitget rwa client is not configured")
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
	raw, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("bitget mcp returned HTTP %d: %s", res.StatusCode, string(raw))
	}
	jsonBody := extractStreamableJSON(raw)
	var rpc jsonRPCResponse
	if err := json.Unmarshal(jsonBody, &rpc); err != nil {
		return fmt.Errorf("decode bitget mcp response: %w", err)
	}
	if rpc.Error != nil {
		return fmt.Errorf("bitget mcp error %d: %s", rpc.Error.Code, rpc.Error.Message)
	}
	if len(rpc.Result.Content) == 0 {
		return errors.New("bitget mcp returned no content")
	}
	if err := json.Unmarshal([]byte(rpc.Result.Content[0].Text), target); err != nil {
		return fmt.Errorf("decode bitget tool content: %w", err)
	}
	return nil
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

func normalizeSymbol(symbol, ticker string) string {
	clean := strings.ToUpper(strings.TrimSpace(symbol))
	if clean != "" && !strings.HasSuffix(strings.ToLower(clean), "on") {
		return clean
	}
	return strings.TrimSuffix(strings.ToUpper(ticker), "ON")
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

type stockInfoPayload struct {
	Data   stockInfoData `json:"data"`
	Status int           `json:"status"`
}

type stockInfoData struct {
	Ticker              string   `json:"ticker"`
	Name                string   `json:"name"`
	CNName              string   `json:"cn_name"`
	Status              string   `json:"status"`
	MarketStatus        string   `json:"market_status"`
	MarketStatusCode    string   `json:"market_status_code"`
	MarketNextCode      string   `json:"market_next_code"`
	DataSource          string   `json:"data_source"`
	ContractPairName    string   `json:"contract_pair_name"`
	LatestPrice         float64  `json:"latest_price"`
	LatestTickerPrice   float64  `json:"latest_ticker_price"`
	Price24hChange      float64  `json:"price_24h_change"`
	Price24hChangeRatio float64  `json:"price_24h_change_ratio"`
	HighPrice24h        float64  `json:"high_price_24h"`
	LowPrice24h         float64  `json:"low_price_24h"`
	Volume24h           float64  `json:"volume_24h"`
	Volume24hUSD        float64  `json:"volume_24h_usd"`
	TradableSessions    []string `json:"tradable_sessions"`
	OrderBookDepths     []int    `json:"order_book_depths"`
}

type klinePayload struct {
	Data   klineData `json:"data"`
	Status int       `json:"status"`
}

type klineData struct {
	List           []klineRow `json:"list"`
	HasVolume      bool       `json:"hasVolume"`
	SSEKey         string     `json:"ssekey"`
	IsHaveMoreData bool       `json:"isHaveMoreData"`
	LimitOrders    []any      `json:"limitOrders"`
}

type klineRow struct {
	Timestamp        int64   `json:"ts"`
	High             float64 `json:"high"`
	Low              float64 `json:"low"`
	Open             float64 `json:"open"`
	Close            float64 `json:"close"`
	Volume           float64 `json:"volume"`
	Amount           float64 `json:"amount"`
	TransactionCount float64 `json:"txn"`
	BuyVolume        float64 `json:"buyVolume"`
	SellVolume       float64 `json:"sellVolume"`
	BuyAmount        float64 `json:"buyAmount"`
	SellAmount       float64 `json:"sellAmount"`
	UserBuyAmount    float64 `json:"userBuyAmount"`
	UserSellAmount   float64 `json:"userSellAmount"`
	UserBuyVolume    float64 `json:"userBuyVolume"`
	UserSellVolume   float64 `json:"userSellVolume"`
	UserAvgBuyPrice  float64 `json:"userAvgBuyPrice"`
	UserAvgSellPrice float64 `json:"userAvgSellPrice"`
}
