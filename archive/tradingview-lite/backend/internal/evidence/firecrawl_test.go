package evidence

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFirecrawlClientSearchPreservesNewsContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/news/search" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) == "" || r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("request body or content type missing")
		}
		_, _ = w.Write([]byte(`{
  "success": true,
  "data": [{"title":"NVDA filing","url":"https://www.sec.gov/Archives/x","provider":"rss_feeds","publishedAt":"2026-07-10T14:00:00Z","discoveredAt":"2026-07-10T14:01:00Z","freshnessConfidence":"high","tickers":["NVDA"],"eventType":"filing"}],
  "metadata": {"providerRoute":"single_stock","strictTicker":true,"returnedCounts":{"rss_feeds":1}}
}`))
	}))
	defer server.Close()

	client := NewFirecrawlClient(server.URL+"/v1/news/search", "test-token", server.Client())
	response, err := client.Search(context.Background(), SearchRequest{Query: "NVDA latest news", Tickers: []string{"NVDA"}, StrictTicker: true, Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if !response.Success || response.Metadata.ProviderRoute != "single_stock" || !response.Metadata.StrictTicker || len(response.Data) != 1 {
		t.Fatalf("unexpected response: %+v", response)
	}
	if got := SourceTier(SourceDomain(response.Data[0].URL, ""), nil); got != "T1_OFFICIAL" {
		t.Fatalf("source tier = %q", got)
	}
}

func TestFirecrawlClientClassifiesRateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte("rate limit exceeded"))
	}))
	defer server.Close()

	client := NewFirecrawlClient(server.URL, "", server.Client())
	_, err := client.Search(context.Background(), SearchRequest{Query: "NVDA latest news"})
	if err == nil || !IsRateLimited(err) {
		t.Fatalf("expected rate limited error, got %v", err)
	}
}
