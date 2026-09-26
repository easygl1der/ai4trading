package store

import (
	"context"
	"testing"
	"time"
)

func TestMemorySymbolProfilesAndPeerMapsRemainExplicitAndReviewable(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryStore(nil)
	profile, err := repo.UpsertSymbolProfile(ctx, SymbolProfile{
		Symbol: "nvda", Role: "tactical_watch", Horizon: "intraday",
		ActionPermissions: []byte(`["WAIT_CONFIRMATION","NO_CHASE"]`),
	})
	if err != nil || profile.Symbol != "NVDA" || profile.Role != "tactical_watch" {
		t.Fatalf("profile=%+v err=%v", profile, err)
	}

	proposed, err := repo.CreatePeerMapVersion(ctx, PeerMapVersion{
		MapKey: "semis", GroupName: "semis_memory", BenchmarkSymbol: "SMH",
		Methodology: "reviewable test map", ReviewStatus: "proposed",
		Members: []PeerMapMember{{Symbol: "NVDA", Relation: "leader"}, {Symbol: "MU", Relation: "peer"}},
	})
	if err != nil || proposed.Version != 1 {
		t.Fatalf("proposed=%+v err=%v", proposed, err)
	}
	if active, err := repo.GetActivePeerMap(ctx, "NVDA", "semis_memory", time.Now().UTC()); err != nil || active != nil {
		t.Fatalf("proposed map must not become active: active=%+v err=%v", active, err)
	}

	approved, err := repo.CreatePeerMapVersion(ctx, PeerMapVersion{
		MapKey: "semis", GroupName: "semis_memory", BenchmarkSymbol: "SMH",
		Methodology: "reviewed test map", ReviewStatus: "approved",
		Members: []PeerMapMember{{Symbol: "NVDA", Relation: "leader"}, {Symbol: "MU", Relation: "peer"}},
	})
	if err != nil || approved.Version != 2 {
		t.Fatalf("approved=%+v err=%v", approved, err)
	}
	active, err := repo.GetActivePeerMap(ctx, "NVDA", "semis_memory", time.Now().UTC())
	if err != nil || active == nil || active.ID != approved.ID || active.ReviewStatus != "approved" {
		t.Fatalf("active=%+v err=%v", active, err)
	}
}
