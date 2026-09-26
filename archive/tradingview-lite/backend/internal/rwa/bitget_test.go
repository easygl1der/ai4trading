package rwa

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBitgetClientStockInfoAndKline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("Authorization header = %q", got)
		}

		var req jsonRPCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		switch req.Params.Name {
		case "rwa_stock_info":
			if got := req.Params.Arguments["ticker"]; got != "NVDAon" {
				t.Fatalf("stock info ticker = %v", got)
			}
			writeRPCJSON(t, w, `{
				"status": 0,
				"data": {
					"ticker": "NVDAon",
					"name": "NVIDIA",
					"cn_name": "英伟达",
					"data_source": "ondo",
					"status": "active",
					"market_status": "open",
					"market_status_code": "regular",
					"latest_price": 177.42,
					"latest_ticker_price": 177.40,
					"price_24h_change": 1.2,
					"price_24h_change_ratio": 0.0068,
					"high_price_24h": 180.0,
					"low_price_24h": 174.0,
					"volume_24h": 10,
					"volume_24h_usd": 1774.2,
					"tradable_sessions": ["premarket", "regular", "postmarket", "overnight", "offhours"],
					"order_book_depths": [1, 5, 10],
					"contract_pair_name": "NVDA/USD"
				}
			}`)
		case "rwa_kline":
			if got := req.Params.Arguments["contract"]; got != "NVDAon" {
				t.Fatalf("kline contract = %v", got)
			}
			writeRPCSSE(t, w, `{
				"status": 0,
				"data": {
					"hasVolume": false,
					"ssekey": "test-key",
					"isHaveMoreData": false,
					"limitOrders": [],
					"list": [
						{
							"ts": 1783767600,
							"open": 177.0,
							"high": 178.0,
							"low": 176.5,
							"close": 177.5,
							"volume": 0,
							"amount": 0,
							"txn": 0
						}
					]
				}
			}`)
		default:
			t.Fatalf("unexpected tool name %q", req.Params.Name)
		}
	}))
	defer server.Close()

	client := NewBitgetClient(server.URL, "test-token", map[string]string{"NVDA": "NVDAon"}, server.Client())

	info, err := client.StockInfo(context.Background(), "NVDA")
	if err != nil {
		t.Fatalf("StockInfo returned error: %v", err)
	}
	if info.Symbol != "NVDA" || info.Ticker != "NVDAon" || info.Provider != ProviderBitgetRWA {
		t.Fatalf("unexpected stock info identity: %+v", info)
	}
	if !contains(info.TradableSessions, "overnight") {
		t.Fatalf("expected overnight session in %+v", info.TradableSessions)
	}
	if info.ProviderWarning == "" {
		t.Fatal("expected provider warning for RWA order book depths")
	}

	kline, err := client.Kline(context.Background(), "NVDA", "1m", 1)
	if err != nil {
		t.Fatalf("Kline returned error: %v", err)
	}
	if kline.Count != 1 || kline.HasVolume {
		t.Fatalf("unexpected kline metadata: %+v", kline)
	}
	if len(kline.Bars) != 1 || kline.Bars[0].Close != 177.5 {
		t.Fatalf("unexpected bars: %+v", kline.Bars)
	}
}

func writeRPCJSON(t *testing.T, w http.ResponseWriter, content string) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(rpcEnvelope(t, content))
}

func writeRPCSSE(t *testing.T, w http.ResponseWriter, content string) {
	t.Helper()
	w.Header().Set("Content-Type", "text/event-stream")
	payload, err := json.Marshal(rpcEnvelope(t, content))
	if err != nil {
		t.Fatalf("marshal rpc envelope: %v", err)
	}
	_, _ = w.Write([]byte("event: message\n"))
	_, _ = w.Write([]byte("data: "))
	_, _ = w.Write(payload)
	_, _ = w.Write([]byte("\n\n"))
}

func rpcEnvelope(t *testing.T, content string) map[string]any {
	t.Helper()
	var compact any
	if err := json.Unmarshal([]byte(content), &compact); err != nil {
		t.Fatalf("test content is invalid json: %v", err)
	}
	normalized, err := json.Marshal(compact)
	if err != nil {
		t.Fatalf("normalize content: %v", err)
	}
	return map[string]any{
		"jsonrpc": "2.0",
		"id":      "test",
		"result": map[string]any{
			"content": []map[string]string{
				{"type": "text", "text": string(normalized)},
			},
			"isError": false,
		},
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
