package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"testing"
	"time"

	"tradingview-lite/backend/internal/features"
	"tradingview-lite/backend/internal/options"
	"tradingview-lite/backend/internal/shadow"
	"tradingview-lite/backend/internal/store"
)

func TestOpeningRejectionCreatesNoChaseCandidateInShadowOnlyMode(t *testing.T) {
	ctx := context.Background()
	repo := store.NewMemoryStore(nil)
	asOf := newYorkTime(9, 45)
	seedPolicyInputs(t, ctx, repo, "NVDA", asOf, 105, 100, 104, 5)
	if _, err := repo.AppendShadowSignal(ctx, store.ShadowSignalObservation{
		SignalType: shadow.SignalImpulseRejection, Symbol: "NVDA", AsOf: asOf, State: shadow.StateRejected,
		Eligible: true, ThresholdVersion: "impulse_shadow_v1", Metrics: json.RawMessage(`{"impulseMovePct":1.1}`),
	}); err != nil {
		t.Fatal(err)
	}

	candidate, err := (Evaluator{Store: repo}).Evaluate(ctx, "NVDA", "semis_memory", asOf)
	if err != nil {
		t.Fatal(err)
	}
	if candidate.State != StateUnsupportedImpulse || candidate.ActionCandidate != ActionNoChase || !candidate.Eligible {
		t.Fatalf("candidate=%+v", candidate)
	}
	if !candidate.WouldNotify || candidate.BudgetSlot != 1 || candidate.DeliveryMode != DeliveryModeShadowOnly {
		t.Fatalf("candidate does not respect the shadow budget contract: %+v", candidate)
	}
	if _, err := repo.AppendPolicyCandidate(ctx, candidate); err != nil {
		t.Fatal(err)
	}
}

func TestDailyCandidateBudgetCapsWouldNotifyAtFive(t *testing.T) {
	ctx := context.Background()
	repo := store.NewMemoryStore(nil)
	asOf := newYorkTime(9, 45)
	evaluator := Evaluator{Store: repo, Config: Config{DailyCandidateBudget: 5}}
	for index := 0; index < 6; index++ {
		symbol := fmt.Sprintf("TST%d", index)
		seedPolicyInputs(t, ctx, repo, symbol, asOf, 105, 100, 104, 5)
		if _, err := repo.AppendShadowSignal(ctx, store.ShadowSignalObservation{
			SignalType: shadow.SignalImpulseRejection, Symbol: symbol, AsOf: asOf, State: shadow.StateRejected,
			Eligible: true, ThresholdVersion: "impulse_shadow_v1", Metrics: json.RawMessage(`{"impulseMovePct":1.1}`),
		}); err != nil {
			t.Fatal(err)
		}
		candidate, err := evaluator.Evaluate(ctx, symbol, "custom", asOf)
		if err != nil {
			t.Fatal(err)
		}
		if index < 5 && (!candidate.WouldNotify || candidate.BudgetSlot != index+1) {
			t.Fatalf("candidate %d should occupy budget slot: %+v", index, candidate)
		}
		if index == 5 && (candidate.WouldNotify || candidate.SuppressedReason != "daily_candidate_budget_exhausted") {
			t.Fatalf("sixth candidate must be suppressed by daily budget: %+v", candidate)
		}
		if _, err := repo.AppendPolicyCandidate(ctx, candidate); err != nil {
			t.Fatal(err)
		}
	}
}

func TestOutcomeEvaluationRecordsObservedExcursionsWithoutChangingCandidateMode(t *testing.T) {
	ctx := context.Background()
	repo := store.NewMemoryStore(nil)
	asOf := newYorkTime(10, 10)
	candidate, err := repo.AppendPolicyCandidate(ctx, store.PolicyCandidateAlert{
		Symbol: "NVDA", TradingDate: asOf, AsOf: asOf, ProfileRole: "research_watch", Horizon: "intraday",
		SessionProduct: "regular", State: StateUnsupportedImpulse, ActionCandidate: ActionNoChase,
		ReferencePrice: 100, Eligible: true, WouldNotify: true, BudgetSlot: 1,
		DeliveryMode: DeliveryModeShadowOnly, ThresholdVersion: ThresholdVersion,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, point := range []struct {
		at    time.Time
		price float64
	}{
		{asOf, 100},
		{asOf.Add(5 * time.Minute), 103},
		{asOf.Add(10 * time.Minute), 98},
		{asOf.Add(15 * time.Minute), 101},
	} {
		if err := repo.AppendQuoteSnapshot(ctx, store.QuoteSnapshot{
			Symbol: "NVDA", Provider: "yahoo_chart", Price: point.price, ProviderTime: point.at,
			ReceivedAt: point.at, DataAgeSeconds: 1, Session: "regular",
		}); err != nil {
			t.Fatal(err)
		}
	}
	written, err := (OutcomeEvaluator{Store: repo, HorizonMinutes: []int{15}}).Evaluate(ctx, "NVDA", asOf.Add(15*time.Minute))
	if err != nil || written != 1 {
		t.Fatalf("Evaluate() written=%d err=%v", written, err)
	}
	outcomes, err := repo.ListPolicyCandidateOutcomes(ctx, candidate.ID)
	if err != nil || len(outcomes) != 1 {
		t.Fatalf("outcomes=%+v err=%v", outcomes, err)
	}
	if !almostEqual(outcomes[0].ReturnPercent, 1) || !almostEqual(outcomes[0].MaxUpPercent, 3) || !almostEqual(outcomes[0].MaxDownPercent, -2) {
		t.Fatalf("outcome=%+v", outcomes[0])
	}
	saved, err := repo.GetPolicyCandidate(ctx, candidate.ID)
	if err != nil || saved == nil || saved.DeliveryMode != DeliveryModeShadowOnly {
		t.Fatalf("candidate delivery mode changed after outcome evaluation: %+v err=%v", saved, err)
	}
}

func TestStoreRejectsAnyNonShadowPolicyCandidate(t *testing.T) {
	_, err := store.NewMemoryStore(nil).AppendPolicyCandidate(context.Background(), store.PolicyCandidateAlert{
		Symbol: "NVDA", AsOf: time.Now().UTC(), ProfileRole: "research_watch", Horizon: "intraday",
		SessionProduct: "regular", State: StateWaitConfirmation, ActionCandidate: ActionWaitConfirmation,
		DeliveryMode: "live", ThresholdVersion: ThresholdVersion,
	})
	if err == nil {
		t.Fatal("expected the store to reject a non-shadow policy candidate")
	}
}

func TestExplicitProfilePermissionsCanSuppressAnActionCandidate(t *testing.T) {
	ctx := context.Background()
	repo := store.NewMemoryStore(nil)
	asOf := newYorkTime(9, 45)
	seedPolicyInputs(t, ctx, repo, "NVDA", asOf, 105, 100, 104, 5)
	if _, err := repo.UpsertSymbolProfile(ctx, store.SymbolProfile{
		Symbol: "NVDA", Role: "tactical_watch", Horizon: "intraday",
		ActionPermissions: []byte(`["WAIT_CONFIRMATION"]`),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.AppendShadowSignal(ctx, store.ShadowSignalObservation{
		SignalType: shadow.SignalImpulseRejection, Symbol: "NVDA", AsOf: asOf, State: shadow.StateRejected,
		Eligible: true, ThresholdVersion: "impulse_shadow_v1", Metrics: json.RawMessage(`{"impulseMovePct":1.1}`),
	}); err != nil {
		t.Fatal(err)
	}
	candidate, err := (Evaluator{Store: repo}).Evaluate(ctx, "NVDA", "semis_memory", asOf)
	if err != nil {
		t.Fatal(err)
	}
	if candidate.ActionCandidate != ActionWaitConfirmation || candidate.Eligible || candidate.SuppressedReason != "action_not_permitted_by_symbol_profile" {
		t.Fatalf("explicit permissions were not enforced: %+v", candidate)
	}
}

func TestRangeCalibrationExcludesTheCurrentIncompleteTradingDate(t *testing.T) {
	ctx := context.Background()
	repo := store.NewMemoryStore(nil)
	asOf := newYorkTime(13, 0)
	previousTradingDay := asOf.AddDate(0, 0, -1)
	if err := repo.UpsertDailySummary(ctx, store.DailyMarketSummary{
		Symbol: "NVDA", Provider: "yahoo_chart", TradingDate: previousTradingDay,
		PreviousClose: 100, DailyHigh: 104, DailyLow: 98, SelectedMovePct: 5,
		SelectedSource: "options_resolved_iv",
	}); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpsertDailySummary(ctx, store.DailyMarketSummary{
		Symbol: "NVDA", Provider: "yahoo_chart", TradingDate: asOf,
		PreviousClose: 100, DailyHigh: 130, DailyLow: 70, SelectedMovePct: 5,
		SelectedSource: "options_resolved_iv",
	}); err != nil {
		t.Fatal(err)
	}
	calibration, err := CalibrateRange(ctx, repo, "NVDA", asOf, Config{RangeMinimumDays: 1})
	if err != nil {
		t.Fatal(err)
	}
	if calibration.SampleDays != 1 || !almostEqual(calibration.MeanRangeUse, 0.8) {
		t.Fatalf("current incomplete day leaked into calibration: %+v", calibration)
	}
}

func TestIVCalibrationMatchesPreOpenSnapshotToCompletedDailyRange(t *testing.T) {
	ctx := context.Background()
	repo := store.NewMemoryStore(nil)
	asOf := newYorkTime(13, 0)
	completedDay := time.Date(2026, time.July, 10, 0, 0, 0, 0, time.UTC)
	if err := repo.UpsertDailySummary(ctx, store.DailyMarketSummary{
		Symbol: "NVDA", Provider: "yahoo_chart", TradingDate: completedDay,
		PreviousClose: 100, DailyHigh: 106, DailyLow: 97, SelectedMovePct: 7,
		SelectedSource: "options_resolved_iv",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.SaveOptionChain(ctx, &options.ChainResponse{
		Symbol: "NVDA", Provider: options.ProviderYahooMCP, ReceivedAt: regularOpenAt(completedDay).Add(-5 * time.Minute),
		ExpirationDate: completedDay.Add(24 * time.Hour), Summary: options.ExpectedMove{
			OneTradingDayExpectedMovePercent: 7,
		},
	}); err != nil {
		t.Fatal(err)
	}
	calibration, err := CalibrateIV(ctx, repo, "NVDA", asOf, Config{RangeMinimumDays: 1})
	if err != nil {
		t.Fatal(err)
	}
	if calibration.MatchedDays != 1 || !almostEqual(calibration.MeanImpliedMovePercent, 7) || !almostEqual(calibration.MeanActualMovePercent, 6) || !almostEqual(calibration.MeanBiasPercent, 1) {
		t.Fatalf("IV calibration=%+v", calibration)
	}
}

func seedPolicyInputs(t *testing.T, ctx context.Context, repo *store.MemoryStore, symbol string, asOf time.Time, current, previousClose, regularOpen, selectedMovePct float64) {
	t.Helper()
	if _, err := repo.AppendEventContext(ctx, store.EventContextSnapshot{
		Symbol: symbol, AsOf: asOf, MarketDataFresh: true, QuoteAgeSeconds: 1, Session: "regular",
		EvidenceCoverage: "current", RiskState: "normal",
	}); err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(features.IntradayResponse{
		Symbol: symbol, ReceivedAt: asOf, CurrentPrice: current, PreviousClose: previousClose, RegularOpen: regularOpen,
		Realized: features.RealizedVol{Last30MinMovePercent: 0.1},
		ReasonableRange: features.ReasonableRange{
			SelectedOneDayMovePercent: selectedMovePct,
			SelectedSource:            "options_resolved_iv",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.AppendFeatureSnapshot(ctx, store.FeatureSnapshot{Symbol: symbol, ComputedAt: asOf, Payload: payload}); err != nil {
		t.Fatal(err)
	}
}

func newYorkTime(hour, minute int) time.Time {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		loc = time.UTC
	}
	return time.Date(2026, time.July, 13, hour, minute, 0, 0, loc).UTC()
}

func almostEqual(actual, expected float64) bool {
	return math.Abs(actual-expected) < 1e-9
}
