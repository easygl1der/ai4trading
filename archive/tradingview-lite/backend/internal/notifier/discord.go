package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"tradingview-lite/backend/internal/alert"
	"tradingview-lite/backend/internal/config"
)

type Discord struct {
	httpClient *http.Client
	webhooks   map[alert.Level]string
}

func NewDiscord(cfg config.DiscordConfig, httpClient *http.Client) *Discord {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Discord{
		httpClient: httpClient,
		webhooks: map[alert.Level]string{
			alert.LevelWatch:           cfg.Watch,
			alert.LevelAlert:           cfg.Alerts,
			alert.LevelAnalysisTrigger: cfg.Alerts,
			alert.LevelUrgent:          cfg.Urgent,
			alert.LevelBlocked:         cfg.System,
		},
	}
}

func (d *Discord) Send(ctx context.Context, event alert.Event) error {
	webhookURL := strings.TrimSpace(d.webhookFor(event.Level))
	if webhookURL == "" {
		return nil
	}

	payload := discordPayload{
		Username:        "Risk Agent",
		AllowedMentions: discordAllowedMentions{Parse: []string{}},
		Embeds:          []discordEmbed{embedForEvent(event)},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	err = d.post(ctx, webhookURL, body)
	if err == nil {
		return nil
	}
	return err
}

func (d *Discord) webhookFor(level alert.Level) string {
	if webhook := d.webhooks[level]; webhook != "" {
		return webhook
	}
	if level == alert.LevelInfo {
		return ""
	}
	return d.webhooks[alert.LevelAlert]
}

func (d *Discord) post(ctx context.Context, webhookURL string, body []byte) error {
	res, err := d.postOnce(ctx, webhookURL, body)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusTooManyRequests {
		retryAfter := retryAfterDuration(res)
		if retryAfter <= 0 || retryAfter > 5*time.Second {
			return responseError(res)
		}
		if _, err := io.Copy(io.Discard, res.Body); err != nil {
			log.Printf("discord response drain failed: %v", err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryAfter):
		}
		retryRes, err := d.postOnce(ctx, webhookURL, body)
		if err != nil {
			return err
		}
		defer retryRes.Body.Close()
		if retryRes.StatusCode < 200 || retryRes.StatusCode >= 300 {
			return responseError(retryRes)
		}
		return nil
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return responseError(res)
	}
	return nil
}

func (d *Discord) postOnce(ctx context.Context, webhookURL string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "tradingview-lite-risk-agent/0.1")
	return d.httpClient.Do(req)
}

func retryAfterDuration(res *http.Response) time.Duration {
	header := strings.TrimSpace(res.Header.Get("Retry-After"))
	if header == "" {
		return 0
	}
	seconds, err := strconv.ParseFloat(header, 64)
	if err != nil {
		return 0
	}
	return time.Duration(seconds * float64(time.Second))
}

func responseError(res *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
	return fmt.Errorf("discord webhook returned HTTP %d: %s", res.StatusCode, string(body))
}

func embedForEvent(event alert.Event) discordEmbed {
	fields := []discordField{
		{Name: "Symbol", Value: emptyDash(event.Symbol), Inline: true},
		{Name: "Level", Value: string(event.Level), Inline: true},
		{Name: "Rule", Value: emptyDash(string(event.RuleType)), Inline: true},
		{Name: "Observed", Value: formatFloat(event.ObservedValue), Inline: true},
		{Name: "Threshold", Value: formatFloat(event.ThresholdValue), Inline: true},
		{Name: "Data age", Value: fmt.Sprintf("%ds", event.DataAgeSeconds), Inline: true},
	}
	if event.Benchmark != "" {
		fields = append(fields, discordField{Name: "Benchmark", Value: event.Benchmark, Inline: true})
	}

	timestamp := event.TriggeredAt
	if timestamp.IsZero() {
		timestamp = time.Now().UTC()
	}
	return discordEmbed{
		Title:       fmt.Sprintf("%s · %s", event.Level, event.Symbol),
		Description: event.Message,
		Color:       colorForLevel(event.Level),
		Fields:      fields,
		Footer:      discordFooter{Text: "TradingView-Lite Risk Agent"},
		Timestamp:   timestamp.Format(time.RFC3339),
	}
}

func colorForLevel(level alert.Level) int {
	switch level {
	case alert.LevelInfo:
		return 8421504
	case alert.LevelWatch:
		return 3447003
	case alert.LevelAlert, alert.LevelAnalysisTrigger:
		return 15105570
	case alert.LevelUrgent:
		return 15158332
	case alert.LevelBlocked:
		return 10181046
	default:
		return 8421504
	}
}

func formatFloat(value float64) string {
	if value == 0 {
		return "-"
	}
	return fmt.Sprintf("%.4f", value)
}

func emptyDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

type discordPayload struct {
	Username        string                 `json:"username,omitempty"`
	Content         string                 `json:"content,omitempty"`
	AllowedMentions discordAllowedMentions `json:"allowed_mentions"`
	Embeds          []discordEmbed         `json:"embeds,omitempty"`
}

type discordAllowedMentions struct {
	Parse []string `json:"parse"`
}

type discordEmbed struct {
	Title       string         `json:"title,omitempty"`
	Description string         `json:"description,omitempty"`
	Color       int            `json:"color,omitempty"`
	Fields      []discordField `json:"fields,omitempty"`
	Footer      discordFooter  `json:"footer,omitempty"`
	Timestamp   string         `json:"timestamp,omitempty"`
}

type discordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type discordFooter struct {
	Text string `json:"text,omitempty"`
}
