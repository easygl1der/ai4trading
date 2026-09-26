package options

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestYahooMCPClientFetchChain(t *testing.T) {
	previousNow := nowUTC
	nowUTC = func() time.Time { return time.Date(2026, time.July, 10, 20, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { nowUTC = previousNow })
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("Authorization header = %q", got)
		}
		var req jsonRPCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Params.Name != "options" {
			t.Fatalf("tool name = %q", req.Params.Name)
		}
		if got := req.Params.Arguments["symbol"]; got != "NVDA" {
			t.Fatalf("symbol = %v", got)
		}
		writeOptionsRPCJSON(t, w, `{
			"result": {
				"underlyingSymbol": "NVDA",
				"expirationDates": ["2026-07-13T00:00:00.000Z"],
				"strikes": [205, 210, 215],
				"quote": {
					"symbol": "NVDA",
					"shortName": "NVIDIA",
					"currency": "USD",
					"exchange": "NMS",
					"marketState": "POSTPOST",
					"quoteSourceName": "Nasdaq Real Time Price",
					"regularMarketPrice": 210.96,
					"regularMarketPreviousClose": 202.78,
					"bid": 210.13,
					"ask": 210.39,
					"regularMarketTime": "2026-07-10T20:00:00.000Z",
					"postMarketTime": "2026-07-10T23:59:59.000Z",
					"exchangeDataDelayedBy": 0,
					"sourceInterval": 15
				},
				"options": [
					{
						"expirationDate": "2026-07-13T00:00:00.000Z",
						"calls": [
							{"contractSymbol":"NVDA260713C00205000","strike":205,"currency":"USD","lastPrice":6.25,"bid":6.2,"ask":6.35,"volume":46734,"openInterest":14104,"impliedVolatility":0.2915,"inTheMoney":true,"lastTradeDate":"2026-07-10T19:59:57.000Z"},
							{"contractSymbol":"NVDA260713C00210000","strike":210,"currency":"USD","lastPrice":2.37,"bid":2.37,"ask":2.44,"volume":105254,"openInterest":12437,"impliedVolatility":0.2524,"inTheMoney":true,"lastTradeDate":"2026-07-10T19:59:59.000Z"},
							{"contractSymbol":"NVDA260713C00215000","strike":215,"currency":"USD","lastPrice":0.5,"bid":0.49,"ask":0.51,"volume":48465,"openInterest":2691,"impliedVolatility":0.2436,"inTheMoney":false,"lastTradeDate":"2026-07-10T19:59:59.000Z"}
						],
						"puts": [
							{"contractSymbol":"NVDA260713P00205000","strike":205,"currency":"USD","lastPrice":0.3,"bid":0.28,"ask":0.3,"volume":41451,"openInterest":5194,"impliedVolatility":0.2690,"inTheMoney":false,"lastTradeDate":"2026-07-10T19:59:58.000Z"},
							{"contractSymbol":"NVDA260713P00210000","strike":210,"currency":"USD","lastPrice":1.4,"bid":1.41,"ask":1.45,"volume":22511,"openInterest":155,"impliedVolatility":0.2485,"inTheMoney":false,"lastTradeDate":"2026-07-10T19:59:59.000Z"},
							{"contractSymbol":"NVDA260713P00215000","strike":215,"currency":"USD","lastPrice":4.63,"bid":4.45,"ask":4.6,"volume":4812,"openInterest":82,"impliedVolatility":0.2529,"inTheMoney":true,"lastTradeDate":"2026-07-10T19:59:42.000Z"}
						]
					}
				]
			}
		}`)
	}))
	defer server.Close()

	client := NewYahooMCPClient(server.URL, "test-token", 0.045, server.Client())
	chain, err := client.FetchChain(context.Background(), "NVDA", 2)
	if err != nil {
		t.Fatalf("FetchChain returned error: %v", err)
	}
	if chain.Symbol != "NVDA" || chain.Provider != ProviderYahooMCP {
		t.Fatalf("unexpected chain identity: %+v", chain)
	}
	if chain.Summary.ATMStrike != 210 {
		t.Fatalf("ATM strike = %v", chain.Summary.ATMStrike)
	}
	if chain.Summary.StraddleMid < 3.83 || chain.Summary.StraddleMid > 3.84 {
		t.Fatalf("straddle mid = %v", chain.Summary.StraddleMid)
	}
	if chain.Summary.AverageResolvedImpliedVolatility <= 0 {
		t.Fatalf("expected resolved IV in summary: %+v", chain.Summary)
	}
	if chain.Calls[0].IVInputPrice <= 0 || chain.Calls[0].IVInputSource == "" {
		t.Fatalf("expected IV input metadata: %+v", chain.Calls[0])
	}
	if len(chain.Calls) != 2 || len(chain.Puts) != 2 {
		t.Fatalf("near filter failed: calls=%d puts=%d", len(chain.Calls), len(chain.Puts))
	}
	if chain.ProviderWarning == "" || chain.Summary.MethodWarning == "" {
		t.Fatal("expected provider and method warnings")
	}
}

func TestExpirationMarketCloseKeepsYahooCalendarDate(t *testing.T) {
	expiration := time.Date(2026, time.July, 13, 0, 0, 0, 0, time.UTC)
	got := expirationMarketClose(expiration)
	want := time.Date(2026, time.July, 13, 20, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("expiration market close = %s, want %s", got, want)
	}
}

func writeOptionsRPCJSON(t *testing.T, w http.ResponseWriter, content string) {
	t.Helper()
	var compact any
	if err := json.Unmarshal([]byte(content), &compact); err != nil {
		t.Fatalf("test content is invalid json: %v", err)
	}
	normalized, err := json.Marshal(compact)
	if err != nil {
		t.Fatalf("normalize content: %v", err)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"jsonrpc": "2.0",
		"id":      "test",
		"result": map[string]any{
			"content": []map[string]string{
				{"type": "text", "text": string(normalized)},
			},
			"isError": false,
		},
	})
}
