package evidence

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const ProviderFirecrawlNews = "firecrawl_news"

type SearchRequest struct {
	Query          string   `json:"query"`
	Tickers        []string `json:"tickers,omitempty"`
	Limit          int      `json:"limit"`
	MaxAgeMinutes  int      `json:"maxAgeMinutes"`
	StrictTicker   bool     `json:"strictTicker"`
	FetchContent   bool     `json:"fetchContent"`
	IncludeDomains []string `json:"includeDomains,omitempty"`
}

type NewsItem struct {
	Title               string          `json:"title"`
	URL                 string          `json:"url"`
	Source              string          `json:"source"`
	Provider            string          `json:"provider"`
	Publisher           string          `json:"publisher"`
	PublishedAt         string          `json:"publishedAt"`
	DiscoveredAt        string          `json:"discoveredAt"`
	AgeMinutes          *int            `json:"ageMinutes"`
	FreshnessConfidence string          `json:"freshnessConfidence"`
	Snippet             string          `json:"snippet"`
	Tickers             []string        `json:"tickers"`
	EventType           string          `json:"eventType"`
	RankScore           float64         `json:"rankScore"`
	Metadata            json.RawMessage `json:"metadata"`
}

type ResponseMetadata struct {
	ProviderRoute    string          `json:"providerRoute"`
	StrictTicker     bool            `json:"strictTicker"`
	CacheHit         bool            `json:"cacheHit"`
	CandidateCounts  json.RawMessage `json:"candidateCounts"`
	NormalizedCounts json.RawMessage `json:"normalizedCounts"`
	ReturnedCounts   json.RawMessage `json:"returnedCounts"`
	Raw              json.RawMessage `json:"-"`
}

type SearchResponse struct {
	Success  bool             `json:"success"`
	Data     []NewsItem       `json:"data"`
	Metadata ResponseMetadata `json:"metadata"`
}

type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("firecrawl news request failed: status %d: %s", e.StatusCode, e.Body)
}

func IsRateLimited(err error) bool {
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr.StatusCode == http.StatusTooManyRequests
	}
	return false
}

type FirecrawlClient struct {
	newsURL string
	apiKey  string
	http    *http.Client
}

func NewFirecrawlClient(newsURL, apiKey string, httpClient *http.Client) *FirecrawlClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}
	return &FirecrawlClient{
		newsURL: strings.TrimRight(strings.TrimSpace(newsURL), "/"),
		apiKey:  strings.TrimSpace(apiKey),
		http:    httpClient,
	}
}

func (c *FirecrawlClient) Enabled() bool {
	return c != nil && c.newsURL != ""
}

func (c *FirecrawlClient) Search(ctx context.Context, request SearchRequest) (*SearchResponse, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("firecrawl news client is not configured")
	}
	if strings.TrimSpace(request.Query) == "" {
		return nil, fmt.Errorf("firecrawl news query is required")
	}
	if request.Limit <= 0 {
		request.Limit = 10
	}
	if request.MaxAgeMinutes <= 0 {
		request.MaxAgeMinutes = 1440
	}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, c.newsURL, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Accept", "application/json")
	if c.apiKey != "" {
		httpRequest.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	response, err := c.http.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, &HTTPError{StatusCode: response.StatusCode, Body: strings.TrimSpace(string(payload))}
	}
	var parsed SearchResponse
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return nil, fmt.Errorf("decode firecrawl news response: %w", err)
	}
	var raw struct {
		Metadata json.RawMessage `json:"metadata"`
	}
	_ = json.Unmarshal(payload, &raw)
	parsed.Metadata.Raw = raw.Metadata
	return &parsed, nil
}

func CanonicalURL(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return strings.TrimSpace(raw)
	}
	parsed.Fragment = ""
	query := parsed.Query()
	for key := range query {
		lower := strings.ToLower(key)
		if strings.HasPrefix(lower, "utm_") || lower == "guccounter" || lower == "ncid" {
			query.Del(key)
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func SourceDomain(rawURL, fallback string) string {
	parsed, err := url.Parse(rawURL)
	if err == nil && parsed.Hostname() != "" {
		return strings.ToLower(parsed.Hostname())
	}
	return strings.ToLower(strings.TrimSpace(fallback))
}

func HeadlineHash(title string) string {
	normalized := strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(title))), " ")
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}

func ParseTime(raw string) time.Time {
	if strings.TrimSpace(raw) == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}
	}
	return parsed.UTC()
}

func NormalizeTickers(tickers []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(tickers))
	for _, ticker := range tickers {
		normalized := strings.ToUpper(strings.TrimSpace(ticker))
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		result = append(result, normalized)
	}
	sort.Strings(result)
	return result
}

func SourceTier(domain string, officialDomains []string) string {
	domain = strings.ToLower(strings.TrimSpace(domain))
	for _, official := range officialDomains {
		official = strings.ToLower(strings.TrimSpace(official))
		if official != "" && (domain == official || strings.HasSuffix(domain, "."+official)) {
			return "T1_OFFICIAL"
		}
	}
	if domain == "sec.gov" || strings.HasSuffix(domain, ".sec.gov") || strings.HasPrefix(domain, "investor.") || strings.HasPrefix(domain, "ir.") {
		return "T1_OFFICIAL"
	}
	if domain == "federalreserve.gov" || strings.HasSuffix(domain, ".gov") || domain == "bls.gov" || domain == "bea.gov" {
		return "T2_PRIMARY"
	}
	if domain == "" {
		return "T4_UNVERIFIED"
	}
	return "T3_REPORTING"
}
