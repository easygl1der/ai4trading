package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"tradingview-lite/backend/internal/alert"
	"tradingview-lite/backend/internal/market"
	"tradingview-lite/backend/internal/options"
	"tradingview-lite/backend/internal/session"
	"tradingview-lite/backend/migrations"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context, databaseURL string, defaultSymbols []string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	s := &PostgresStore{pool: pool}
	if err := s.runMigrations(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	for _, symbol := range defaultSymbols {
		_ = s.AddWatchlist(ctx, WatchlistItem{
			Symbol: normalizeSymbol(symbol),
			Group:  defaultGroup(symbol),
		})
	}
	return s, nil
}

func (s *PostgresStore) Close() {
	s.pool.Close()
}

func (s *PostgresStore) runMigrations(ctx context.Context) error {
	db := stdlib.OpenDBFromPool(s.pool)
	defer db.Close()
	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.UpContext(ctx, db, ".")
}

func (s *PostgresStore) ListWatchlist(ctx context.Context) ([]WatchlistItem, error) {
	rows, err := s.pool.Query(ctx, `SELECT symbol, group_name FROM watchlist_items ORDER BY group_name, symbol`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []WatchlistItem
	for rows.Next() {
		var item WatchlistItem
		if err := rows.Scan(&item.Symbol, &item.Group); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *PostgresStore) AddWatchlist(ctx context.Context, item WatchlistItem) error {
	item.Symbol = normalizeSymbol(item.Symbol)
	if item.Symbol == "" {
		return errors.New("symbol is required")
	}
	if item.Group == "" {
		item.Group = "custom"
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO watchlist_items(symbol, group_name)
		VALUES($1, $2)
		ON CONFLICT(symbol) DO UPDATE SET group_name = EXCLUDED.group_name
	`, item.Symbol, item.Group)
	return err
}

func (s *PostgresStore) DeleteWatchlist(ctx context.Context, symbol string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM watchlist_items WHERE symbol = $1`, normalizeSymbol(symbol))
	return err
}

func (s *PostgresStore) UpsertBars(ctx context.Context, symbol, timeframe, provider string, bars []market.Bar) (int, error) {
	if len(bars) == 0 {
		return 0, nil
	}

	batch := &pgx.Batch{}
	queued := 0
	for _, bar := range bars {
		if bar.Time.IsZero() {
			continue
		}
		batch.Queue(`
			INSERT INTO ohlcv_bars(symbol, timeframe, provider, ts, open, high, low, close, volume, session)
			VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT(symbol, timeframe, ts) DO UPDATE SET
			  provider = EXCLUDED.provider,
			  open = EXCLUDED.open,
			  high = EXCLUDED.high,
			  low = EXCLUDED.low,
			  close = EXCLUDED.close,
			  volume = EXCLUDED.volume,
		  received_at = NOW()
		`, normalizeSymbol(symbol), strings.ToLower(timeframe), provider, bar.Time, bar.Open, bar.High, bar.Low, bar.Close, bar.Volume, session.ClassifyUS(bar.Time))
		queued++
	}
	if queued == 0 {
		return 0, nil
	}

	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()
	count := 0
	for range queued {
		if _, err := br.Exec(); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (s *PostgresStore) GetBars(ctx context.Context, symbol, timeframe string, limit int) ([]market.Bar, error) {
	if limit <= 0 {
		limit = 390
	}
	rows, err := s.pool.Query(ctx, `
		SELECT ts, open, high, low, close, volume
		FROM ohlcv_bars
		WHERE symbol = $1 AND timeframe = $2
		ORDER BY ts DESC
		LIMIT $3
	`, normalizeSymbol(symbol), strings.ToLower(timeframe), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bars []market.Bar
	for rows.Next() {
		var bar market.Bar
		if err := rows.Scan(&bar.Time, &bar.Open, &bar.High, &bar.Low, &bar.Close, &bar.Volume); err != nil {
			return nil, err
		}
		bars = append(bars, bar)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Slice(bars, func(i, j int) bool {
		return bars[i].Time.Before(bars[j].Time)
	})
	return bars, nil
}

func (s *PostgresStore) UpsertQuote(ctx context.Context, quote market.Quote) error {
	payload, err := json.Marshal(quote)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO latest_quotes(symbol, provider, price, provider_time, received_at, data_age_seconds,
		  exchange_name, exchange_timezone, data_granularity, provider_warning, payload_json)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT(symbol) DO UPDATE SET
		  provider = EXCLUDED.provider,
		  price = EXCLUDED.price,
		  provider_time = EXCLUDED.provider_time,
		  received_at = EXCLUDED.received_at,
		  data_age_seconds = EXCLUDED.data_age_seconds,
		  exchange_name = EXCLUDED.exchange_name,
		  exchange_timezone = EXCLUDED.exchange_timezone,
		  data_granularity = EXCLUDED.data_granularity,
		  provider_warning = EXCLUDED.provider_warning,
		  payload_json = EXCLUDED.payload_json
	`, normalizeSymbol(quote.Symbol), quote.Provider, quote.Price, quote.RegularMarketTime, quote.ReceivedAt,
		quote.DataAgeSeconds, quote.ExchangeName, quote.ExchangeTimezone, quote.DataGranularity, quote.ProviderWarning, payload)
	return err
}

func (s *PostgresStore) GetQuote(ctx context.Context, symbol string) (*market.Quote, error) {
	var payload []byte
	err := s.pool.QueryRow(ctx, `SELECT payload_json FROM latest_quotes WHERE symbol = $1`, normalizeSymbol(symbol)).Scan(&payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var quote market.Quote
	if err := json.Unmarshal(payload, &quote); err != nil {
		return nil, err
	}
	return &quote, nil
}

func (s *PostgresStore) ListQuotes(ctx context.Context) ([]market.Quote, error) {
	rows, err := s.pool.Query(ctx, `SELECT payload_json FROM latest_quotes ORDER BY symbol`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quotes []market.Quote
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var quote market.Quote
		if err := json.Unmarshal(payload, &quote); err != nil {
			return nil, err
		}
		quotes = append(quotes, quote)
	}
	return quotes, rows.Err()
}

func (s *PostgresStore) AppendQuoteSnapshot(ctx context.Context, snapshot QuoteSnapshot) error {
	snapshot.Symbol = normalizeSymbol(snapshot.Symbol)
	if snapshot.Symbol == "" || snapshot.Provider == "" || snapshot.Price <= 0 {
		return errors.New("quote snapshot requires symbol, provider, and positive price")
	}
	if snapshot.ReceivedAt.IsZero() {
		snapshot.ReceivedAt = time.Now().UTC()
	}
	if snapshot.ProviderTime.IsZero() {
		snapshot.ProviderTime = snapshot.ReceivedAt
	}
	return s.pool.QueryRow(ctx, `
		INSERT INTO market_quote_snapshots(symbol, provider, price, provider_time, received_at,
		  data_age_seconds, session, market_state, warning)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id
	`, snapshot.Symbol, snapshot.Provider, snapshot.Price, snapshot.ProviderTime, snapshot.ReceivedAt,
		snapshot.DataAgeSeconds, snapshot.Session, snapshot.MarketState, snapshot.Warning).Scan(&snapshot.ID)
}

func (s *PostgresStore) ListQuoteSnapshots(ctx context.Context, symbol, provider string, bounds TimeRange) ([]QuoteSnapshot, error) {
	query, args := selectRange(`
		SELECT id, symbol, provider, price, provider_time, received_at, data_age_seconds, session,
		  COALESCE(market_state, ''), COALESCE(warning, '')
		FROM market_quote_snapshots
	`, "received_at", normalizeSymbol(symbol), provider, bounds)
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var snapshots []QuoteSnapshot
	for rows.Next() {
		var snapshot QuoteSnapshot
		if err := rows.Scan(&snapshot.ID, &snapshot.Symbol, &snapshot.Provider, &snapshot.Price, &snapshot.ProviderTime,
			&snapshot.ReceivedAt, &snapshot.DataAgeSeconds, &snapshot.Session, &snapshot.MarketState, &snapshot.Warning); err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snapshot)
	}
	return snapshots, rows.Err()
}

func (s *PostgresStore) AppendRWASnapshot(ctx context.Context, snapshot RWASnapshot) error {
	snapshot.Symbol = normalizeSymbol(snapshot.Symbol)
	if snapshot.Symbol == "" || snapshot.Provider == "" || snapshot.Ticker == "" || snapshot.Price <= 0 {
		return errors.New("RWA snapshot requires symbol, ticker, provider, and positive price")
	}
	if snapshot.ReceivedAt.IsZero() {
		snapshot.ReceivedAt = time.Now().UTC()
	}
	return s.pool.QueryRow(ctx, `
		INSERT INTO rwa_quote_snapshots(symbol, ticker, provider, data_source, price, received_at,
		  session, provider_market_status, warning)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id
	`, snapshot.Symbol, snapshot.Ticker, snapshot.Provider, snapshot.DataSource, snapshot.Price, snapshot.ReceivedAt,
		snapshot.Session, snapshot.ProviderMarketStatus, snapshot.Warning).Scan(&snapshot.ID)
}

func (s *PostgresStore) ListRWASnapshots(ctx context.Context, symbol, provider string, bounds TimeRange) ([]RWASnapshot, error) {
	query, args := selectRange(`
		SELECT id, symbol, ticker, provider, COALESCE(data_source, ''), price, received_at, session,
		  COALESCE(provider_market_status, ''), COALESCE(warning, '')
		FROM rwa_quote_snapshots
	`, "received_at", normalizeSymbol(symbol), provider, bounds)
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var snapshots []RWASnapshot
	for rows.Next() {
		var snapshot RWASnapshot
		if err := rows.Scan(&snapshot.ID, &snapshot.Symbol, &snapshot.Ticker, &snapshot.Provider, &snapshot.DataSource,
			&snapshot.Price, &snapshot.ReceivedAt, &snapshot.Session, &snapshot.ProviderMarketStatus, &snapshot.Warning); err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snapshot)
	}
	return snapshots, rows.Err()
}

func (s *PostgresStore) UpsertRWABars(ctx context.Context, rows []RWABarSnapshot) (int, error) {
	if len(rows) == 0 {
		return 0, nil
	}
	batch := &pgx.Batch{}
	queued := 0
	for _, row := range rows {
		row.Symbol = normalizeSymbol(row.Symbol)
		if row.Symbol == "" || row.Provider == "" || row.Timeframe == "" || row.Timestamp.IsZero() {
			continue
		}
		if row.ReceivedAt.IsZero() {
			row.ReceivedAt = time.Now().UTC()
		}
		batch.Queue(`
			INSERT INTO rwa_ohlcv_bars(symbol, ticker, provider, timeframe, ts, open, high, low, close, volume, amount, received_at, session)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
			ON CONFLICT(symbol, provider, timeframe, ts) DO UPDATE SET
			  ticker = EXCLUDED.ticker, open = EXCLUDED.open, high = EXCLUDED.high, low = EXCLUDED.low,
			  close = EXCLUDED.close, volume = EXCLUDED.volume, amount = EXCLUDED.amount,
			  received_at = EXCLUDED.received_at, session = EXCLUDED.session
		`, row.Symbol, row.Ticker, row.Provider, strings.ToLower(row.Timeframe), row.Timestamp,
			row.Open, row.High, row.Low, row.Close, row.Volume, row.Amount, row.ReceivedAt, row.Session)
		queued++
	}
	if queued == 0 {
		return 0, nil
	}
	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()
	for i := 0; i < queued; i++ {
		if _, err := br.Exec(); err != nil {
			return i, err
		}
	}
	return queued, nil
}

func (s *PostgresStore) GetRWABars(ctx context.Context, symbol, timeframe string, limit int) ([]RWABarSnapshot, error) {
	if limit <= 0 {
		limit = 390
	}
	rows, err := s.pool.Query(ctx, `
		SELECT symbol, ticker, provider, timeframe, ts, open, high, low, close, volume, amount, received_at, session
		FROM rwa_ohlcv_bars
		WHERE symbol = $1 AND timeframe = $2
		ORDER BY ts DESC
		LIMIT $3
	`, normalizeSymbol(symbol), strings.ToLower(timeframe), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var bars []RWABarSnapshot
	for rows.Next() {
		var row RWABarSnapshot
		if err := rows.Scan(&row.Symbol, &row.Ticker, &row.Provider, &row.Timeframe, &row.Timestamp, &row.Open,
			&row.High, &row.Low, &row.Close, &row.Volume, &row.Amount, &row.ReceivedAt, &row.Session); err != nil {
			return nil, err
		}
		bars = append(bars, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Slice(bars, func(i, j int) bool { return bars[i].Timestamp.Before(bars[j].Timestamp) })
	return bars, nil
}

func (s *PostgresStore) SaveOptionChain(ctx context.Context, chain *options.ChainResponse) (int, error) {
	if chain == nil {
		return 0, errors.New("option chain is required")
	}
	contracts := OptionContracts(chain)
	iv := IVFromChain(chain)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	for _, row := range contracts {
		_, err := tx.Exec(ctx, `
			INSERT INTO option_contract_snapshots(symbol, provider, contract_symbol, expiration_date, strike, right_type,
			 bid, ask, mid, last_price, volume, open_interest, raw_implied_volatility,
			 resolved_implied_volatility, resolved_iv_error, iv_input_price, iv_input_source,
			 contract_trade_at, underlying_spot, underlying_provider_time, received_at, warning)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)
		`, normalizeSymbol(row.Symbol), row.Provider, row.ContractSymbol, row.ExpirationDate, row.Strike, row.Right,
			row.Bid, row.Ask, row.Mid, row.Last, row.Volume, row.OpenInterest, row.RawImpliedVolatility,
			row.ResolvedImpliedVolatility, row.ResolvedIVError, row.IVInputPrice, row.IVInputSource,
			nullTime(row.ContractTradeAt), row.UnderlyingSpot, nullTime(row.UnderlyingProviderTime), row.ReceivedAt, row.Warning)
		if err != nil {
			return 0, err
		}
	}
	if iv != nil {
		_, err := tx.Exec(ctx, `
			INSERT INTO iv_snapshots(symbol, provider, expiration_date, atm_strike, spot, call_resolved_iv,
			 put_resolved_iv, average_resolved_iv, straddle_move, straddle_move_percent, one_day_move,
			 one_day_move_percent, risk_free_rate, dividend_yield_assumption, underlying_provider_time,
			 underlying_data_age_seconds, contract_trade_data_age_seconds, received_at, warning)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
		`, normalizeSymbol(iv.Symbol), iv.Provider, iv.ExpirationDate, iv.ATMStrike, iv.Spot, iv.CallResolvedIV,
			iv.PutResolvedIV, iv.AverageResolvedIV, iv.StraddleMove, iv.StraddleMovePercent, iv.OneDayMove,
			iv.OneDayMovePercent, iv.RiskFreeRate, iv.DividendYieldAssumption, nullTime(iv.UnderlyingProviderTime),
			iv.UnderlyingDataAgeSeconds, iv.ContractTradeDataAgeSeconds, iv.ReceivedAt, iv.Warning)
		if err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return len(contracts), nil
}

func (s *PostgresStore) ListOptionContracts(ctx context.Context, symbol string, bounds TimeRange) ([]OptionContractSnapshot, error) {
	query, args := selectRange(`
		SELECT id, symbol, provider, contract_symbol, expiration_date, strike, right_type, bid, ask, mid,
		  last_price, volume, open_interest, raw_implied_volatility, resolved_implied_volatility,
		  COALESCE(resolved_iv_error, ''), iv_input_price, COALESCE(iv_input_source, ''), contract_trade_at,
		  underlying_spot, underlying_provider_time, received_at, COALESCE(warning, '')
		FROM option_contract_snapshots
	`, "received_at", normalizeSymbol(symbol), "", bounds)
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var snapshots []OptionContractSnapshot
	for rows.Next() {
		var row OptionContractSnapshot
		var contractTradeAt *time.Time
		var underlyingProviderTime *time.Time
		if err := rows.Scan(&row.ID, &row.Symbol, &row.Provider, &row.ContractSymbol, &row.ExpirationDate,
			&row.Strike, &row.Right, &row.Bid, &row.Ask, &row.Mid, &row.Last, &row.Volume, &row.OpenInterest,
			&row.RawImpliedVolatility, &row.ResolvedImpliedVolatility, &row.ResolvedIVError, &row.IVInputPrice,
			&row.IVInputSource, &contractTradeAt, &row.UnderlyingSpot, &underlyingProviderTime,
			&row.ReceivedAt, &row.Warning); err != nil {
			return nil, err
		}
		if contractTradeAt != nil {
			row.ContractTradeAt = *contractTradeAt
		}
		if underlyingProviderTime != nil {
			row.UnderlyingProviderTime = *underlyingProviderTime
		}
		snapshots = append(snapshots, row)
	}
	return snapshots, rows.Err()
}

func (s *PostgresStore) ListIVSnapshots(ctx context.Context, symbol string, bounds TimeRange) ([]IVSnapshot, error) {
	query, args := selectRange(`
		SELECT id, symbol, provider, expiration_date, atm_strike, spot, call_resolved_iv, put_resolved_iv,
		  average_resolved_iv, straddle_move, straddle_move_percent, one_day_move, one_day_move_percent,
		  risk_free_rate, dividend_yield_assumption, underlying_provider_time, underlying_data_age_seconds,
		  contract_trade_data_age_seconds, received_at, COALESCE(warning, '')
		FROM iv_snapshots
	`, "received_at", normalizeSymbol(symbol), "", bounds)
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var snapshots []IVSnapshot
	for rows.Next() {
		var row IVSnapshot
		var underlyingProviderTime *time.Time
		if err := rows.Scan(&row.ID, &row.Symbol, &row.Provider, &row.ExpirationDate, &row.ATMStrike,
			&row.Spot, &row.CallResolvedIV, &row.PutResolvedIV, &row.AverageResolvedIV, &row.StraddleMove,
			&row.StraddleMovePercent, &row.OneDayMove, &row.OneDayMovePercent, &row.RiskFreeRate,
			&row.DividendYieldAssumption, &underlyingProviderTime, &row.UnderlyingDataAgeSeconds,
			&row.ContractTradeDataAgeSeconds, &row.ReceivedAt, &row.Warning); err != nil {
			return nil, err
		}
		if underlyingProviderTime != nil {
			row.UnderlyingProviderTime = *underlyingProviderTime
		}
		snapshots = append(snapshots, row)
	}
	return snapshots, rows.Err()
}

func (s *PostgresStore) AppendFeatureSnapshot(ctx context.Context, snapshot FeatureSnapshot) error {
	snapshot.Symbol = normalizeSymbol(snapshot.Symbol)
	if snapshot.Symbol == "" {
		return errors.New("feature snapshot requires symbol")
	}
	if snapshot.ComputedAt.IsZero() {
		snapshot.ComputedAt = time.Now().UTC()
	}
	snapshot.ProviderTimes = requiredJSON(snapshot.ProviderTimes, []byte(`{}`))
	snapshot.InputFreshness = requiredJSON(snapshot.InputFreshness, []byte(`{}`))
	snapshot.Warnings = requiredJSON(snapshot.Warnings, []byte(`[]`))
	snapshot.Payload = requiredJSON(snapshot.Payload, []byte(`{}`))
	return s.pool.QueryRow(ctx, `
		INSERT INTO intraday_feature_snapshots(symbol, computed_at, provider_times_json, input_freshness_json,
		  selected_source, selected_move_pct, warnings_json, payload_json)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id
	`, snapshot.Symbol, snapshot.ComputedAt, snapshot.ProviderTimes, snapshot.InputFreshness, snapshot.SelectedSource,
		snapshot.SelectedMovePct, snapshot.Warnings, snapshot.Payload).Scan(&snapshot.ID)
}

func (s *PostgresStore) ListFeatureSnapshots(ctx context.Context, symbol string, bounds TimeRange) ([]FeatureSnapshot, error) {
	query, args := selectRange(`
		SELECT id, symbol, computed_at, provider_times_json, input_freshness_json, COALESCE(selected_source, ''),
		  selected_move_pct, warnings_json, payload_json
		FROM intraday_feature_snapshots
	`, "computed_at", normalizeSymbol(symbol), "", bounds)
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var snapshots []FeatureSnapshot
	for rows.Next() {
		var row FeatureSnapshot
		if err := rows.Scan(&row.ID, &row.Symbol, &row.ComputedAt, &row.ProviderTimes, &row.InputFreshness,
			&row.SelectedSource, &row.SelectedMovePct, &row.Warnings, &row.Payload); err != nil {
			return nil, err
		}
		snapshots = append(snapshots, row)
	}
	return snapshots, rows.Err()
}

func (s *PostgresStore) UpsertDailySummary(ctx context.Context, summary DailyMarketSummary) error {
	summary.Symbol = normalizeSymbol(summary.Symbol)
	if summary.Symbol == "" || summary.Provider == "" || summary.TradingDate.IsZero() {
		return errors.New("daily summary requires symbol, provider, and trading date")
	}
	if summary.UpdatedAt.IsZero() {
		summary.UpdatedAt = time.Now().UTC()
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO daily_market_summaries(symbol, trading_date, provider, previous_close, premarket_high,
		  premarket_low, preopen_five_minute, regular_open, daily_high, daily_low, selected_move,
		  selected_move_pct, selected_source, updated_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		ON CONFLICT(symbol, trading_date, provider) DO UPDATE SET
		  previous_close = CASE WHEN EXCLUDED.previous_close > 0 THEN EXCLUDED.previous_close ELSE daily_market_summaries.previous_close END,
		  premarket_high = GREATEST(daily_market_summaries.premarket_high, EXCLUDED.premarket_high),
		  premarket_low = CASE
		    WHEN daily_market_summaries.premarket_low = 0 THEN EXCLUDED.premarket_low
		    WHEN EXCLUDED.premarket_low = 0 THEN daily_market_summaries.premarket_low
		    ELSE LEAST(daily_market_summaries.premarket_low, EXCLUDED.premarket_low)
		  END,
		  preopen_five_minute = CASE WHEN EXCLUDED.preopen_five_minute > 0 THEN EXCLUDED.preopen_five_minute ELSE daily_market_summaries.preopen_five_minute END,
		  regular_open = CASE WHEN EXCLUDED.regular_open > 0 THEN EXCLUDED.regular_open ELSE daily_market_summaries.regular_open END,
		  daily_high = GREATEST(daily_market_summaries.daily_high, EXCLUDED.daily_high),
		  daily_low = CASE
		    WHEN daily_market_summaries.daily_low = 0 THEN EXCLUDED.daily_low
		    WHEN EXCLUDED.daily_low = 0 THEN daily_market_summaries.daily_low
		    ELSE LEAST(daily_market_summaries.daily_low, EXCLUDED.daily_low)
		  END,
		  selected_move = CASE
		    WHEN EXCLUDED.selected_source = 'unavailable' THEN 0
		    WHEN EXCLUDED.selected_move > 0 THEN EXCLUDED.selected_move
		    ELSE daily_market_summaries.selected_move
		  END,
		  selected_move_pct = CASE
		    WHEN EXCLUDED.selected_source = 'unavailable' THEN 0
		    WHEN EXCLUDED.selected_move_pct > 0 THEN EXCLUDED.selected_move_pct
		    ELSE daily_market_summaries.selected_move_pct
		  END,
		  selected_source = COALESCE(NULLIF(EXCLUDED.selected_source, ''), daily_market_summaries.selected_source),
		  updated_at = EXCLUDED.updated_at
	`, summary.Symbol, summary.TradingDate.UTC(), summary.Provider, summary.PreviousClose, summary.PremarketHigh,
		summary.PremarketLow, summary.PreOpenFiveMinute, summary.RegularOpen, summary.DailyHigh, summary.DailyLow,
		summary.SelectedMove, summary.SelectedMovePct, summary.SelectedSource, summary.UpdatedAt)
	return err
}

func (s *PostgresStore) GetDailySummary(ctx context.Context, symbol string, tradingDate time.Time) (*DailyMarketSummary, error) {
	var summary DailyMarketSummary
	err := s.pool.QueryRow(ctx, `
		SELECT symbol, trading_date, provider, previous_close, premarket_high, premarket_low,
		  preopen_five_minute, regular_open, daily_high, daily_low, selected_move, selected_move_pct,
		  COALESCE(selected_source, ''), updated_at
		FROM daily_market_summaries
		WHERE symbol = $1 AND trading_date = $2
		ORDER BY updated_at DESC
		LIMIT 1
	`, normalizeSymbol(symbol), dateOnlyUTC(tradingDate)).Scan(&summary.Symbol, &summary.TradingDate, &summary.Provider,
		&summary.PreviousClose, &summary.PremarketHigh, &summary.PremarketLow, &summary.PreOpenFiveMinute,
		&summary.RegularOpen, &summary.DailyHigh, &summary.DailyLow, &summary.SelectedMove, &summary.SelectedMovePct,
		&summary.SelectedSource, &summary.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

func (s *PostgresStore) ListDailySummaries(ctx context.Context, symbol string, bounds TimeRange) ([]DailyMarketSummary, error) {
	conditions := []string{"1=1"}
	args := []any{}
	if symbol != "" {
		args = append(args, normalizeSymbol(symbol))
		conditions = append(conditions, fmt.Sprintf("symbol = $%d", len(args)))
	}
	if !bounds.From.IsZero() {
		args = append(args, dateOnlyUTC(bounds.From))
		conditions = append(conditions, fmt.Sprintf("trading_date >= $%d", len(args)))
	}
	if !bounds.To.IsZero() {
		args = append(args, dateOnlyUTC(bounds.To))
		conditions = append(conditions, fmt.Sprintf("trading_date <= $%d", len(args)))
	}
	limit := bounds.Limit
	if limit <= 0 {
		limit = 1000
	}
	args = append(args, limit)
	query := `SELECT symbol, trading_date, provider, previous_close, premarket_high, premarket_low,
	  preopen_five_minute, regular_open, daily_high, daily_low, selected_move, selected_move_pct,
	  COALESCE(selected_source, ''), updated_at
	  FROM daily_market_summaries WHERE ` + strings.Join(conditions, " AND ") + fmt.Sprintf(" ORDER BY trading_date ASC, updated_at ASC LIMIT $%d", len(args))
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var summaries []DailyMarketSummary
	for rows.Next() {
		var summary DailyMarketSummary
		if err := rows.Scan(&summary.Symbol, &summary.TradingDate, &summary.Provider, &summary.PreviousClose,
			&summary.PremarketHigh, &summary.PremarketLow, &summary.PreOpenFiveMinute, &summary.RegularOpen,
			&summary.DailyHigh, &summary.DailyLow, &summary.SelectedMove, &summary.SelectedMovePct,
			&summary.SelectedSource, &summary.UpdatedAt); err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}
	return summaries, rows.Err()
}

func (s *PostgresStore) AddCollectionRun(ctx context.Context, run CollectionRun) error {
	if run.Collector == "" || run.Provider == "" {
		return errors.New("collection run requires collector and provider")
	}
	if run.StartedAt.IsZero() {
		run.StartedAt = time.Now().UTC()
	}
	if run.FinishedAt.IsZero() {
		run.FinishedAt = run.StartedAt
	}
	return s.pool.QueryRow(ctx, `
		INSERT INTO collection_runs(collector, provider, started_at, finished_at, symbols_attempted,
		  symbols_succeeded, records_written, rate_limited, backoff_seconds, error_summary)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id
	`, run.Collector, run.Provider, run.StartedAt, run.FinishedAt, run.SymbolsAttempted, run.SymbolsSucceeded,
		run.RecordsWritten, run.RateLimited, run.BackoffSeconds, run.ErrorSummary).Scan(&run.ID)
}

func (s *PostgresStore) ListCollectionRuns(ctx context.Context, limit int) ([]CollectionRun, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, collector, provider, started_at, finished_at, symbols_attempted, symbols_succeeded,
		  records_written, rate_limited, backoff_seconds, COALESCE(error_summary, '')
		FROM collection_runs
		ORDER BY finished_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var runs []CollectionRun
	for rows.Next() {
		var run CollectionRun
		if err := rows.Scan(&run.ID, &run.Collector, &run.Provider, &run.StartedAt, &run.FinishedAt,
			&run.SymbolsAttempted, &run.SymbolsSucceeded, &run.RecordsWritten, &run.RateLimited,
			&run.BackoffSeconds, &run.ErrorSummary); err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

func (s *PostgresStore) AddEventIngestionRun(ctx context.Context, run EventIngestionRun) (EventIngestionRun, error) {
	if run.Provider == "" || run.Query == "" || run.Scope == "" {
		return run, errors.New("event ingestion run requires provider, query, and scope")
	}
	if run.StartedAt.IsZero() {
		run.StartedAt = time.Now().UTC()
	}
	if run.FinishedAt.IsZero() {
		run.FinishedAt = run.StartedAt
	}
	run.RequestedSymbols = defaultJSON(run.RequestedSymbols, []byte("[]"))
	run.Metadata = defaultJSON(run.Metadata, []byte("{}"))
	err := s.pool.QueryRow(ctx, `
		INSERT INTO event_ingestion_runs(provider, query, scope, group_name, requested_symbols,
		  strict_ticker, provider_route, started_at, finished_at, result_count, records_written,
		  rate_limited, backoff_seconds, error_summary, metadata_json)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		RETURNING id
	`, run.Provider, run.Query, run.Scope, nullableString(run.GroupName), run.RequestedSymbols,
		run.StrictTicker, nullableString(run.ProviderRoute), run.StartedAt, run.FinishedAt, run.ResultCount,
		run.RecordsWritten, run.RateLimited, run.BackoffSeconds, nullableString(run.ErrorSummary), run.Metadata).Scan(&run.ID)
	return run, err
}

func (s *PostgresStore) ListEventIngestionRuns(ctx context.Context, limit int) ([]EventIngestionRun, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, provider, query, scope, COALESCE(group_name, ''), requested_symbols, strict_ticker,
		  COALESCE(provider_route, ''), started_at, finished_at, result_count, records_written,
		  rate_limited, backoff_seconds, COALESCE(error_summary, ''), metadata_json
		FROM event_ingestion_runs
		ORDER BY finished_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []EventIngestionRun
	for rows.Next() {
		var run EventIngestionRun
		if err := rows.Scan(&run.ID, &run.Provider, &run.Query, &run.Scope, &run.GroupName, &run.RequestedSymbols,
			&run.StrictTicker, &run.ProviderRoute, &run.StartedAt, &run.FinishedAt, &run.ResultCount,
			&run.RecordsWritten, &run.RateLimited, &run.BackoffSeconds, &run.ErrorSummary, &run.Metadata); err != nil {
			return nil, err
		}
		result = append(result, run)
	}
	return result, rows.Err()
}

func (s *PostgresStore) SaveEventEvidence(ctx context.Context, runID int64, rows []EventEvidence) (int, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	written := 0
	for _, row := range rows {
		if row.CanonicalURL == "" || row.HeadlineHash == "" || row.Title == "" || row.Provider == "" || row.Scope == "" || row.EventType == "" {
			return written, errors.New("event evidence requires URL, headline hash, title, provider, scope, and event type")
		}
		if row.ReceivedAt.IsZero() {
			row.ReceivedAt = time.Now().UTC()
		}
		row.MatchedSymbols = defaultJSON(row.MatchedSymbols, []byte("[]"))
		row.Metadata = defaultJSON(row.Metadata, []byte("{}"))
		var evidenceID int64
		err := tx.QueryRow(ctx, `
			INSERT INTO event_evidence(ingestion_run_id, canonical_url, headline_hash, title, snippet,
			  source_domain, provider, publisher, source_tier, published_at, discovered_at, received_at,
			  freshness, matched_symbols, scope, group_name, event_type, metadata_json, warning)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
			ON CONFLICT(canonical_url) DO NOTHING
			RETURNING id
		`, nullableID(runID), row.CanonicalURL, row.HeadlineHash, row.Title, nullableString(row.Snippet),
			nullableString(row.SourceDomain), row.Provider, nullableString(row.Publisher), row.SourceTier,
			nullableTime(row.PublishedAt), nullableTime(row.DiscoveredAt), row.ReceivedAt, nullableString(row.Freshness),
			row.MatchedSymbols, row.Scope, nullableString(row.GroupName), row.EventType, row.Metadata, nullableString(row.Warning)).Scan(&evidenceID)
		if errors.Is(err, pgx.ErrNoRows) {
			if err = tx.QueryRow(ctx, `SELECT id FROM event_evidence WHERE canonical_url = $1`, row.CanonicalURL).Scan(&evidenceID); err != nil {
				return written, err
			}
		} else if err != nil {
			return written, err
		} else {
			written++
		}

		key, primarySymbol := marketEventKey(row)
		var eventID int64
		if err := tx.QueryRow(ctx, `
			INSERT INTO market_events(event_key, scope, primary_symbol, group_name, event_type,
			  evidence_status, severity, state, first_seen_at, last_seen_at)
			VALUES($1,$2,$3,$4,$5,$6,'unknown','active',$7,$7)
			ON CONFLICT(event_key) DO UPDATE SET
			  last_seen_at = GREATEST(market_events.last_seen_at, EXCLUDED.last_seen_at),
			  evidence_status = CASE WHEN EXCLUDED.evidence_status = 'supported' THEN 'supported' ELSE market_events.evidence_status END
			RETURNING id
		`, key, row.Scope, nullableString(primarySymbol), nullableString(row.GroupName), row.EventType,
			evidenceStatus(row.SourceTier), row.ReceivedAt).Scan(&eventID); err != nil {
			return written, err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO market_event_evidence(market_event_id, evidence_id, relation_type, confidence)
			VALUES($1,$2,'supports',$3)
			ON CONFLICT(market_event_id, evidence_id) DO NOTHING
		`, eventID, evidenceID, sourceTierConfidence(row.SourceTier)); err != nil {
			return written, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return written, err
	}
	return written, nil
}

func (s *PostgresStore) ListEventEvidence(ctx context.Context, filter EventFilter) ([]EventEvidence, error) {
	conditions := []string{"TRUE"}
	args := make([]any, 0, 4)
	add := func(condition string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(condition, len(args)))
	}
	if filter.Symbol != "" {
		add("matched_symbols ? $%d", normalizeSymbol(filter.Symbol))
	}
	if filter.Group != "" {
		add("group_name = $%d", filter.Group)
	}
	if !filter.Bounds.From.IsZero() {
		add("received_at >= $%d", filter.Bounds.From)
	}
	if !filter.Bounds.To.IsZero() {
		add("received_at <= $%d", filter.Bounds.To)
	}
	limit := filter.Bounds.Limit
	if limit <= 0 {
		limit = 1000
	}
	args = append(args, limit)
	innerQuery := `SELECT id, COALESCE(ingestion_run_id, 0), canonical_url, headline_hash, title, COALESCE(snippet, ''),
	  COALESCE(source_domain, ''), provider, COALESCE(publisher, ''), source_tier, published_at, discovered_at,
	  received_at, COALESCE(freshness, ''), matched_symbols, scope, COALESCE(group_name, ''), event_type,
	  metadata_json, COALESCE(warning, '')
	  FROM event_evidence WHERE ` + strings.Join(conditions, " AND ")
	query := "SELECT * FROM (" + innerQuery + fmt.Sprintf(" ORDER BY received_at DESC LIMIT $%d", len(args)) + ") latest ORDER BY received_at ASC"
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []EventEvidence
	for rows.Next() {
		var evidence EventEvidence
		var publishedAt, discoveredAt *time.Time
		if err := rows.Scan(&evidence.ID, &evidence.IngestionRunID, &evidence.CanonicalURL, &evidence.HeadlineHash,
			&evidence.Title, &evidence.Snippet, &evidence.SourceDomain, &evidence.Provider, &evidence.Publisher,
			&evidence.SourceTier, &publishedAt, &discoveredAt, &evidence.ReceivedAt, &evidence.Freshness,
			&evidence.MatchedSymbols, &evidence.Scope, &evidence.GroupName, &evidence.EventType, &evidence.Metadata,
			&evidence.Warning); err != nil {
			return nil, err
		}
		if publishedAt != nil {
			evidence.PublishedAt = *publishedAt
		}
		if discoveredAt != nil {
			evidence.DiscoveredAt = *discoveredAt
		}
		result = append(result, evidence)
	}
	return result, rows.Err()
}

func (s *PostgresStore) ListMarketEvents(ctx context.Context, filter EventFilter) ([]MarketEvent, error) {
	conditions := []string{"TRUE"}
	args := make([]any, 0, 4)
	add := func(condition string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(condition, len(args)))
	}
	if filter.Symbol != "" {
		args = append(args, normalizeSymbol(filter.Symbol))
		conditions = append(conditions, fmt.Sprintf("(primary_symbol = $%d OR scope IN ('macro', 'market'))", len(args)))
	}
	if filter.Group != "" {
		args = append(args, filter.Group)
		conditions = append(conditions, fmt.Sprintf("(scope <> 'group' OR group_name = $%d)", len(args)))
	}
	if !filter.Bounds.From.IsZero() {
		add("last_seen_at >= $%d", filter.Bounds.From)
	}
	if !filter.Bounds.To.IsZero() {
		add("last_seen_at <= $%d", filter.Bounds.To)
	}
	limit := filter.Bounds.Limit
	if limit <= 0 {
		limit = 1000
	}
	args = append(args, limit)
	innerQuery := `SELECT id, event_key, scope, COALESCE(primary_symbol, ''), COALESCE(group_name, ''), event_type,
	  evidence_status, severity, state, first_seen_at, last_seen_at
	  FROM market_events WHERE ` + strings.Join(conditions, " AND ")
	query := "SELECT * FROM (" + innerQuery + fmt.Sprintf(" ORDER BY last_seen_at DESC LIMIT $%d", len(args)) + ") latest ORDER BY last_seen_at ASC"
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []MarketEvent
	for rows.Next() {
		var event MarketEvent
		if err := rows.Scan(&event.ID, &event.EventKey, &event.Scope, &event.PrimarySymbol, &event.GroupName,
			&event.EventType, &event.EvidenceStatus, &event.Severity, &event.State, &event.FirstSeenAt, &event.LastSeenAt); err != nil {
			return nil, err
		}
		result = append(result, event)
	}
	return result, rows.Err()
}

func (s *PostgresStore) AppendEventContext(ctx context.Context, snapshot EventContextSnapshot) (EventContextSnapshot, error) {
	snapshot.Symbol = normalizeSymbol(snapshot.Symbol)
	if snapshot.Symbol == "" || snapshot.AsOf.IsZero() || snapshot.EvidenceCoverage == "" || snapshot.RiskState == "" {
		return snapshot, errors.New("event context requires symbol, asOf, coverage, and risk state")
	}
	if snapshot.CreatedAt.IsZero() {
		snapshot.CreatedAt = time.Now().UTC()
	}
	snapshot.KnownEventIDs = defaultJSON(snapshot.KnownEventIDs, []byte("[]"))
	snapshot.ReasonCodes = defaultJSON(snapshot.ReasonCodes, []byte("[]"))
	snapshot.Warnings = defaultJSON(snapshot.Warnings, []byte("[]"))
	snapshot.Payload = defaultJSON(snapshot.Payload, []byte("{}"))
	err := s.pool.QueryRow(ctx, `
		INSERT INTO event_context_snapshots(symbol, as_of, market_data_fresh, quote_age_seconds, session,
		  selected_move_pct, evidence_coverage, last_ingestion_at, known_event_ids_json, risk_state,
		  reason_codes_json, warnings_json, payload_json, created_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		RETURNING id
	`, snapshot.Symbol, snapshot.AsOf, snapshot.MarketDataFresh, snapshot.QuoteAgeSeconds, snapshot.Session,
		snapshot.SelectedMovePct, snapshot.EvidenceCoverage, nullableTime(snapshot.LastIngestionAt), snapshot.KnownEventIDs,
		snapshot.RiskState, snapshot.ReasonCodes, snapshot.Warnings, snapshot.Payload, snapshot.CreatedAt).Scan(&snapshot.ID)
	return snapshot, err
}

func (s *PostgresStore) GetEventContext(ctx context.Context, symbol string, asOf time.Time) (*EventContextSnapshot, error) {
	conditions := []string{"symbol = $1"}
	args := []any{normalizeSymbol(symbol)}
	if !asOf.IsZero() {
		args = append(args, asOf)
		conditions = append(conditions, "$2 >= as_of")
	}
	query := `SELECT id, symbol, as_of, market_data_fresh, quote_age_seconds, session, selected_move_pct,
	  evidence_coverage, last_ingestion_at, known_event_ids_json, risk_state, reason_codes_json, warnings_json,
	  payload_json, created_at
	  FROM event_context_snapshots WHERE ` + strings.Join(conditions, " AND ") + " ORDER BY as_of DESC LIMIT 1"
	var snapshot EventContextSnapshot
	var lastIngestion *time.Time
	err := s.pool.QueryRow(ctx, query, args...).Scan(&snapshot.ID, &snapshot.Symbol, &snapshot.AsOf, &snapshot.MarketDataFresh,
		&snapshot.QuoteAgeSeconds, &snapshot.Session, &snapshot.SelectedMovePct, &snapshot.EvidenceCoverage, &lastIngestion,
		&snapshot.KnownEventIDs, &snapshot.RiskState, &snapshot.ReasonCodes, &snapshot.Warnings, &snapshot.Payload, &snapshot.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if lastIngestion != nil {
		snapshot.LastIngestionAt = *lastIngestion
	}
	return &snapshot, nil
}

func (s *PostgresStore) AppendShadowSignal(ctx context.Context, observation ShadowSignalObservation) (ShadowSignalObservation, error) {
	observation.Symbol = normalizeSymbol(observation.Symbol)
	if observation.SignalType == "" || observation.Symbol == "" || observation.AsOf.IsZero() || observation.State == "" || observation.ThresholdVersion == "" {
		return observation, errors.New("shadow signal requires type, symbol, asOf, state, and threshold version")
	}
	if observation.CreatedAt.IsZero() {
		observation.CreatedAt = time.Now().UTC()
	}
	observation.ReasonCodes = defaultJSON(observation.ReasonCodes, []byte("[]"))
	observation.Metrics = defaultJSON(observation.Metrics, []byte("{}"))
	err := s.pool.QueryRow(ctx, `
		INSERT INTO shadow_signal_observations(signal_type, symbol, as_of, state, eligible, blocked_reason,
		  threshold_version, event_context_id, feature_snapshot_id, reason_codes_json, metrics_json, created_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id
	`, observation.SignalType, observation.Symbol, observation.AsOf, observation.State, observation.Eligible,
		nullableString(observation.BlockedReason), observation.ThresholdVersion, nullableID(observation.EventContextID),
		nullableID(observation.FeatureSnapshotID), observation.ReasonCodes, observation.Metrics, observation.CreatedAt).Scan(&observation.ID)
	return observation, err
}

func (s *PostgresStore) ListShadowSignals(ctx context.Context, symbol string, bounds TimeRange) ([]ShadowSignalObservation, error) {
	conditions := []string{"symbol = $1"}
	args := []any{normalizeSymbol(symbol)}
	if !bounds.From.IsZero() {
		args = append(args, bounds.From)
		conditions = append(conditions, fmt.Sprintf("as_of >= $%d", len(args)))
	}
	if !bounds.To.IsZero() {
		args = append(args, bounds.To)
		conditions = append(conditions, fmt.Sprintf("as_of <= $%d", len(args)))
	}
	limit := bounds.Limit
	if limit <= 0 {
		limit = 1000
	}
	args = append(args, limit)
	innerQuery := `SELECT id, signal_type, symbol, as_of, state, eligible, COALESCE(blocked_reason, ''), threshold_version,
	  COALESCE(event_context_id, 0), COALESCE(feature_snapshot_id, 0), reason_codes_json, metrics_json, created_at
	  FROM shadow_signal_observations WHERE ` + strings.Join(conditions, " AND ")
	query := "SELECT * FROM (" + innerQuery + fmt.Sprintf(" ORDER BY as_of DESC LIMIT $%d", len(args)) + ") latest ORDER BY as_of ASC"
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []ShadowSignalObservation
	for rows.Next() {
		var observation ShadowSignalObservation
		if err := rows.Scan(&observation.ID, &observation.SignalType, &observation.Symbol, &observation.AsOf,
			&observation.State, &observation.Eligible, &observation.BlockedReason, &observation.ThresholdVersion,
			&observation.EventContextID, &observation.FeatureSnapshotID, &observation.ReasonCodes, &observation.Metrics,
			&observation.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, observation)
	}
	return result, rows.Err()
}

func (s *PostgresStore) ListSymbolProfiles(ctx context.Context) ([]SymbolProfile, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT symbol, role, horizon, action_permissions_json, cost_basis, COALESCE(notes, ''), updated_at
		FROM symbol_profiles
		ORDER BY symbol
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	profiles := make([]SymbolProfile, 0)
	for rows.Next() {
		var profile SymbolProfile
		if err := rows.Scan(&profile.Symbol, &profile.Role, &profile.Horizon, &profile.ActionPermissions,
			&profile.CostBasis, &profile.Notes, &profile.UpdatedAt); err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}
	return profiles, rows.Err()
}

func (s *PostgresStore) GetSymbolProfile(ctx context.Context, symbol string) (*SymbolProfile, error) {
	var profile SymbolProfile
	err := s.pool.QueryRow(ctx, `
		SELECT symbol, role, horizon, action_permissions_json, cost_basis, COALESCE(notes, ''), updated_at
		FROM symbol_profiles WHERE symbol = $1
	`, normalizeSymbol(symbol)).Scan(&profile.Symbol, &profile.Role, &profile.Horizon, &profile.ActionPermissions,
		&profile.CostBasis, &profile.Notes, &profile.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (s *PostgresStore) UpsertSymbolProfile(ctx context.Context, profile SymbolProfile) (SymbolProfile, error) {
	profile, err := normalizeSymbolProfile(profile)
	if err != nil {
		return profile, err
	}
	err = s.pool.QueryRow(ctx, `
		INSERT INTO symbol_profiles(symbol, role, horizon, action_permissions_json, cost_basis, notes, updated_at)
		VALUES($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT(symbol) DO UPDATE SET
		  role = EXCLUDED.role, horizon = EXCLUDED.horizon,
		  action_permissions_json = EXCLUDED.action_permissions_json, cost_basis = EXCLUDED.cost_basis,
		  notes = EXCLUDED.notes, updated_at = EXCLUDED.updated_at
		RETURNING symbol, role, horizon, action_permissions_json, cost_basis, COALESCE(notes, ''), updated_at
	`, profile.Symbol, profile.Role, profile.Horizon, profile.ActionPermissions, profile.CostBasis,
		nullableString(profile.Notes), profile.UpdatedAt).Scan(&profile.Symbol, &profile.Role, &profile.Horizon,
		&profile.ActionPermissions, &profile.CostBasis, &profile.Notes, &profile.UpdatedAt)
	return profile, err
}

func (s *PostgresStore) CreatePeerMapVersion(ctx context.Context, peerMap PeerMapVersion) (PeerMapVersion, error) {
	peerMap, err := normalizePeerMapVersion(peerMap)
	if err != nil {
		return peerMap, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return peerMap, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err = tx.Exec(ctx, `LOCK TABLE peer_map_versions IN SHARE ROW EXCLUSIVE MODE`); err != nil {
		return peerMap, err
	}
	var latest int
	if err = tx.QueryRow(ctx, `SELECT COALESCE(MAX(version), 0) FROM peer_map_versions WHERE map_key = $1`, peerMap.MapKey).Scan(&latest); err != nil {
		return peerMap, err
	}
	if peerMap.Version <= latest {
		peerMap.Version = latest + 1
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO peer_map_versions(map_key, version, group_name, benchmark_symbol, methodology, review_status, effective_from, created_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id
	`, peerMap.MapKey, peerMap.Version, nullableString(peerMap.GroupName), nullableString(peerMap.BenchmarkSymbol),
		peerMap.Methodology, peerMap.ReviewStatus, peerMap.EffectiveFrom, peerMap.CreatedAt).Scan(&peerMap.ID)
	if err != nil {
		return peerMap, err
	}
	for _, member := range peerMap.Members {
		if _, err = tx.Exec(ctx, `
			INSERT INTO peer_map_members(peer_map_version_id, symbol, relation, weight)
			VALUES($1,$2,$3,$4)
		`, peerMap.ID, member.Symbol, member.Relation, member.Weight); err != nil {
			return peerMap, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return peerMap, err
	}
	return peerMap, nil
}

func (s *PostgresStore) ListPeerMapVersions(ctx context.Context, mapKey string) ([]PeerMapVersion, error) {
	conditions := ""
	args := []any{}
	if mapKey != "" {
		args = append(args, normalizeMapKey(mapKey))
		conditions = "WHERE map_key = $1"
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, map_key, version, COALESCE(group_name, ''), COALESCE(benchmark_symbol, ''), methodology,
		  review_status, effective_from, created_at
		FROM peer_map_versions `+conditions+` ORDER BY map_key, version
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	maps := make([]PeerMapVersion, 0)
	for rows.Next() {
		var peerMap PeerMapVersion
		if err := rows.Scan(&peerMap.ID, &peerMap.MapKey, &peerMap.Version, &peerMap.GroupName,
			&peerMap.BenchmarkSymbol, &peerMap.Methodology, &peerMap.ReviewStatus, &peerMap.EffectiveFrom, &peerMap.CreatedAt); err != nil {
			return nil, err
		}
		members, err := s.peerMapMembers(ctx, peerMap.ID)
		if err != nil {
			return nil, err
		}
		peerMap.Members = members
		maps = append(maps, peerMap)
	}
	return maps, rows.Err()
}

func (s *PostgresStore) GetActivePeerMap(ctx context.Context, symbol, group string, asOf time.Time) (*PeerMapVersion, error) {
	if asOf.IsZero() {
		asOf = time.Now().UTC()
	}
	symbol = normalizeSymbol(symbol)
	var peerMap PeerMapVersion
	err := s.pool.QueryRow(ctx, `
		SELECT v.id, v.map_key, v.version, COALESCE(v.group_name, ''), COALESCE(v.benchmark_symbol, ''),
		  v.methodology, v.review_status, v.effective_from, v.created_at,
		  EXISTS(SELECT 1 FROM peer_map_members own WHERE own.peer_map_version_id = v.id AND own.symbol = $1) AS specific
		FROM peer_map_versions v
		WHERE v.review_status = 'approved' AND v.effective_from <= $3
		  AND (v.group_name = $2 OR EXISTS(SELECT 1 FROM peer_map_members member WHERE member.peer_map_version_id = v.id AND member.symbol = $1))
		ORDER BY specific DESC, v.effective_from DESC, v.version DESC
		LIMIT 1
	`, symbol, strings.TrimSpace(group), asOf).Scan(&peerMap.ID, &peerMap.MapKey, &peerMap.Version, &peerMap.GroupName,
		&peerMap.BenchmarkSymbol, &peerMap.Methodology, &peerMap.ReviewStatus, &peerMap.EffectiveFrom, &peerMap.CreatedAt, new(bool))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	members, err := s.peerMapMembers(ctx, peerMap.ID)
	if err != nil {
		return nil, err
	}
	peerMap.Members = members
	return &peerMap, nil
}

func (s *PostgresStore) AppendPolicyCandidate(ctx context.Context, candidate PolicyCandidateAlert) (PolicyCandidateAlert, error) {
	candidate, err := normalizePolicyCandidate(candidate)
	if err != nil {
		return candidate, err
	}
	err = s.pool.QueryRow(ctx, `
		INSERT INTO policy_candidate_alerts(symbol, trading_date, as_of, profile_role, horizon, session_product,
		  state, action_candidate, direction, reference_price, eligible, would_notify, budget_slot, delivery_mode,
		  suppressed_reason, threshold_version, event_context_id, peer_map_version_id, reason_codes_json,
		  metrics_json, warnings_json, created_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)
		RETURNING id
	`, candidate.Symbol, candidate.TradingDate, candidate.AsOf, candidate.ProfileRole, candidate.Horizon,
		candidate.SessionProduct, candidate.State, candidate.ActionCandidate, nullableString(candidate.Direction),
		candidate.ReferencePrice, candidate.Eligible, candidate.WouldNotify, nullableInt(candidate.BudgetSlot),
		candidate.DeliveryMode, nullableString(candidate.SuppressedReason), candidate.ThresholdVersion,
		nullableID(candidate.EventContextID), nullableID(candidate.PeerMapVersionID), candidate.ReasonCodes,
		candidate.Metrics, candidate.Warnings, candidate.CreatedAt).Scan(&candidate.ID)
	return candidate, err
}

func (s *PostgresStore) GetPolicyCandidate(ctx context.Context, id int64) (*PolicyCandidateAlert, error) {
	var candidate PolicyCandidateAlert
	var budgetSlot *int
	err := s.pool.QueryRow(ctx, `
		SELECT id, symbol, trading_date, as_of, profile_role, horizon, session_product, state, action_candidate,
		  COALESCE(direction, ''), reference_price, eligible, would_notify, budget_slot, delivery_mode,
		  COALESCE(suppressed_reason, ''), threshold_version, COALESCE(event_context_id, 0),
		  COALESCE(peer_map_version_id, 0), reason_codes_json, metrics_json, warnings_json, created_at
		FROM policy_candidate_alerts WHERE id = $1
	`, id).Scan(&candidate.ID, &candidate.Symbol, &candidate.TradingDate, &candidate.AsOf, &candidate.ProfileRole,
		&candidate.Horizon, &candidate.SessionProduct, &candidate.State, &candidate.ActionCandidate, &candidate.Direction,
		&candidate.ReferencePrice, &candidate.Eligible, &candidate.WouldNotify, &budgetSlot, &candidate.DeliveryMode,
		&candidate.SuppressedReason, &candidate.ThresholdVersion, &candidate.EventContextID, &candidate.PeerMapVersionID,
		&candidate.ReasonCodes, &candidate.Metrics, &candidate.Warnings, &candidate.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if budgetSlot != nil {
		candidate.BudgetSlot = *budgetSlot
	}
	return &candidate, nil
}

func (s *PostgresStore) ListPolicyCandidates(ctx context.Context, filter PolicyCandidateFilter) ([]PolicyCandidateAlert, error) {
	conditions := []string{"1=1"}
	args := []any{}
	add := func(condition string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(condition, len(args)))
	}
	if filter.Symbol != "" {
		add("symbol = $%d", normalizeSymbol(filter.Symbol))
	}
	if !filter.TradingDate.IsZero() {
		add("trading_date = $%d", dateOnlyUTC(filter.TradingDate))
	}
	if filter.OnlyWouldNotify {
		conditions = append(conditions, "would_notify = TRUE")
	}
	if !filter.Bounds.From.IsZero() {
		add("as_of >= $%d", filter.Bounds.From)
	}
	if !filter.Bounds.To.IsZero() {
		add("as_of <= $%d", filter.Bounds.To)
	}
	limit := filter.Bounds.Limit
	if limit <= 0 {
		limit = 1000
	}
	args = append(args, limit)
	query := `SELECT id, symbol, trading_date, as_of, profile_role, horizon, session_product, state,
	  action_candidate, COALESCE(direction, ''), reference_price, eligible, would_notify, budget_slot,
	  delivery_mode, COALESCE(suppressed_reason, ''), threshold_version, COALESCE(event_context_id, 0),
	  COALESCE(peer_map_version_id, 0), reason_codes_json, metrics_json, warnings_json, created_at
	  FROM policy_candidate_alerts WHERE ` + strings.Join(conditions, " AND ") + fmt.Sprintf(" ORDER BY as_of ASC LIMIT $%d", len(args))
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var candidates []PolicyCandidateAlert
	for rows.Next() {
		var candidate PolicyCandidateAlert
		var budgetSlot *int
		if err := rows.Scan(&candidate.ID, &candidate.Symbol, &candidate.TradingDate, &candidate.AsOf,
			&candidate.ProfileRole, &candidate.Horizon, &candidate.SessionProduct, &candidate.State,
			&candidate.ActionCandidate, &candidate.Direction, &candidate.ReferencePrice, &candidate.Eligible,
			&candidate.WouldNotify, &budgetSlot, &candidate.DeliveryMode, &candidate.SuppressedReason,
			&candidate.ThresholdVersion, &candidate.EventContextID, &candidate.PeerMapVersionID,
			&candidate.ReasonCodes, &candidate.Metrics, &candidate.Warnings, &candidate.CreatedAt); err != nil {
			return nil, err
		}
		if budgetSlot != nil {
			candidate.BudgetSlot = *budgetSlot
		}
		candidates = append(candidates, candidate)
	}
	return candidates, rows.Err()
}

func (s *PostgresStore) UpsertPolicyCandidateOutcome(ctx context.Context, outcome PolicyCandidateOutcome) (PolicyCandidateOutcome, error) {
	if outcome.CandidateAlertID == 0 || outcome.HorizonMinutes <= 0 || outcome.TargetAt.IsZero() || outcome.ObservedAt.IsZero() || outcome.SampleCount <= 0 {
		return outcome, errors.New("policy outcome requires candidate, horizon, target, observation, and samples")
	}
	if outcome.EvaluatedAt.IsZero() {
		outcome.EvaluatedAt = time.Now().UTC()
	}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO policy_candidate_outcomes(candidate_alert_id, horizon_minutes, target_at, observed_at,
		  observed_price, return_percent, max_up_percent, max_down_percent, sample_count, warning, evaluated_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT(candidate_alert_id, horizon_minutes) DO UPDATE SET
		  target_at = EXCLUDED.target_at, observed_at = EXCLUDED.observed_at, observed_price = EXCLUDED.observed_price,
		  return_percent = EXCLUDED.return_percent, max_up_percent = EXCLUDED.max_up_percent,
		  max_down_percent = EXCLUDED.max_down_percent, sample_count = EXCLUDED.sample_count,
		  warning = EXCLUDED.warning, evaluated_at = EXCLUDED.evaluated_at
		RETURNING id
	`, outcome.CandidateAlertID, outcome.HorizonMinutes, outcome.TargetAt, outcome.ObservedAt, outcome.ObservedPrice,
		outcome.ReturnPercent, outcome.MaxUpPercent, outcome.MaxDownPercent, outcome.SampleCount,
		nullableString(outcome.Warning), outcome.EvaluatedAt).Scan(&outcome.ID)
	return outcome, err
}

func (s *PostgresStore) ListPolicyCandidateOutcomes(ctx context.Context, candidateID int64) ([]PolicyCandidateOutcome, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, candidate_alert_id, horizon_minutes, target_at, observed_at, observed_price, return_percent,
		  max_up_percent, max_down_percent, sample_count, COALESCE(warning, ''), evaluated_at
		FROM policy_candidate_outcomes WHERE candidate_alert_id = $1 ORDER BY horizon_minutes
	`, candidateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var outcomes []PolicyCandidateOutcome
	for rows.Next() {
		var outcome PolicyCandidateOutcome
		if err := rows.Scan(&outcome.ID, &outcome.CandidateAlertID, &outcome.HorizonMinutes, &outcome.TargetAt,
			&outcome.ObservedAt, &outcome.ObservedPrice, &outcome.ReturnPercent, &outcome.MaxUpPercent,
			&outcome.MaxDownPercent, &outcome.SampleCount, &outcome.Warning, &outcome.EvaluatedAt); err != nil {
			return nil, err
		}
		outcomes = append(outcomes, outcome)
	}
	return outcomes, rows.Err()
}

func (s *PostgresStore) UpsertPolicyCandidateFeedback(ctx context.Context, feedback PolicyCandidateFeedback) (PolicyCandidateFeedback, error) {
	if feedback.CandidateAlertID == 0 {
		return feedback, errors.New("policy feedback requires a candidate")
	}
	if feedback.EmotionIntensity != nil && (*feedback.EmotionIntensity < 0 || *feedback.EmotionIntensity > 5) {
		return feedback, errors.New("emotion intensity must be between 0 and 5")
	}
	if feedback.UpdatedAt.IsZero() {
		feedback.UpdatedAt = time.Now().UTC()
	}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO policy_candidate_feedback(candidate_alert_id, helpful, acted, emotion_intensity, notes, updated_at)
		VALUES($1,$2,$3,$4,$5,$6)
		ON CONFLICT(candidate_alert_id) DO UPDATE SET helpful = EXCLUDED.helpful, acted = EXCLUDED.acted,
		  emotion_intensity = EXCLUDED.emotion_intensity, notes = EXCLUDED.notes, updated_at = EXCLUDED.updated_at
		RETURNING candidate_alert_id, helpful, acted, emotion_intensity, COALESCE(notes, ''), updated_at
	`, feedback.CandidateAlertID, feedback.Helpful, feedback.Acted, feedback.EmotionIntensity,
		nullableString(feedback.Notes), feedback.UpdatedAt).Scan(&feedback.CandidateAlertID, &feedback.Helpful,
		&feedback.Acted, &feedback.EmotionIntensity, &feedback.Notes, &feedback.UpdatedAt)
	return feedback, err
}

func (s *PostgresStore) GetPolicyCandidateFeedback(ctx context.Context, candidateID int64) (*PolicyCandidateFeedback, error) {
	var feedback PolicyCandidateFeedback
	err := s.pool.QueryRow(ctx, `
		SELECT candidate_alert_id, helpful, acted, emotion_intensity, COALESCE(notes, ''), updated_at
		FROM policy_candidate_feedback WHERE candidate_alert_id = $1
	`, candidateID).Scan(&feedback.CandidateAlertID, &feedback.Helpful, &feedback.Acted,
		&feedback.EmotionIntensity, &feedback.Notes, &feedback.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &feedback, nil
}

func (s *PostgresStore) peerMapMembers(ctx context.Context, peerMapVersionID int64) ([]PeerMapMember, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT symbol, relation, weight FROM peer_map_members
		WHERE peer_map_version_id = $1 ORDER BY symbol
	`, peerMapVersionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var members []PeerMapMember
	for rows.Next() {
		var member PeerMapMember
		if err := rows.Scan(&member.Symbol, &member.Relation, &member.Weight); err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, rows.Err()
}

func (s *PostgresStore) PruneSnapshots(ctx context.Context, policy RetentionPolicy, now time.Time) (RetentionResult, error) {
	result := RetentionResult{}
	var err error
	if result.QuoteSnapshots, err = s.deleteOlderThan(ctx, "market_quote_snapshots", "received_at", policy.QuoteSnapshots, now); err != nil {
		return result, err
	}
	if result.RWASnapshots, err = s.deleteOlderThan(ctx, "rwa_quote_snapshots", "received_at", policy.RWASnapshots, now); err != nil {
		return result, err
	}
	if result.OptionSnapshots, err = s.deleteOlderThan(ctx, "option_contract_snapshots", "received_at", policy.OptionSnapshots, now); err != nil {
		return result, err
	}
	if result.IVSnapshots, err = s.deleteOlderThan(ctx, "iv_snapshots", "received_at", policy.OptionSnapshots, now); err != nil {
		return result, err
	}
	if result.FeatureSnapshots, err = s.deleteOlderThan(ctx, "intraday_feature_snapshots", "computed_at", policy.FeatureSnapshots, now); err != nil {
		return result, err
	}
	if result.CollectionRuns, err = s.deleteOlderThan(ctx, "collection_runs", "finished_at", policy.CollectionRuns, now); err != nil {
		return result, err
	}
	if result.EventEvidence, err = s.deleteOlderThan(ctx, "event_evidence", "received_at", policy.EventEvidence, now); err != nil {
		return result, err
	}
	if result.EventContexts, err = s.deleteOlderThan(ctx, "event_context_snapshots", "as_of", policy.EventContexts, now); err != nil {
		return result, err
	}
	if result.ShadowSignals, err = s.deleteOlderThan(ctx, "shadow_signal_observations", "as_of", policy.ShadowSignals, now); err != nil {
		return result, err
	}
	if result.EventRuns, err = s.deleteOlderThan(ctx, "event_ingestion_runs", "finished_at", policy.EventRuns, now); err != nil {
		return result, err
	}
	if result.PolicyOutcomes, err = s.deleteOlderThan(ctx, "policy_candidate_outcomes", "evaluated_at", policy.PolicyOutcomes, now); err != nil {
		return result, err
	}
	if result.PolicyCandidates, err = s.deleteOlderThan(ctx, "policy_candidate_alerts", "as_of", policy.PolicyCandidates, now); err != nil {
		return result, err
	}
	if policy.EventEvidence > 0 {
		if _, err = s.pool.Exec(ctx, `DELETE FROM market_events WHERE last_seen_at < $1`, now.Add(-policy.EventEvidence)); err != nil {
			return result, err
		}
	}
	return result, nil
}

func (s *PostgresStore) StorageDiagnostics(ctx context.Context) (StorageDiagnostic, error) {
	tables := []string{
		"market_quote_snapshots",
		"rwa_quote_snapshots",
		"rwa_ohlcv_bars",
		"option_contract_snapshots",
		"iv_snapshots",
		"intraday_feature_snapshots",
		"collection_runs",
		"event_ingestion_runs",
		"event_evidence",
		"market_events",
		"market_event_evidence",
		"event_context_snapshots",
		"shadow_signal_observations",
		"ohlcv_bars",
		"daily_market_summaries",
		"symbol_profiles",
		"peer_map_versions",
		"peer_map_members",
		"policy_candidate_alerts",
		"policy_candidate_outcomes",
		"policy_candidate_feedback",
	}
	diagnostic := StorageDiagnostic{
		Store:        "postgres",
		TableBytes:   map[string]int64{},
		RowEstimates: map[string]int64{},
		CollectedAt:  time.Now().UTC(),
	}
	for _, table := range tables {
		var bytes int64
		err := s.pool.QueryRow(ctx, `
			SELECT pg_total_relation_size(c.oid)
			FROM pg_class c
			WHERE c.oid = $1::regclass
		`, table).Scan(&bytes)
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil {
			return diagnostic, err
		}
		var rows int64
		if err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM "+table).Scan(&rows); err != nil {
			return diagnostic, err
		}
		diagnostic.TableBytes[table] = bytes
		diagnostic.RowEstimates[table] = rows
	}
	return diagnostic, nil
}

func (s *PostgresStore) deleteOlderThan(ctx context.Context, table, column string, retention time.Duration, now time.Time) (int64, error) {
	if retention <= 0 {
		return 0, nil
	}
	result, err := s.pool.Exec(ctx, "DELETE FROM "+table+" WHERE "+column+" < $1", now.Add(-retention))
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func nullableTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}

func nullableID(value int64) any {
	if value == 0 {
		return nil
	}
	return value
}

func nullableInt(value int) any {
	if value == 0 {
		return nil
	}
	return value
}

func sourceTierConfidence(tier string) float64 {
	switch tier {
	case "T1_OFFICIAL":
		return 1.0
	case "T2_PRIMARY":
		return 0.8
	case "T3_REPORTING":
		return 0.5
	default:
		return 0.1
	}
}

func selectRange(selectSQL, timestampColumn, symbol, provider string, bounds TimeRange) (string, []any) {
	conditions := make([]string, 0, 4)
	args := make([]any, 0, 5)
	add := func(condition string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(condition, len(args)))
	}
	if symbol != "" {
		add("symbol = $%d", symbol)
	}
	if provider != "" {
		add("provider = $%d", provider)
	}
	if !bounds.From.IsZero() {
		add(timestampColumn+" >= $%d", bounds.From)
	}
	if !bounds.To.IsZero() {
		add(timestampColumn+" <= $%d", bounds.To)
	}
	query := selectSQL
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	limit := bounds.Limit
	if limit <= 0 {
		limit = 1000
	}
	args = append(args, limit)
	if bounds.From.IsZero() {
		query += fmt.Sprintf(" ORDER BY %s DESC LIMIT $%d", timestampColumn, len(args))
		query = "SELECT * FROM (" + query + ") AS recent_snapshots ORDER BY " + timestampColumn + " ASC"
	} else {
		query += fmt.Sprintf(" ORDER BY %s ASC LIMIT $%d", timestampColumn, len(args))
	}
	return query, args
}

func nullTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}

func requiredJSON(value json.RawMessage, fallback []byte) []byte {
	if len(value) == 0 {
		return fallback
	}
	return value
}

func (s *PostgresStore) CreateAlert(ctx context.Context, cfg alert.AlertConfig) (alert.AlertConfig, error) {
	cfg.Symbol = normalizeSymbol(cfg.Symbol)
	if cfg.Symbol == "" {
		return cfg, errors.New("alert symbol is required")
	}
	if len(cfg.Rules) == 0 {
		return cfg, errors.New("alert requires at least one rule")
	}
	if cfg.ID == "" {
		cfg.ID = fmt.Sprintf("al_%d", time.Now().UTC().UnixNano())
	}
	if cfg.CreatedAt.IsZero() {
		cfg.CreatedAt = time.Now().UTC()
	}
	cfg.Enabled = true

	rules, err := json.Marshal(cfg.Rules)
	if err != nil {
		return cfg, err
	}
	notify, err := json.Marshal(cfg.Notify)
	if err != nil {
		return cfg, err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO alert_configs(id, symbol, group_name, rules_json, notify_json, analysis_on_trigger, enabled, created_at)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8)
	`, cfg.ID, cfg.Symbol, cfg.Group, rules, notify, cfg.AnalysisOnTrigger, cfg.Enabled, cfg.CreatedAt)
	return cfg, err
}

func (s *PostgresStore) ListAlerts(ctx context.Context) ([]alert.AlertConfig, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, symbol, group_name, rules_json, notify_json, analysis_on_trigger, enabled, created_at
		FROM alert_configs
		ORDER BY created_at
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []alert.AlertConfig
	for rows.Next() {
		var cfg alert.AlertConfig
		var rules []byte
		var notify []byte
		if err := rows.Scan(&cfg.ID, &cfg.Symbol, &cfg.Group, &rules, &notify, &cfg.AnalysisOnTrigger, &cfg.Enabled, &cfg.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(rules, &cfg.Rules); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(notify, &cfg.Notify); err != nil {
			return nil, err
		}
		configs = append(configs, cfg)
	}
	return configs, rows.Err()
}

func (s *PostgresStore) AddEvent(ctx context.Context, event alert.Event) error {
	if event.ID == "" {
		event.ID = fmt.Sprintf("ev_%d", time.Now().UTC().UnixNano())
	}
	if event.TriggeredAt.IsZero() {
		event.TriggeredAt = time.Now().UTC()
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO alert_events(id, alert_id, symbol, level, rule_type, message, observed_value,
		  threshold_value, benchmark, data_age_seconds, provider_time, received_at, triggered_at, analysis_on_trigger)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, event.ID, event.AlertID, event.Symbol, event.Level, event.RuleType, event.Message, event.ObservedValue,
		event.ThresholdValue, event.Benchmark, event.DataAgeSeconds, event.ProviderTime, event.ReceivedAt, event.TriggeredAt, event.AnalysisOnTrigger)
	return err
}

func (s *PostgresStore) ListEvents(ctx context.Context, limit int) ([]alert.Event, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, alert_id, symbol, level, rule_type, message, observed_value, threshold_value,
		  benchmark, data_age_seconds, provider_time, received_at, triggered_at, analysis_on_trigger
		FROM alert_events
		ORDER BY triggered_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []alert.Event
	for rows.Next() {
		var event alert.Event
		if err := rows.Scan(&event.ID, &event.AlertID, &event.Symbol, &event.Level, &event.RuleType, &event.Message,
			&event.ObservedValue, &event.ThresholdValue, &event.Benchmark, &event.DataAgeSeconds, &event.ProviderTime,
			&event.ReceivedAt, &event.TriggeredAt, &event.AnalysisOnTrigger); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (s *PostgresStore) EnqueueDiscordDelivery(ctx context.Context, delivery DiscordDelivery) (bool, error) {
	if delivery.IdempotencyKey == "" {
		return false, errors.New("discord delivery idempotency key is required")
	}
	payload, err := json.Marshal(delivery.Event)
	if err != nil {
		return false, err
	}
	channel := string(delivery.Event.Level)
	if channel == "" {
		channel = "ALERT"
	}
	command, err := s.pool.Exec(ctx, `INSERT INTO discord_delivery_outbox(idempotency_key, alert_event_id, level, webhook_channel, payload_json, status, next_attempt_at) VALUES($1,$2,$3,$4,$5,'pending',$6) ON CONFLICT (idempotency_key) DO NOTHING`, delivery.IdempotencyKey, delivery.Event.ID, delivery.Event.Level, channel, payload, delivery.NextAttemptAt)
	return command.RowsAffected() == 1, err
}

func (s *PostgresStore) ClaimDiscordDeliveries(ctx context.Context, now time.Time, limit int) ([]DiscordDelivery, error) {
	if limit <= 0 {
		limit = 20
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	rows, err := tx.Query(ctx, `WITH claimed AS (SELECT id FROM discord_delivery_outbox WHERE status='pending' AND next_attempt_at <= $1 ORDER BY created_at FOR UPDATE SKIP LOCKED LIMIT $2) UPDATE discord_delivery_outbox d SET status='processing', attempts=d.attempts+1, locked_at=$1, updated_at=$1 FROM claimed WHERE d.id=claimed.id RETURNING d.id,d.idempotency_key,d.payload_json,d.status,d.attempts,d.next_attempt_at,COALESCE(d.last_error,'')`, now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []DiscordDelivery{}
	for rows.Next() {
		var d DiscordDelivery
		var payload []byte
		if err := rows.Scan(&d.ID, &d.IdempotencyKey, &payload, &d.Status, &d.Attempts, &d.NextAttemptAt, &d.LastError); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(payload, &d.Event); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, tx.Commit(ctx)
}

func (s *PostgresStore) CompleteDiscordDelivery(ctx context.Context, id int64) error {
	_, err := s.pool.Exec(ctx, `UPDATE discord_delivery_outbox SET status='delivered', delivered_at=NOW(), updated_at=NOW() WHERE id=$1`, id)
	return err
}

func (s *PostgresStore) RetryDiscordDelivery(ctx context.Context, id int64, next time.Time, message string, dead bool) error {
	status := "pending"
	if dead {
		status = "dead_letter"
	}
	_, err := s.pool.Exec(ctx, `UPDATE discord_delivery_outbox SET status=$2,next_attempt_at=$3,last_error=$4,updated_at=NOW() WHERE id=$1`, id, status, next, message)
	return err
}
