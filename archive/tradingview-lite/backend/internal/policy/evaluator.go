package policy

import (
	"context"
	"encoding/json"
	"math"
	"sort"
	"time"

	"tradingview-lite/backend/internal/features"
	"tradingview-lite/backend/internal/session"
	"tradingview-lite/backend/internal/shadow"
	"tradingview-lite/backend/internal/store"
)

type Evaluator struct {
	Store  store.Store
	Config Config
}

func (e Evaluator) Evaluate(ctx context.Context, symbol, group string, asOf time.Time) (store.PolicyCandidateAlert, error) {
	if asOf.IsZero() {
		asOf = time.Now().UTC()
	}
	cfg := normalizeConfig(e.Config)
	profile, err := e.profile(ctx, symbol)
	if err != nil {
		return store.PolicyCandidateAlert{}, err
	}
	candidate := store.PolicyCandidateAlert{
		Symbol:           symbol,
		TradingDate:      session.TradingDate(asOf),
		AsOf:             asOf,
		ProfileRole:      profile.Role,
		Horizon:          profile.Horizon,
		SessionProduct:   SessionProduct(asOf),
		State:            StateBlocked,
		ActionCandidate:  ActionNoAction,
		DeliveryMode:     DeliveryModeShadowOnly,
		ThresholdVersion: cfg.ThresholdVersion,
	}

	contextSnapshot, err := e.Store.GetEventContext(ctx, symbol, asOf)
	if err != nil {
		return candidate, err
	}
	if contextSnapshot == nil || !contextSnapshot.MarketDataFresh || contextSnapshot.RiskState == "blocked" || asOf.Sub(contextSnapshot.AsOf) > cfg.DataStaleAfter {
		return e.blocked(ctx, candidate, "market_or_event_context_unavailable")
	}
	candidate.EventContextID = contextSnapshot.ID

	featureSnapshot, err := e.latestFeature(ctx, symbol, asOf, cfg.DataStaleAfter)
	if err != nil {
		return candidate, err
	}
	if featureSnapshot == nil {
		return e.blocked(ctx, candidate, "intraday_feature_missing_or_stale")
	}
	var feature features.IntradayResponse
	if err := json.Unmarshal(featureSnapshot.Payload, &feature); err != nil {
		return candidate, err
	}
	if feature.CurrentPrice <= 0 || feature.ReasonableRange.SelectedOneDayMovePercent <= 0 {
		return e.blocked(ctx, candidate, "range_inputs_unavailable")
	}
	candidate.ReferencePrice = feature.CurrentPrice

	peerMap, err := e.Store.GetActivePeerMap(ctx, symbol, group, asOf)
	if err != nil {
		return candidate, err
	}
	if peerMap != nil {
		candidate.PeerMapVersionID = peerMap.ID
	}
	shadowObservation, err := e.latestShadow(ctx, symbol, asOf)
	if err != nil {
		return candidate, err
	}
	rwaCalibration, err := CalibrateRWA(ctx, e.Store, symbol, asOf, cfg)
	if err != nil {
		return candidate, err
	}
	rangeCalibration, err := CalibrateRange(ctx, e.Store, symbol, asOf, cfg)
	if err != nil {
		return candidate, err
	}
	ivCalibration, err := CalibrateIV(ctx, e.Store, symbol, asOf, cfg)
	if err != nil {
		return candidate, err
	}

	previousCloseUse := rangeUse(feature.CurrentPrice, feature.PreviousClose, feature.ReasonableRange.SelectedOneDayMovePercent)
	openUse := rangeUse(feature.CurrentPrice, feature.RegularOpen, feature.ReasonableRange.SelectedOneDayMovePercent)
	metrics := map[string]any{
		"schemaVersion":           "behavior_policy_candidate_v1",
		"dataFresh":               true,
		"evidenceCoverage":        contextSnapshot.EvidenceCoverage,
		"eventRiskState":          contextSnapshot.RiskState,
		"selectedMovePercent":     feature.ReasonableRange.SelectedOneDayMovePercent,
		"selectedMoveSource":      feature.ReasonableRange.SelectedSource,
		"previousCloseRangeUse":   previousCloseUse,
		"openRangeUse":            openUse,
		"last30MinMovePercent":    feature.Realized.Last30MinMovePercent,
		"regularOpen":             feature.RegularOpen,
		"previousClose":           feature.PreviousClose,
		"rangeCalibration":        rangeCalibration,
		"ivCalibration":           ivCalibration,
		"rwaCalibration":          rwaCalibration,
		"peerMapConfigured":       peerMap != nil,
		"profileIsExplicit":       profile.UpdatedAt.IsZero() == false,
		"normalMarketDelivery":    false,
		"automaticTradingAllowed": false,
	}
	if peerMap != nil {
		metrics["peerMapKey"] = peerMap.MapKey
		metrics["peerMapReviewStatus"] = peerMap.ReviewStatus
		metrics["benchmarkSymbol"] = peerMap.BenchmarkSymbol
	}
	if shadowObservation != nil {
		metrics["impulseShadowState"] = shadowObservation.State
	}
	if feature.RWA != nil && feature.PreviousClose > 0 {
		metrics["rwaPremiumToPreviousClosePercent"] = 100 * (feature.RWA.LatestPrice/feature.PreviousClose - 1)
	}

	reasons := []string{"shadow_only_policy_evaluation"}
	warnings := append([]string(nil), feature.Warnings...)
	if peerMap == nil {
		warnings = append(warnings, "peer_map_unconfigured_or_unapproved")
	}
	if rwaCalibration.Status != "descriptive" {
		warnings = append(warnings, "rwa_proxy_not_calibrated_for_decision_use")
	}

	state, action, eligible, direction, additionalReasons := classify(candidate.SessionProduct, profile.Role, contextSnapshot.EvidenceCoverage, previousCloseUse, openUse, feature.Realized.Last30MinMovePercent, feature.ReasonableRange.SelectedOneDayMovePercent, directionFromPrices(feature.CurrentPrice, feature.PreviousClose), feature.RWA, rwaCalibration, shadowObservation)
	candidate.State = state
	if !actionPermitted(profile, action) {
		action = ActionWaitConfirmation
		eligible = false
		additionalReasons = append(additionalReasons, "action_not_permitted_by_symbol_profile")
		candidate.SuppressedReason = "action_not_permitted_by_symbol_profile"
	}
	candidate.ActionCandidate = action
	candidate.Eligible = eligible
	candidate.Direction = direction
	reasons = append(reasons, additionalReasons...)
	if state == StateBlocked {
		candidate.Eligible = false
	}
	candidate.ReasonCodes, _ = json.Marshal(reasons)
	candidate.Metrics, _ = json.Marshal(metrics)
	candidate.Warnings, _ = json.Marshal(warnings)
	return e.applyAttentionBudget(ctx, candidate, cfg)
}

func (e Evaluator) blocked(ctx context.Context, candidate store.PolicyCandidateAlert, reason string) (store.PolicyCandidateAlert, error) {
	candidate.State = StateBlocked
	candidate.ActionCandidate = ActionNoAction
	candidate.Eligible = false
	candidate.WouldNotify = false
	candidate.SuppressedReason = reason
	candidate.ReasonCodes, _ = json.Marshal([]string{reason})
	candidate.Metrics, _ = json.Marshal(map[string]any{
		"schemaVersion":           "behavior_policy_candidate_v1",
		"normalMarketDelivery":    false,
		"automaticTradingAllowed": false,
	})
	candidate.Warnings, _ = json.Marshal([]string{"candidate blocked because required decision inputs are unavailable or stale"})
	return candidate, nil
}

func (e Evaluator) profile(ctx context.Context, symbol string) (store.SymbolProfile, error) {
	profile, err := e.Store.GetSymbolProfile(ctx, symbol)
	if err != nil {
		return store.SymbolProfile{}, err
	}
	if profile == nil {
		return store.SymbolProfile{
			Symbol:            symbol,
			Role:              "research_watch",
			Horizon:           "unspecified",
			ActionPermissions: json.RawMessage(`["WAIT_CONFIRMATION","NO_CHASE"]`),
		}, nil
	}
	return *profile, nil
}

func (e Evaluator) latestFeature(ctx context.Context, symbol string, asOf time.Time, staleAfter time.Duration) (*store.FeatureSnapshot, error) {
	rows, err := e.Store.ListFeatureSnapshots(ctx, symbol, store.TimeRange{From: asOf.Add(-5 * staleAfter), To: asOf, Limit: 100})
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	latest := rows[len(rows)-1]
	if asOf.Sub(latest.ComputedAt) > staleAfter {
		return nil, nil
	}
	return &latest, nil
}

func (e Evaluator) latestShadow(ctx context.Context, symbol string, asOf time.Time) (*store.ShadowSignalObservation, error) {
	rows, err := e.Store.ListShadowSignals(ctx, symbol, store.TimeRange{From: asOf.Add(-10 * time.Minute), To: asOf, Limit: 100})
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	latest := rows[len(rows)-1]
	return &latest, nil
}

func (e Evaluator) applyAttentionBudget(ctx context.Context, candidate store.PolicyCandidateAlert, cfg Config) (store.PolicyCandidateAlert, error) {
	if !candidate.Eligible {
		if candidate.SuppressedReason == "" {
			candidate.SuppressedReason = "not_candidate_state"
		}
		return candidate, nil
	}
	rows, err := e.Store.ListPolicyCandidates(ctx, store.PolicyCandidateFilter{
		Symbol:      candidate.Symbol,
		TradingDate: candidate.TradingDate,
		Bounds:      store.TimeRange{From: candidate.TradingDate, To: candidate.AsOf, Limit: 1000},
	})
	if err != nil {
		return candidate, err
	}
	for index := len(rows) - 1; index >= 0; index-- {
		prior := rows[index]
		if prior.SessionProduct == candidate.SessionProduct && prior.State == candidate.State && candidate.AsOf.Sub(prior.AsOf) < cfg.CandidateCooldown {
			candidate.SuppressedReason = "symbol_state_cooldown"
			return candidate, nil
		}
	}
	allRows, err := e.Store.ListPolicyCandidates(ctx, store.PolicyCandidateFilter{
		TradingDate:     candidate.TradingDate,
		OnlyWouldNotify: true,
		Bounds:          store.TimeRange{From: candidate.TradingDate, To: candidate.AsOf, Limit: 1000},
	})
	if err != nil {
		return candidate, err
	}
	if len(allRows) >= cfg.DailyCandidateBudget {
		candidate.SuppressedReason = "daily_candidate_budget_exhausted"
		return candidate, nil
	}
	candidate.WouldNotify = true
	candidate.BudgetSlot = len(allRows) + 1
	return candidate, nil
}

func SessionProduct(at time.Time) string {
	if session.ClassifyUS(at) == session.Premarket {
		return "premarket"
	}
	if session.ClassifyUS(at) != session.Regular {
		return "overnight"
	}
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		loc = time.UTC
	}
	local := at.In(loc)
	minutes := local.Hour()*60 + local.Minute()
	if minutes < 10*60 {
		return "opening"
	}
	if minutes >= 12*60 && minutes < 14*60+30 {
		return "midday"
	}
	return "regular"
}

func classify(product, role, evidenceCoverage string, previousCloseUse, openUse, last30Move, selectedMove float64, direction string, rwaFeature *features.RWAFeature, rwaCalibration RWAProxyCalibration, impulse *store.ShadowSignalObservation) (string, string, bool, string, []string) {
	currentUse := previousCloseUse
	if product == "opening" && openUse > 0 {
		currentUse = openUse
	}
	if impulse != nil {
		if metricsDirection(impulse.Metrics) != "" {
			direction = metricsDirection(impulse.Metrics)
		}
	}
	if direction == "" {
		direction = "unknown"
	}
	unsupportedEvidence := evidenceCoverage != "supported"

	switch product {
	case "premarket":
		if previousCloseUse >= 1 && unsupportedEvidence {
			return StateGapUnsupported, ActionWaitConfirmation, true, direction, []string{"premarket_gap_exceeds_selected_range_without_supported_event"}
		}
		if previousCloseUse >= 1 {
			return StateGapExtended, ActionWaitConfirmation, false, direction, []string{"premarket_gap_exceeds_selected_range_with_supported_event_context"}
		}
		return StateGapWithinRange, ActionWaitConfirmation, false, direction, []string{"premarket_gap_within_selected_range"}
	case "opening":
		if impulse != nil && impulse.State == shadow.StateRejected {
			return StateUnsupportedImpulse, actionForRole(role, StateUnsupportedImpulse, currentUse), true, direction, []string{"opening_impulse_rejected_by_shadow_model", "opening_window_defaults_to_no_chase"}
		}
		if currentUse >= 1 && unsupportedEvidence {
			return StateUnsupportedImpulse, actionForRole(role, StateUnsupportedImpulse, currentUse), true, direction, []string{"opening_move_extended_without_supported_event", "opening_window_defaults_to_no_chase"}
		}
		return StatePriceDiscovery, ActionWaitConfirmation, false, direction, []string{"opening_price_discovery_window"}
	case "midday":
		if currentUse >= 0.8 && math.Abs(last30Move) >= math.Max(0.5, selectedMove*0.25) {
			return StateReaccelerationRisk, actionForRole(role, StateReaccelerationRisk, currentUse), true, direction, []string{"midday_range_extension_with_recent_reacceleration"}
		}
		return StateStabilizing, ActionHoldObserve, false, direction, []string{"midday_reacceleration_not_detected"}
	case "overnight":
		if rwaFeature != nil && math.Abs(rwaCalibration.LatestPremiumPercent) >= selectedMove {
			if rwaCalibration.Status == "descriptive" {
				return StateRWADivergence, ActionWaitConfirmation, false, direction, []string{"rwa_proxy_divergence_recorded_not_trade_signal"}
			}
			return StateRWAUncalibratedDivergence, ActionWaitConfirmation, false, direction, []string{"rwa_proxy_divergence_without_sufficient_calibration"}
		}
		return StateWaitConfirmation, ActionWaitConfirmation, false, direction, []string{"overnight_context_requires_verified_evidence_or_cash_market_confirmation"}
	default:
		if impulse != nil && impulse.State == shadow.StateRejected {
			return StateUnsupportedImpulse, actionForRole(role, StateUnsupportedImpulse, currentUse), true, direction, []string{"impulse_rejection_shadow_state"}
		}
		if currentUse >= 1.5 && unsupportedEvidence {
			return StateDislocationRisk, actionForRole(role, StateDislocationRisk, currentUse), true, direction, []string{"range_extension_exceeds_local_policy_without_supported_event"}
		}
		return StateWaitConfirmation, ActionWaitConfirmation, false, direction, []string{"no_high_confidence_behavioral_candidate"}
	}
}

func actionForRole(role, state string, rangeUse float64) string {
	if role != "position" {
		return ActionNoChase
	}
	if state == StateDislocationRisk && rangeUse >= 2 {
		return ActionStopReview
	}
	if state == StateDislocationRisk || state == StateReaccelerationRisk {
		return ActionReductionReview
	}
	return ActionHoldObserve
}

func actionPermitted(profile store.SymbolProfile, action string) bool {
	if action == ActionNoAction || action == ActionWaitConfirmation {
		return true
	}
	if len(profile.ActionPermissions) == 0 {
		return true
	}
	var permissions []string
	if json.Unmarshal(profile.ActionPermissions, &permissions) != nil || len(permissions) == 0 {
		return false
	}
	for _, permitted := range permissions {
		if permitted == action {
			return true
		}
	}
	return false
}

func rangeUse(price, anchor, selectedMovePct float64) float64 {
	if price <= 0 || anchor <= 0 || selectedMovePct <= 0 {
		return 0
	}
	return math.Abs(100*(price/anchor-1)) / selectedMovePct
}

func directionFromPrices(price, anchor float64) string {
	if price > anchor && anchor > 0 {
		return "up"
	}
	if price < anchor && anchor > 0 {
		return "down"
	}
	return ""
}

func metricsDirection(raw json.RawMessage) string {
	var metrics map[string]any
	if json.Unmarshal(raw, &metrics) != nil {
		return ""
	}
	move, ok := metrics["impulseMovePct"].(float64)
	if !ok {
		return ""
	}
	if move > 0 {
		return "up"
	}
	if move < 0 {
		return "down"
	}
	return ""
}

func CalibrateRWA(ctx context.Context, repo store.Store, symbol string, asOf time.Time, config Config) (RWAProxyCalibration, error) {
	cfg := normalizeConfig(config)
	result := RWAProxyCalibration{Symbol: symbol, AsOf: asOf, MaxAlignmentAgeSeconds: int64(cfg.RWAAlignmentWindow.Seconds()), Status: "uncalibrated"}
	quotes, err := repo.ListQuoteSnapshots(ctx, symbol, "yahoo_chart", store.TimeRange{From: asOf.Add(-14 * 24 * time.Hour), To: asOf, Limit: 10000})
	if err != nil {
		return result, err
	}
	rwas, err := repo.ListRWASnapshots(ctx, symbol, "bitget_rwa", store.TimeRange{From: asOf.Add(-14 * 24 * time.Hour), To: asOf, Limit: 10000})
	if err != nil {
		return result, err
	}
	type pair struct {
		at      time.Time
		equity  float64
		rwa     float64
		premium float64
	}
	pairs := make([]pair, 0)
	rwaIndex := 0
	for _, quote := range quotes {
		for rwaIndex+1 < len(rwas) && !rwas[rwaIndex+1].ReceivedAt.After(quote.ReceivedAt) {
			rwaIndex++
		}
		if len(rwas) == 0 || rwas[rwaIndex].ReceivedAt.After(quote.ReceivedAt) || quote.ReceivedAt.Sub(rwas[rwaIndex].ReceivedAt) > cfg.RWAAlignmentWindow || quote.Price <= 0 || rwas[rwaIndex].Price <= 0 {
			continue
		}
		pairs = append(pairs, pair{at: quote.ReceivedAt, equity: quote.Price, rwa: rwas[rwaIndex].Price, premium: 100 * (rwas[rwaIndex].Price/quote.Price - 1)})
	}
	result.MatchedSamples = len(pairs)
	if len(pairs) == 0 {
		result.Warning = "no point-in-time RWA/equity pairs within the configured alignment window"
		return result, nil
	}
	for _, point := range pairs {
		result.MeanPremiumPercent += point.premium
	}
	result.MeanPremiumPercent /= float64(len(pairs))
	result.LatestPremiumPercent = pairs[len(pairs)-1].premium
	equityReturns := make([]float64, 0, len(pairs)-1)
	rwaReturns := make([]float64, 0, len(pairs)-1)
	for index := 1; index < len(pairs); index++ {
		if pairs[index-1].equity <= 0 || pairs[index-1].rwa <= 0 {
			continue
		}
		equityReturns = append(equityReturns, 100*(pairs[index].equity/pairs[index-1].equity-1))
		rwaReturns = append(rwaReturns, 100*(pairs[index].rwa/pairs[index-1].rwa-1))
	}
	result.ReturnPairs = len(equityReturns)
	if len(equityReturns) >= cfg.RWAMinimumSamples {
		result.ContemporaneousCorr = pearson(equityReturns, rwaReturns)
		result.Status = "descriptive"
		result.Warning = "descriptive contemporaneous alignment only; this does not establish price discovery or lead-lag"
	} else {
		result.Warning = "insufficient aligned return pairs; RWA remains explanatory context only"
	}
	return result, nil
}

func CalibrateRange(ctx context.Context, repo store.Store, symbol string, asOf time.Time, config Config) (RangeCalibration, error) {
	cfg := normalizeConfig(config)
	result := RangeCalibration{Symbol: symbol, AsOf: asOf, Status: "uncalibrated"}
	rows, err := repo.ListDailySummaries(ctx, symbol, store.TimeRange{From: asOf.AddDate(0, 0, -cfg.RangeLookbackDays), To: asOf, Limit: cfg.RangeLookbackDays + 10})
	if err != nil {
		return result, err
	}
	currentTradingDate := session.TradingDate(asOf)
	sources := map[string]bool{}
	for _, row := range rows {
		if !row.TradingDate.Before(currentTradingDate) {
			continue
		}
		if row.PreviousClose <= 0 || row.SelectedMovePct <= 0 || row.DailyHigh <= 0 || row.DailyLow <= 0 {
			continue
		}
		actual := math.Max(math.Abs(100*(row.DailyHigh/row.PreviousClose-1)), math.Abs(100*(row.DailyLow/row.PreviousClose-1)))
		use := actual / row.SelectedMovePct
		result.MeanRangeUse += use
		if use <= 1 {
			result.CoveragePercent += 100
		}
		result.SampleDays++
		if row.SelectedSource != "" {
			sources[row.SelectedSource] = true
		}
	}
	if result.SampleDays == 0 {
		result.Warning = "no completed daily summaries with a usable selected range"
		return result, nil
	}
	result.MeanRangeUse /= float64(result.SampleDays)
	result.CoveragePercent /= float64(result.SampleDays)
	for source := range sources {
		result.SelectedSources = append(result.SelectedSources, source)
	}
	sort.Strings(result.SelectedSources)
	if result.SampleDays >= cfg.RangeMinimumDays {
		result.Status = "descriptive"
		result.Warning = "descriptive historical calibration; expected-move coverage is not a directional probability"
	} else {
		result.Warning = "insufficient completed daily summaries; selected range remains provisional"
	}
	return result, nil
}

func CalibrateIV(ctx context.Context, repo store.Store, symbol string, asOf time.Time, config Config) (IVCalibration, error) {
	cfg := normalizeConfig(config)
	result := IVCalibration{Symbol: symbol, AsOf: asOf, Status: "uncalibrated"}
	summaries, err := repo.ListDailySummaries(ctx, symbol, store.TimeRange{
		From:  asOf.AddDate(0, 0, -cfg.RangeLookbackDays),
		To:    asOf,
		Limit: cfg.RangeLookbackDays + 10,
	})
	if err != nil {
		return result, err
	}
	ivRows, err := repo.ListIVSnapshots(ctx, symbol, store.TimeRange{
		From:  asOf.AddDate(0, 0, -cfg.RangeLookbackDays),
		To:    asOf,
		Limit: 10000,
	})
	if err != nil {
		return result, err
	}
	result.SnapshotCount = len(ivRows)
	currentTradingDate := session.TradingDate(asOf)
	for _, summary := range summaries {
		if !summary.TradingDate.Before(currentTradingDate) || summary.PreviousClose <= 0 || summary.DailyHigh <= 0 || summary.DailyLow <= 0 {
			continue
		}
		cutoff := regularOpenAt(summary.TradingDate)
		var selected *store.IVSnapshot
		for index := range ivRows {
			row := &ivRows[index]
			if row.ReceivedAt.After(cutoff) {
				break
			}
			if row.OneDayMovePercent > 0 {
				selected = row
			}
		}
		if selected == nil {
			continue
		}
		actual := math.Max(
			math.Abs(100*(summary.DailyHigh/summary.PreviousClose-1)),
			math.Abs(100*(summary.DailyLow/summary.PreviousClose-1)),
		)
		result.MatchedDays++
		result.MeanImpliedMovePercent += selected.OneDayMovePercent
		result.MeanActualMovePercent += actual
		result.MeanBiasPercent += selected.OneDayMovePercent - actual
		if actual <= selected.OneDayMovePercent {
			result.CoveragePercent += 100
		}
	}
	if result.MatchedDays == 0 {
		result.Warning = "no completed daily ranges could be matched to a pre-open IV snapshot"
		return result, nil
	}
	result.MeanImpliedMovePercent /= float64(result.MatchedDays)
	result.MeanActualMovePercent /= float64(result.MatchedDays)
	result.MeanBiasPercent /= float64(result.MatchedDays)
	result.CoveragePercent /= float64(result.MatchedDays)
	if result.MatchedDays >= cfg.RangeMinimumDays {
		result.Status = "descriptive"
		result.Warning = "descriptive IV-to-realized-range comparison; it does not provide directional probability"
	} else {
		result.Warning = "insufficient matched completed days; IV calibration remains provisional"
	}
	return result, nil
}

func regularOpenAt(tradingDate time.Time) time.Time {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		loc = time.UTC
	}
	return time.Date(tradingDate.Year(), tradingDate.Month(), tradingDate.Day(), 9, 30, 0, 0, loc).UTC()
}

func pearson(left, right []float64) float64 {
	if len(left) != len(right) || len(left) < 2 {
		return 0
	}
	leftMean := 0.0
	rightMean := 0.0
	for index := range left {
		leftMean += left[index]
		rightMean += right[index]
	}
	leftMean /= float64(len(left))
	rightMean /= float64(len(right))
	numerator := 0.0
	leftVariance := 0.0
	rightVariance := 0.0
	for index := range left {
		leftDelta := left[index] - leftMean
		rightDelta := right[index] - rightMean
		numerator += leftDelta * rightDelta
		leftVariance += leftDelta * leftDelta
		rightVariance += rightDelta * rightDelta
	}
	if leftVariance == 0 || rightVariance == 0 {
		return 0
	}
	return numerator / math.Sqrt(leftVariance*rightVariance)
}
