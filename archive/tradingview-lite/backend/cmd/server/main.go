package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"tradingview-lite/backend/internal/alert"
	"tradingview-lite/backend/internal/collector"
	"tradingview-lite/backend/internal/config"
	"tradingview-lite/backend/internal/evidence"
	"tradingview-lite/backend/internal/features"
	"tradingview-lite/backend/internal/market"
	"tradingview-lite/backend/internal/monitor"
	"tradingview-lite/backend/internal/notifier"
	"tradingview-lite/backend/internal/options"
	"tradingview-lite/backend/internal/policy"
	"tradingview-lite/backend/internal/rwa"
	"tradingview-lite/backend/internal/store"
)

type errorResponse struct {
	Error string `json:"error"`
}

func main() {
	cfg := config.Load()
	app := fiber.New(fiber.Config{
		AppName:      "tradingview-lite-backend",
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 20 * time.Second,
	})
	yahoo := market.NewYahooClient(&http.Client{Timeout: 15 * time.Second})
	var repo store.Store
	storeMode := "memory"
	if cfg.DatabaseURL != "" {
		initCtx, initCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer initCancel()
		postgresStore, err := store.NewPostgres(initCtx, cfg.DatabaseURL, cfg.DefaultWatchlistSymbols)
		if err != nil {
			log.Fatalf("connect postgres: %v", err)
		}
		defer postgresStore.Close()
		repo = postgresStore
		storeMode = "postgres"
	} else {
		repo = store.NewMemoryStore(cfg.DefaultWatchlistSymbols)
	}

	var notifiers []alert.Notifier
	var discordNotifier *notifier.Discord
	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var discordOutbox *notifier.Outbox
	if cfg.Discord.Enabled() {
		discordNotifier = notifier.NewDiscord(cfg.Discord, &http.Client{Timeout: 10 * time.Second})
		discordOutbox = notifier.NewOutbox(repo, discordNotifier)
		notifiers = append(notifiers, discordOutbox)
	}

	var rwaClient *rwa.BitgetClient
	if cfg.RWA.Enabled() {
		rwaClient = rwa.NewBitgetClient(cfg.RWA.BitgetMCPURL, cfg.RWA.BitgetMCPToken, cfg.RWA.SymbolMap, &http.Client{Timeout: 12 * time.Second})
	}

	var optionsClient *options.YahooMCPClient
	if cfg.Options.Enabled() {
		optionsClient = options.NewYahooMCPClient(cfg.Options.YahooMCPURL, cfg.Options.YahooMCPToken, cfg.Options.RiskFreeRate, &http.Client{Timeout: 18 * time.Second})
	}

	evaluator := alert.NewEvaluator(repo, cfg.DataStaleAfter, notifiers...)
	mon := monitor.New(repo, yahoo, evaluator)
	dataCollector := collector.New(repo, yahoo, optionsClient, rwaClient, collector.Config{
		DataStaleAfter:           cfg.DataStaleAfter,
		MarketInterval:           cfg.Collection.MarketInterval,
		FocusQuoteInterval:       cfg.Collection.FocusQuoteInterval,
		RWAInterval:              cfg.Collection.RWAInterval,
		RWAFocusInterval:         cfg.Collection.RWAFocusInterval,
		OptionsInterval:          cfg.Collection.OptionsInterval,
		FeatureInterval:          cfg.Collection.FeatureInterval,
		AlertInterval:            cfg.Collection.AlertInterval,
		RetentionInterval:        cfg.Collection.RetentionInterval,
		EventSymbolInterval:      cfg.Events.SymbolInterval,
		EventMacroInterval:       cfg.Events.MacroInterval,
		EventContextInterval:     cfg.Events.ContextInterval,
		ShadowSignalInterval:     cfg.Events.ShadowInterval,
		EventTriggeredCooldown:   cfg.Events.TriggeredCooldown,
		MaxBackoff:               cfg.Collection.MaxBackoff,
		YahooRequestsPerMinute:   cfg.Collection.YahooRequestsPerMinute,
		OptionsRequestsPerMinute: cfg.Collection.OptionsRequestsPerMinute,
		RWARequestsPerMinute:     cfg.Collection.RWARequestsPerMinute,
		EventRequestsPerMinute:   cfg.Events.RequestsPerMinute,
		FocusSymbols:             cfg.Collection.FocusSymbols,
		RWASymbols:               cfg.Collection.RWASymbols,
		RWAFocusSymbols:          cfg.Collection.RWAFocusSymbols,
		OptionSymbols:            cfg.Collection.OptionSymbols,
		EventSymbols:             cfg.Events.Symbols,
		ShadowSymbols:            cfg.Events.ShadowSymbols,
		MacroQueries:             cfg.Events.MacroQueries,
		GroupQueries:             cfg.Events.GroupQueries,
		OfficialDomains:          cfg.Events.OfficialDomains,
		EventMaxAgeMinutes:       cfg.Events.MaxAgeMinutes,
		PolicyEnabled:            cfg.Policy.Enabled,
		PolicyCandidateInterval:  cfg.Policy.CandidateInterval,
		PolicyOutcomeInterval:    cfg.Policy.OutcomeInterval,
		PolicySymbols:            cfg.Policy.Symbols,
		PolicyConfig: policy.Config{
			DataStaleAfter:       cfg.DataStaleAfter,
			CandidateCooldown:    cfg.Policy.CandidateCooldown,
			DailyCandidateBudget: cfg.Policy.DailyCandidateBudget,
		},
		Retention: store.RetentionPolicy{
			QuoteSnapshots:   cfg.Collection.QuoteRetention,
			RWASnapshots:     cfg.Collection.RWARetention,
			OptionSnapshots:  cfg.Collection.OptionRetention,
			FeatureSnapshots: cfg.Collection.FeatureRetention,
			CollectionRuns:   cfg.Collection.RunRetention,
			EventEvidence:    cfg.Events.EvidenceRetention,
			EventContexts:    cfg.Events.ContextRetention,
			ShadowSignals:    cfg.Events.ShadowRetention,
			EventRuns:        cfg.Events.RunRetention,
			PolicyCandidates: cfg.Policy.CandidateRetention,
			PolicyOutcomes:   cfg.Policy.OutcomeRetention,
		},
	})
	dataCollector.WithAlertEvaluator(evaluator)
	if cfg.Events.EnabledForCollection() {
		dataCollector.WithEvidence(evidence.NewFirecrawlClient(cfg.Events.FirecrawlNewsURL, cfg.Events.FirecrawlAPIKey, &http.Client{Timeout: 25 * time.Second}))
	}
	if discordOutbox != nil {
		discordOutbox.Start(rootCtx, cfg.Discord.DeliveryInterval)
	}

	if cfg.EnableScheduler && cfg.Collection.Enabled {
		dataCollector.Start(rootCtx)
	} else if cfg.EnableScheduler {
		mon.Start(rootCtx, cfg.PollInterval)
	}

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"ok":                             true,
			"host":                           cfg.Host,
			"provider":                       "yahoo_chart",
			"store":                          storeMode,
			"discord_enabled":                cfg.Discord.Enabled(),
			"rwa_enabled":                    cfg.RWA.Enabled(),
			"options_enabled":                cfg.Options.Enabled(),
			"event_evidence_enabled":         cfg.Events.EnabledForCollection(),
			"behavior_policy_shadow_enabled": cfg.Policy.Enabled,
			"policy_daily_candidate_budget":  cfg.Policy.DailyCandidateBudget,
			"scheduler":                      cfg.EnableScheduler,
			"poll_interval":                  cfg.PollInterval.String(),
			"collectors_enabled":             cfg.EnableScheduler && cfg.Collection.Enabled,
			"collector_market_interval":      cfg.Collection.MarketInterval.String(),
			"collector_budgets_per_minute": fiber.Map{
				"yahoo_chart":    cfg.Collection.YahooRequestsPerMinute,
				"yahoo_options":  cfg.Collection.OptionsRequestsPerMinute,
				"bitget_rwa":     cfg.Collection.RWARequestsPerMinute,
				"firecrawl_news": cfg.Events.RequestsPerMinute,
			},
			"data_stale_after": cfg.DataStaleAfter.String(),
			"time":             time.Now().UTC(),
		})
	})

	api := app.Group("/api/v1")
	api.Get("/watchlist", func(c *fiber.Ctx) error {
		items, err := repo.ListWatchlist(c.Context())
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(items)
	})

	api.Post("/watchlist", func(c *fiber.Ctx) error {
		var item store.WatchlistItem
		if err := c.BodyParser(&item); err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		if err := repo.AddWatchlist(c.Context(), item); err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		return c.Status(fiber.StatusCreated).JSON(item)
	})

	api.Delete("/watchlist/:symbol", func(c *fiber.Ctx) error {
		if err := repo.DeleteWatchlist(c.Context(), c.Params("symbol")); err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	api.Get("/quote/:symbol", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 15*time.Second)
		defer cancel()

		quote, err := yahoo.FetchQuote(ctx, c.Params("symbol"))
		if err != nil {
			return jsonError(c, fiber.StatusBadGateway, err)
		}
		return c.JSON(quote)
	})

	api.Get("/quotes", func(c *fiber.Ctx) error {
		quotes, err := repo.ListQuotes(c.Context())
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(quotes)
	})

	api.Get("/bars", func(c *fiber.Ctx) error {
		symbol := c.Query("symbol")
		if symbol == "" {
			return jsonError(c, fiber.StatusBadRequest, errors.New("symbol query parameter is required"))
		}

		ctx, cancel := context.WithTimeout(c.Context(), 15*time.Second)
		defer cancel()

		bars, err := yahoo.FetchBars(
			ctx,
			symbol,
			c.Query("range", "1d"),
			c.Query("interval", "1m"),
			c.QueryBool("includePrePost", false),
		)
		if err != nil {
			return jsonError(c, fiber.StatusBadGateway, err)
		}
		return c.JSON(bars)
	})

	api.Get("/features/:symbol/intraday", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 35*time.Second)
		defer cancel()

		result, err := features.ComputeIntraday(ctx, c.Params("symbol"), yahoo, optionsClient, rwaClient)
		if err != nil {
			return jsonError(c, fiber.StatusBadGateway, err)
		}
		return c.JSON(result)
	})

	api.Get("/options/:symbol/chain", func(c *fiber.Ctx) error {
		if optionsClient == nil {
			return jsonError(c, fiber.StatusServiceUnavailable, errors.New("yahoo options provider is not configured"))
		}
		ctx, cancel := context.WithTimeout(c.Context(), 20*time.Second)
		defer cancel()

		chain, err := optionsClient.FetchChain(ctx, c.Params("symbol"), c.QueryInt("near", 20))
		if err != nil {
			return jsonError(c, fiber.StatusBadGateway, err)
		}
		return c.JSON(chain)
	})

	api.Get("/options/:symbol/expected-move", func(c *fiber.Ctx) error {
		if optionsClient == nil {
			return jsonError(c, fiber.StatusServiceUnavailable, errors.New("yahoo options provider is not configured"))
		}
		ctx, cancel := context.WithTimeout(c.Context(), 20*time.Second)
		defer cancel()

		chain, err := optionsClient.FetchChain(ctx, c.Params("symbol"), c.QueryInt("near", 8))
		if err != nil {
			return jsonError(c, fiber.StatusBadGateway, err)
		}
		return c.JSON(fiber.Map{
			"symbol":          chain.Symbol,
			"provider":        chain.Provider,
			"source":          chain.Source,
			"receivedAt":      chain.ReceivedAt,
			"expirationDate":  chain.ExpirationDate,
			"underlying":      chain.Underlying,
			"expectedMove":    chain.Summary,
			"providerWarning": chain.ProviderWarning,
		})
	})

	api.Get("/rwa/:symbol/info", func(c *fiber.Ctx) error {
		if rwaClient == nil {
			return jsonError(c, fiber.StatusServiceUnavailable, errors.New("bitget rwa provider is not configured"))
		}
		ctx, cancel := context.WithTimeout(c.Context(), 12*time.Second)
		defer cancel()

		info, err := rwaClient.StockInfo(ctx, c.Params("symbol"))
		if err != nil {
			return jsonError(c, fiber.StatusBadGateway, err)
		}
		return c.JSON(info)
	})

	api.Get("/rwa/:symbol/kline", func(c *fiber.Ctx) error {
		if rwaClient == nil {
			return jsonError(c, fiber.StatusServiceUnavailable, errors.New("bitget rwa provider is not configured"))
		}
		ctx, cancel := context.WithTimeout(c.Context(), 12*time.Second)
		defer cancel()

		bars, err := rwaClient.Kline(ctx, c.Params("symbol"), c.Query("period", "1m"), c.QueryInt("size", 60))
		if err != nil {
			return jsonError(c, fiber.StatusBadGateway, err)
		}
		return c.JSON(bars)
	})

	api.Get("/rwa/:symbol/summary", func(c *fiber.Ctx) error {
		if rwaClient == nil {
			return jsonError(c, fiber.StatusServiceUnavailable, errors.New("bitget rwa provider is not configured"))
		}
		ctx, cancel := context.WithTimeout(c.Context(), 18*time.Second)
		defer cancel()

		symbol := c.Params("symbol")
		info, err := rwaClient.StockInfo(ctx, symbol)
		if err != nil {
			return jsonError(c, fiber.StatusBadGateway, err)
		}
		kline, err := rwaClient.Kline(ctx, symbol, c.Query("period", "1m"), c.QueryInt("size", 60))
		if err != nil {
			return jsonError(c, fiber.StatusBadGateway, err)
		}
		return c.JSON(fiber.Map{
			"symbol":      info.Symbol,
			"ticker":      info.Ticker,
			"provider":    info.Provider,
			"received_at": time.Now().UTC(),
			"info":        info,
			"kline":       kline,
		})
	})

	api.Get("/stored-bars", func(c *fiber.Ctx) error {
		symbol := c.Query("symbol")
		if symbol == "" {
			return jsonError(c, fiber.StatusBadRequest, errors.New("symbol query parameter is required"))
		}
		limit := c.QueryInt("limit", 390)
		bars, err := repo.GetBars(c.Context(), symbol, c.Query("interval", "1m"), limit)
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(fiber.Map{
			"symbol":   symbol,
			"interval": c.Query("interval", "1m"),
			"count":    len(bars),
			"bars":     bars,
		})
	})

	api.Get("/snapshots/quotes", func(c *fiber.Ctx) error {
		bounds, err := queryTimeRange(c)
		if err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		rows, err := repo.ListQuoteSnapshots(c.Context(), c.Query("symbol"), c.Query("provider"), bounds)
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(rows)
	})

	api.Get("/snapshots/rwa", func(c *fiber.Ctx) error {
		bounds, err := queryTimeRange(c)
		if err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		rows, err := repo.ListRWASnapshots(c.Context(), c.Query("symbol"), c.Query("provider"), bounds)
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(rows)
	})

	api.Get("/snapshots/rwa/bars", func(c *fiber.Ctx) error {
		symbol := c.Query("symbol")
		if symbol == "" {
			return jsonError(c, fiber.StatusBadRequest, errors.New("symbol query parameter is required"))
		}
		bars, err := repo.GetRWABars(c.Context(), symbol, c.Query("interval", "1m"), c.QueryInt("limit", 390))
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(fiber.Map{"symbol": symbol, "interval": c.Query("interval", "1m"), "count": len(bars), "bars": bars})
	})

	api.Get("/snapshots/options/contracts", func(c *fiber.Ctx) error {
		bounds, err := queryTimeRange(c)
		if err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		rows, err := repo.ListOptionContracts(c.Context(), c.Query("symbol"), bounds)
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(rows)
	})

	api.Get("/snapshots/options/iv", func(c *fiber.Ctx) error {
		bounds, err := queryTimeRange(c)
		if err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		rows, err := repo.ListIVSnapshots(c.Context(), c.Query("symbol"), bounds)
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(rows)
	})

	api.Get("/snapshots/features", func(c *fiber.Ctx) error {
		bounds, err := queryTimeRange(c)
		if err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		rows, err := repo.ListFeatureSnapshots(c.Context(), c.Query("symbol"), bounds)
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(rows)
	})

	api.Get("/event-evidence", func(c *fiber.Ctx) error {
		bounds, err := queryTimeRange(c)
		if err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		rows, err := repo.ListEventEvidence(c.Context(), store.EventFilter{
			Symbol: c.Query("symbol"),
			Group:  c.Query("group"),
			Bounds: bounds,
		})
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(rows)
	})

	api.Get("/market-events", func(c *fiber.Ctx) error {
		bounds, err := queryTimeRange(c)
		if err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		rows, err := repo.ListMarketEvents(c.Context(), store.EventFilter{
			Symbol: c.Query("symbol"),
			Group:  c.Query("group"),
			Bounds: bounds,
		})
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(rows)
	})

	api.Get("/event-context/:symbol", func(c *fiber.Ctx) error {
		asOf, err := queryAsOf(c)
		if err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		snapshot, err := repo.GetEventContext(c.Context(), c.Params("symbol"), asOf)
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		if snapshot == nil {
			return c.Status(fiber.StatusNotFound).JSON(errorResponse{Error: "event context not found"})
		}
		return c.JSON(snapshot)
	})

	api.Get("/shadow-signals", func(c *fiber.Ctx) error {
		symbol := c.Query("symbol")
		if symbol == "" {
			return jsonError(c, fiber.StatusBadRequest, errors.New("symbol query parameter is required"))
		}
		bounds, err := queryTimeRange(c)
		if err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		rows, err := repo.ListShadowSignals(c.Context(), symbol, bounds)
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(rows)
	})

	api.Get("/policy/profiles", func(c *fiber.Ctx) error {
		profiles, err := repo.ListSymbolProfiles(c.Context())
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(profiles)
	})

	api.Put("/policy/profiles/:symbol", func(c *fiber.Ctx) error {
		var profile store.SymbolProfile
		if err := c.BodyParser(&profile); err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		profile.Symbol = c.Params("symbol")
		saved, err := repo.UpsertSymbolProfile(c.Context(), profile)
		if err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		return c.JSON(saved)
	})

	api.Get("/policy/peer-maps", func(c *fiber.Ctx) error {
		maps, err := repo.ListPeerMapVersions(c.Context(), c.Query("map_key"))
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(maps)
	})

	api.Post("/policy/peer-maps", func(c *fiber.Ctx) error {
		var peerMap store.PeerMapVersion
		if err := c.BodyParser(&peerMap); err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		created, err := repo.CreatePeerMapVersion(c.Context(), peerMap)
		if err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		return c.Status(fiber.StatusCreated).JSON(created)
	})

	api.Get("/policy/candidates", func(c *fiber.Ctx) error {
		bounds, err := queryTimeRange(c)
		if err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		tradingDate, err := queryDate(c, "trading_date")
		if err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		rows, err := repo.ListPolicyCandidates(c.Context(), store.PolicyCandidateFilter{
			Symbol:          c.Query("symbol"),
			TradingDate:     tradingDate,
			OnlyWouldNotify: c.QueryBool("would_notify", false),
			Bounds:          bounds,
		})
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(rows)
	})

	api.Get("/policy/candidates/:id", func(c *fiber.Ctx) error {
		id, err := policyCandidateID(c)
		if err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		candidate, err := repo.GetPolicyCandidate(c.Context(), id)
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		if candidate == nil {
			return c.Status(fiber.StatusNotFound).JSON(errorResponse{Error: "policy candidate not found"})
		}
		return c.JSON(candidate)
	})

	api.Get("/policy/candidates/:id/outcomes", func(c *fiber.Ctx) error {
		id, err := policyCandidateID(c)
		if err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		rows, err := repo.ListPolicyCandidateOutcomes(c.Context(), id)
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(rows)
	})

	api.Get("/policy/candidates/:id/feedback", func(c *fiber.Ctx) error {
		id, err := policyCandidateID(c)
		if err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		feedback, err := repo.GetPolicyCandidateFeedback(c.Context(), id)
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		if feedback == nil {
			return c.Status(fiber.StatusNotFound).JSON(errorResponse{Error: "policy candidate feedback not found"})
		}
		return c.JSON(feedback)
	})

	api.Put("/policy/candidates/:id/feedback", func(c *fiber.Ctx) error {
		id, err := policyCandidateID(c)
		if err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		var feedback store.PolicyCandidateFeedback
		if err := c.BodyParser(&feedback); err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		feedback.CandidateAlertID = id
		saved, err := repo.UpsertPolicyCandidateFeedback(c.Context(), feedback)
		if err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		return c.JSON(saved)
	})

	api.Get("/policy/candidates/:id/discord-preview", func(c *fiber.Ctx) error {
		id, err := policyCandidateID(c)
		if err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		candidate, err := repo.GetPolicyCandidate(c.Context(), id)
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		if candidate == nil {
			return c.Status(fiber.StatusNotFound).JSON(errorResponse{Error: "policy candidate not found"})
		}
		return c.JSON(fiber.Map{
			"deliveryMode":                "shadow_only",
			"normalDiscordWebhookInvoked": false,
			"wouldNotifyIfApproved":       candidate.WouldNotify,
			"budgetSlot":                  candidate.BudgetSlot,
			"embed": fiber.Map{
				"title":       "SHADOW ONLY · " + candidate.Symbol + " · " + candidate.State,
				"description": candidate.ActionCandidate,
				"fields": []fiber.Map{
					{"name": "Session", "value": candidate.SessionProduct, "inline": true},
					{"name": "Profile", "value": candidate.ProfileRole, "inline": true},
					{"name": "Action", "value": candidate.ActionCandidate, "inline": true},
				},
			},
		})
	})

	api.Post("/policy/:symbol/evaluate", func(c *fiber.Ctx) error {
		asOf, err := queryAsOf(c)
		if err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		if asOf.IsZero() {
			asOf = time.Now().UTC()
		}
		candidate, err := (policy.Evaluator{Store: repo, Config: behaviorPolicyConfig(cfg)}).Evaluate(c.Context(), c.Params("symbol"), watchlistGroup(c.Context(), repo, c.Params("symbol")), asOf)
		if err != nil {
			return jsonError(c, fiber.StatusBadGateway, err)
		}
		saved, err := repo.AppendPolicyCandidate(c.Context(), candidate)
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.Status(fiber.StatusCreated).JSON(saved)
	})

	api.Post("/policy/:symbol/outcomes/evaluate", func(c *fiber.Ctx) error {
		asOf, err := queryAsOf(c)
		if err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		if asOf.IsZero() {
			asOf = time.Now().UTC()
		}
		count, err := (policy.OutcomeEvaluator{Store: repo}).Evaluate(c.Context(), c.Params("symbol"), asOf)
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(fiber.Map{"recordsWritten": count, "deliveryMode": "shadow_only"})
	})

	api.Get("/policy/calibration/rwa/:symbol", func(c *fiber.Ctx) error {
		asOf, err := queryAsOf(c)
		if err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		if asOf.IsZero() {
			asOf = time.Now().UTC()
		}
		result, err := policy.CalibrateRWA(c.Context(), repo, c.Params("symbol"), asOf, behaviorPolicyConfig(cfg))
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(result)
	})

	api.Get("/policy/calibration/range/:symbol", func(c *fiber.Ctx) error {
		asOf, err := queryAsOf(c)
		if err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		if asOf.IsZero() {
			asOf = time.Now().UTC()
		}
		result, err := policy.CalibrateRange(c.Context(), repo, c.Params("symbol"), asOf, behaviorPolicyConfig(cfg))
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(result)
	})

	api.Get("/policy/calibration/iv/:symbol", func(c *fiber.Ctx) error {
		asOf, err := queryAsOf(c)
		if err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		if asOf.IsZero() {
			asOf = time.Now().UTC()
		}
		result, err := policy.CalibrateIV(c.Context(), repo, c.Params("symbol"), asOf, behaviorPolicyConfig(cfg))
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(result)
	})

	api.Get("/event-ingestion/runs", func(c *fiber.Ctx) error {
		runs, err := repo.ListEventIngestionRuns(c.Context(), c.QueryInt("limit", 100))
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(runs)
	})

	api.Get("/event-ingestion/storage", func(c *fiber.Ctx) error {
		diagnostic, err := repo.StorageDiagnostics(c.Context())
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(diagnostic)
	})

	api.Get("/daily/:symbol", func(c *fiber.Ctx) error {
		date, err := time.Parse("2006-01-02", c.Query("date", time.Now().UTC().Format("2006-01-02")))
		if err != nil {
			return jsonError(c, fiber.StatusBadRequest, errors.New("date must use YYYY-MM-DD"))
		}
		summary, err := repo.GetDailySummary(c.Context(), c.Params("symbol"), date.UTC())
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		if summary == nil {
			return c.Status(fiber.StatusNotFound).JSON(errorResponse{Error: "daily summary not found"})
		}
		return c.JSON(summary)
	})

	api.Get("/collection/runs", func(c *fiber.Ctx) error {
		runs, err := repo.ListCollectionRuns(c.Context(), c.QueryInt("limit", 100))
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(runs)
	})

	api.Get("/collection/storage", func(c *fiber.Ctx) error {
		diagnostic, err := repo.StorageDiagnostics(c.Context())
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(diagnostic)
	})

	api.Post("/collection/:collector/run", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 90*time.Second)
		defer cancel()
		startedAt := time.Now().UTC()
		var (
			result collector.Result
			err    error
		)
		switch c.Params("collector") {
		case "market":
			result, err = dataCollector.CollectMarket(ctx)
		case "focus":
			result, err = dataCollector.CollectFocusQuotes(ctx)
		case "rwa":
			result, err = dataCollector.CollectRWA(ctx)
		case "rwa-focus":
			result, err = dataCollector.CollectFocusRWA(ctx)
		case "options":
			result, err = dataCollector.CollectOptions(ctx)
		case "features":
			result, err = dataCollector.CollectFeatures(ctx)
		case "event-symbol":
			result, err = dataCollector.CollectEventEvidence(ctx)
		case "event-macro":
			result, err = dataCollector.CollectMacroEvidence(ctx)
		case "event-context":
			result, err = dataCollector.CollectEventContexts(ctx)
		case "shadow":
			result, err = dataCollector.CollectShadowSignals(ctx)
		case "behavior-policy":
			result, err = dataCollector.CollectPolicyCandidates(ctx)
		case "policy-outcomes":
			result, err = dataCollector.CollectPolicyOutcomes(ctx)
		case "retention":
			result, err = dataCollector.CollectRetention(ctx)
		default:
			return jsonError(c, fiber.StatusBadRequest, errors.New("collector must be market, focus, rwa, rwa-focus, options, features, event-symbol, event-macro, event-context, shadow, behavior-policy, policy-outcomes, or retention"))
		}
		if err != nil {
			return jsonError(c, fiber.StatusBadGateway, err)
		}
		finishedAt := time.Now().UTC()
		if err := repo.AddCollectionRun(c.Context(), store.CollectionRun{
			Collector:        result.Collector,
			Provider:         result.Provider,
			StartedAt:        startedAt,
			FinishedAt:       finishedAt,
			SymbolsAttempted: result.SymbolsAttempted,
			SymbolsSucceeded: result.SymbolsSucceeded,
			RecordsWritten:   result.RecordsWritten,
			RateLimited:      result.RateLimited,
			ErrorSummary:     result.ErrorSummary,
		}); err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(result)
	})

	api.Post("/poll", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 60*time.Second)
		defer cancel()
		result, err := mon.PollOnce(ctx)
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(result)
	})

	api.Get("/poll/last", func(c *fiber.Ctx) error {
		result := mon.LastResult()
		if result == nil {
			return c.JSON(fiber.Map{"status": "not_run"})
		}
		return c.JSON(result)
	})

	api.Get("/alerts", func(c *fiber.Ctx) error {
		configs, err := repo.ListAlerts(c.Context())
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(configs)
	})

	api.Post("/alerts", func(c *fiber.Ctx) error {
		var cfg alert.AlertConfig
		if err := c.BodyParser(&cfg); err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		created, err := repo.CreateAlert(c.Context(), cfg)
		if err != nil {
			return jsonError(c, fiber.StatusBadRequest, err)
		}
		return c.Status(fiber.StatusCreated).JSON(created)
	})

	api.Post("/alerts/evaluate", func(c *fiber.Ctx) error {
		events, err := evaluator.EvaluateAll(c.Context())
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(fiber.Map{"count": len(events), "events": events})
	})

	api.Get("/events", func(c *fiber.Ctx) error {
		events, err := repo.ListEvents(c.Context(), c.QueryInt("limit", 100))
		if err != nil {
			return jsonError(c, fiber.StatusInternalServerError, err)
		}
		return c.JSON(events)
	})

	api.Post("/notifiers/discord/test", func(c *fiber.Ctx) error {
		if discordNotifier == nil {
			return jsonError(c, fiber.StatusBadRequest, errors.New("discord webhook is not configured"))
		}
		event := alert.Event{
			ID:                "discord_test",
			Symbol:            "TEST",
			Level:             alert.LevelAlert,
			RuleType:          alert.RulePriceLevel,
			Message:           "Discord webhook test from TradingView-Lite Risk Agent.",
			ObservedValue:     1,
			ThresholdValue:    1,
			DataAgeSeconds:    0,
			ProviderTime:      time.Now().UTC(),
			ReceivedAt:        time.Now().UTC(),
			TriggeredAt:       time.Now().UTC(),
			AnalysisOnTrigger: false,
		}
		if err := discordNotifier.Send(c.Context(), event); err != nil {
			return jsonError(c, fiber.StatusBadGateway, err)
		}
		return c.JSON(fiber.Map{"ok": true})
	})

	addr := net.JoinHostPort(cfg.Host, cfg.Port)
	log.Printf("tradingview-lite backend listening on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatal(err)
	}
}

func jsonError(c *fiber.Ctx, status int, err error) error {
	return c.Status(status).JSON(errorResponse{Error: err.Error()})
}

func queryTimeRange(c *fiber.Ctx) (store.TimeRange, error) {
	bounds := store.TimeRange{Limit: c.QueryInt("limit", 1000)}
	if bounds.Limit < 1 || bounds.Limit > 10000 {
		return bounds, errors.New("limit must be between 1 and 10000")
	}
	parse := func(name string) (time.Time, error) {
		value := c.Query(name)
		if value == "" {
			return time.Time{}, nil
		}
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return time.Time{}, errors.New(name + " must use RFC3339 format")
		}
		return parsed.UTC(), nil
	}
	var err error
	if bounds.From, err = parse("from"); err != nil {
		return bounds, err
	}
	if bounds.To, err = parse("to"); err != nil {
		return bounds, err
	}
	if !bounds.From.IsZero() && !bounds.To.IsZero() && bounds.To.Before(bounds.From) {
		return bounds, errors.New("to must not be earlier than from")
	}
	return bounds, nil
}

func queryAsOf(c *fiber.Ctx) (time.Time, error) {
	value := c.Query("as_of")
	if value == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, errors.New("as_of must use RFC3339 format")
	}
	return parsed.UTC(), nil
}

func queryDate(c *fiber.Ctx, name string) (time.Time, error) {
	value := c.Query(name)
	if value == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, errors.New(name + " must use YYYY-MM-DD")
	}
	return parsed.UTC(), nil
}

func policyCandidateID(c *fiber.Ctx) (int64, error) {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("policy candidate id must be a positive integer")
	}
	return id, nil
}

func behaviorPolicyConfig(cfg config.Config) policy.Config {
	return policy.Config{
		DataStaleAfter:       cfg.DataStaleAfter,
		CandidateCooldown:    cfg.Policy.CandidateCooldown,
		DailyCandidateBudget: cfg.Policy.DailyCandidateBudget,
	}
}

func watchlistGroup(ctx context.Context, repo store.Store, symbol string) string {
	items, err := repo.ListWatchlist(ctx)
	if err != nil {
		return "custom"
	}
	for _, item := range items {
		if item.Symbol == symbol {
			return item.Group
		}
	}
	return "custom"
}
