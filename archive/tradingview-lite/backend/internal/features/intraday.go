package features

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"tradingview-lite/backend/internal/market"
	"tradingview-lite/backend/internal/options"
	"tradingview-lite/backend/internal/rwa"
)

type MarketClient interface {
	FetchQuote(ctx context.Context, symbol string) (*market.Quote, error)
	FetchBars(ctx context.Context, symbol, rangeValue, interval string, includePrePost bool) (*market.BarsResponse, error)
}

type OptionsClient interface {
	FetchChain(ctx context.Context, symbol string, nearStrikes int) (*options.ChainResponse, error)
}

type RWAClient interface {
	StockInfo(ctx context.Context, symbol string) (*rwa.StockInfo, error)
	Kline(ctx context.Context, symbol, period string, size int) (*rwa.KlineResponse, error)
}

type IntradayInput struct {
	Quote       *market.Quote
	Bars        *market.BarsResponse
	OptionChain *options.ChainResponse
	RWAInfo     *rwa.StockInfo
	RWABars     *rwa.KlineResponse
	Warnings    []string
	ComputedAt  time.Time
}

func ComputeIntraday(ctx context.Context, symbol string, marketClient MarketClient, optionsClient OptionsClient, rwaClient RWAClient) (*IntradayResponse, error) {
	quote, err := marketClient.FetchQuote(ctx, symbol)
	if err != nil {
		return nil, err
	}
	bars, err := marketClient.FetchBars(ctx, symbol, "1d", "1m", true)
	if err != nil {
		return nil, err
	}

	input := IntradayInput{
		Quote:      quote,
		Bars:       bars,
		ComputedAt: time.Now().UTC(),
	}
	if optionsClient != nil {
		if fetched, err := optionsClient.FetchChain(ctx, symbol, 10); err == nil {
			input.OptionChain = fetched
		} else {
			input.Warnings = append(input.Warnings, "options unavailable: "+err.Error())
		}
	}

	if rwaClient != nil {
		if info, err := rwaClient.StockInfo(ctx, symbol); err == nil {
			input.RWAInfo = info
			if kline, err := rwaClient.Kline(ctx, symbol, "1m", 20); err == nil {
				input.RWABars = kline
			} else {
				input.Warnings = append(input.Warnings, "RWA kline unavailable: "+err.Error())
			}
		} else {
			input.Warnings = append(input.Warnings, "RWA unavailable: "+err.Error())
		}
	}
	return ComputeFromData(input)
}

func ComputeFromData(input IntradayInput) (*IntradayResponse, error) {
	if input.Quote == nil {
		return nil, fmt.Errorf("intraday input requires quote")
	}
	if input.Bars == nil {
		return nil, fmt.Errorf("intraday input requires bars")
	}
	now := input.ComputedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}
	quote := input.Quote
	bars := input.Bars
	out := &IntradayResponse{
		Symbol:        quote.Symbol,
		ReceivedAt:    now,
		CurrentPrice:  quote.Price,
		RegularOpen:   regularOpen(bars.Bars),
		ProviderTimes: map[string]string{"yahoo_quote": quote.RegularMarketTime.Format(time.RFC3339), "yahoo_bars": bars.RegularMarketTime.Format(time.RFC3339)},
		Realized:      computeRealized(bars.Bars),
		PriceSpeed:    computeSpeeds(bars.Bars, []int{1, 3, 5, 15}),
		Warnings:      append([]string(nil), input.Warnings...),
	}
	if quote.ProviderWarning != "" {
		out.Warnings = append(out.Warnings, quote.ProviderWarning)
	}
	if chain := input.OptionChain; chain != nil {
		out.PreviousClose = chain.Underlying.PreviousClose
		out.ProviderTimes["yahoo_options"] = chain.ReceivedAt.Format(time.RFC3339)
		out.Options = &OptionsFeature{
			ExpirationDate:                   chain.ExpirationDate,
			ATMStrike:                        chain.Summary.ATMStrike,
			StraddleExpectedMove:             chain.Summary.ExpectedMove,
			StraddleExpectedMovePercent:      chain.Summary.ExpectedMovePercent,
			AverageResolvedImpliedVolatility: chain.Summary.AverageResolvedImpliedVolatility,
			AverageRawImpliedVolatility:      chain.Summary.AverageRawImpliedVolatility,
			OneTradingDayExpectedMove:        chain.Summary.OneTradingDayExpectedMove,
			OneTradingDayExpectedMovePercent: chain.Summary.OneTradingDayExpectedMovePercent,
			TimeToExpirationYears:            chain.Summary.TimeToExpirationYears,
			UnderlyingDataAgeSeconds:         chain.Summary.UnderlyingDataAgeSeconds,
			ContractTradeDataAgeSeconds:      chain.Summary.ContractTradeDataAgeSeconds,
			MethodWarning:                    chain.Summary.MethodWarning,
		}
		if chain.ProviderWarning != "" {
			out.Warnings = append(out.Warnings, chain.ProviderWarning)
		}
		if chain.Summary.UnderlyingDataAgeSeconds > 120 {
			out.Warnings = append(out.Warnings, "options underlying quote is stale; option-derived range is excluded")
		}
	}
	if out.PreviousClose == 0 && len(bars.Bars) > 0 {
		out.PreviousClose = bars.Bars[0].Open
		out.Warnings = append(out.Warnings, "previous close unavailable; using first returned bar open as fallback")
	}
	if info := input.RWAInfo; info != nil {
		feature := &RWAFeature{
			Symbol:       info.Symbol,
			Ticker:       info.Ticker,
			LatestPrice:  info.LatestPrice,
			MarketStatus: info.MarketStatus,
			DataSource:   info.DataSource,
			Warning:      info.ProviderWarning,
		}
		if input.RWABars != nil {
			feature.Speeds = computeRWASpeeds(input.RWABars.Bars, []int{1, 3, 5, 15})
			out.ProviderTimes["bitget_rwa"] = input.RWABars.ReceivedAt.Format(time.RFC3339)
		}
		out.RWA = feature
	}

	rangeRealized := out.Realized
	rangeOptions := out.Options
	if quote.DataAgeSeconds > 120 || bars.DataAgeSeconds > 120 {
		out.Warnings = append(out.Warnings, "Yahoo quote or bar input is stale; current reasonable range is unavailable")
		rangeRealized.ProjectedDailyMovePercent = 0
		rangeOptions = nil
	}
	out.ReasonableRange = computeRange(out.CurrentPrice, out.PreviousClose, out.RegularOpen, rangeRealized, rangeOptions, now)
	out.Warnings = append(out.Warnings, out.ReasonableRange.Warnings...)
	return out, nil
}

func computeRealized(bars []market.Bar) RealizedVol {
	allReturns := logReturns(bars)
	regular := regularSessionBars(bars)
	regularReturns := logReturns(regular)
	return RealizedVol{
		BarsUsed:                   len(bars),
		RegularBarsUsed:            len(regular),
		RealizedMovePercent:        realizedMovePercent(allReturns),
		RegularRealizedMovePercent: realizedMovePercent(regularReturns),
		Last30MinMovePercent:       realizedMovePercent(tailReturns(allReturns, 30)),
		ProjectedDailyMovePercent:  projectedDailyMovePercent(regularReturns),
		Method:                     "1-minute log returns; realized move is sqrt(sum r^2), projected daily move is sample std times sqrt(390)",
	}
}

func computeSpeeds(bars []market.Bar, horizons []int) []SpeedPoint {
	if len(bars) < 2 {
		return nil
	}
	out := make([]SpeedPoint, 0, len(horizons))
	last := bars[len(bars)-1]
	for _, horizon := range horizons {
		if len(bars) <= horizon {
			continue
		}
		from := bars[len(bars)-1-horizon]
		out = append(out, speedPoint(labelMinutes(horizon), from.Time, last.Time, from.Close, last.Close))
	}
	return out
}

func computeRWASpeeds(bars []rwa.Bar, horizons []int) []SpeedPoint {
	if len(bars) < 2 {
		return nil
	}
	out := make([]SpeedPoint, 0, len(horizons))
	last := bars[len(bars)-1]
	for _, horizon := range horizons {
		if len(bars) <= horizon {
			continue
		}
		from := bars[len(bars)-1-horizon]
		out = append(out, speedPoint(labelMinutes(horizon), from.Time, last.Time, from.Close, last.Close))
	}
	return out
}

func speedPoint(label string, fromTime, toTime time.Time, fromPrice, toPrice float64) SpeedPoint {
	elapsed := int64(toTime.Sub(fromTime).Seconds())
	ret := 0.0
	velocity := 0.0
	if fromPrice > 0 {
		ret = (toPrice/fromPrice - 1) * 100
	}
	if elapsed > 0 {
		velocity = ret / float64(elapsed)
	}
	return SpeedPoint{
		Horizon:              label,
		ElapsedSeconds:       elapsed,
		FromTime:             fromTime,
		ToTime:               toTime,
		FromPrice:            fromPrice,
		ToPrice:              toPrice,
		ReturnPercent:        ret,
		VelocityPctPerSecond: velocity,
	}
}

func computeRange(current, previousClose, open float64, realized RealizedVol, optionFeature *OptionsFeature, now time.Time) ReasonableRange {
	warnings := []string{}
	optionMove := 0.0
	optionMovePct := 0.0
	if optionFeature != nil && optionFeature.UnderlyingDataAgeSeconds <= 120 {
		optionMove = optionFeature.OneTradingDayExpectedMove
		optionMovePct = optionFeature.OneTradingDayExpectedMovePercent
	} else if optionFeature != nil {
		warnings = append(warnings, "option-derived move excluded because underlying option quote is stale")
	}
	realizedMove := 0.0
	realizedMovePct := realized.ProjectedDailyMovePercent
	if current > 0 && realizedMovePct > 0 {
		realizedMove = current * realizedMovePct / 100
	}
	selected := optionMove
	selectedPct := optionMovePct
	source := "unavailable"
	if selected > 0 {
		source = "options_resolved_iv"
	}
	if realizedMove > selected {
		selected = realizedMove
		selectedPct = realizedMovePct
		source = "realized_volatility_projection"
	}
	if selected == 0 {
		warnings = append(warnings, "no option-derived or realized-volatility-derived move available")
	}

	out := ReasonableRange{
		SelectedOneDayMove:        selected,
		SelectedOneDayMovePercent: selectedPct,
		SelectedSource:            source,
		Method:                    "selected one-day move is max(option one-day move from resolved IV, projected daily realized move from 1m regular-session returns)",
		Warnings:                  warnings,
	}
	if previousClose > 0 && selected > 0 {
		out.PreviousCloseRange = rangeBand("previous_close", previousClose, selected)
	}
	if open > 0 && selected > 0 {
		out.OpenRange = rangeBand("regular_open", open, selected)
	}
	remainingMinutes := regularMinutesRemaining(now)
	out.RegularMinutesRemaining = remainingMinutes
	if current > 0 && selected > 0 && remainingMinutes > 0 {
		remainingMove := selected * math.Sqrt(float64(remainingMinutes)/390.0)
		out.CurrentRemainingRange = rangeBand("current_price_remaining_regular_session", current, remainingMove)
	} else if remainingMinutes == 0 {
		out.Warnings = append(out.Warnings, "current time is outside regular session; current remaining range is not computed")
	}
	return out
}

func rangeBand(anchor string, price, move float64) *RangeBand {
	return &RangeBand{
		Anchor:      anchor,
		AnchorPrice: price,
		Move1Sigma:  move,
		Low1Sigma:   price - move,
		High1Sigma:  price + move,
		Move2Sigma:  move * 2,
		Low2Sigma:   price - 2*move,
		High2Sigma:  price + 2*move,
	}
}

func logReturns(bars []market.Bar) []float64 {
	if len(bars) < 2 {
		return nil
	}
	out := make([]float64, 0, len(bars)-1)
	for i := 1; i < len(bars); i++ {
		prev := bars[i-1].Close
		next := bars[i].Close
		if prev <= 0 || next <= 0 {
			continue
		}
		out = append(out, math.Log(next/prev))
	}
	return out
}

func realizedMovePercent(returns []float64) float64 {
	sum := 0.0
	for _, value := range returns {
		sum += value * value
	}
	return math.Sqrt(sum) * 100
}

func projectedDailyMovePercent(returns []float64) float64 {
	if len(returns) < 2 {
		return 0
	}
	mean := 0.0
	for _, value := range returns {
		mean += value
	}
	mean /= float64(len(returns))
	variance := 0.0
	for _, value := range returns {
		diff := value - mean
		variance += diff * diff
	}
	std := math.Sqrt(variance / float64(len(returns)-1))
	return std * math.Sqrt(390) * 100
}

func tailReturns(values []float64, n int) []float64 {
	if len(values) <= n {
		return values
	}
	return values[len(values)-n:]
}

func regularOpen(bars []market.Bar) float64 {
	for _, bar := range bars {
		ny := bar.Time.In(nyLocation())
		if ny.Hour() == 9 && ny.Minute() == 30 {
			return bar.Open
		}
	}
	return 0
}

func regularSessionBars(bars []market.Bar) []market.Bar {
	out := make([]market.Bar, 0, len(bars))
	for _, bar := range bars {
		if isRegularSession(bar.Time) {
			out = append(out, bar)
		}
	}
	return out
}

func isRegularSession(t time.Time) bool {
	ny := t.In(nyLocation())
	minutes := ny.Hour()*60 + ny.Minute()
	return minutes >= 9*60+30 && minutes <= 16*60
}

func regularMinutesRemaining(now time.Time) int {
	ny := now.In(nyLocation())
	minutes := ny.Hour()*60 + ny.Minute()
	open := 9*60 + 30
	close := 16 * 60
	if minutes < open || minutes >= close {
		return 0
	}
	return close - minutes
}

func nyLocation() *time.Location {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return time.UTC
	}
	return loc
}

func labelMinutes(minutes int) string {
	return strconv.Itoa(minutes) + "m"
}

func appendWarning(existing, next string) string {
	if existing == "" {
		return next
	}
	if next == "" {
		return existing
	}
	return existing + "; " + next
}
