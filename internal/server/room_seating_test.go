package server

import (
	"math/rand"
	"testing"

	"starcup-engine/internal/model"
)

func maxCampRun(lineup []lineupPlayer) int {
	if len(lineup) == 0 {
		return 0
	}
	maxRun := 1
	run := 1
	for i := 1; i < len(lineup); i++ {
		if lineup[i].camp == lineup[i-1].camp {
			run++
			if run > maxRun {
				maxRun = run
			}
			continue
		}
		run = 1
	}
	return maxRun
}

func TestBuildInterleavedLineup_AvoidsThreeSameCampInRow(t *testing.T) {
	base := []lineupPlayer{
		{id: "p1", camp: model.RedCamp},
		{id: "p2", camp: model.RedCamp},
		{id: "p3", camp: model.RedCamp},
		{id: "p4", camp: model.BlueCamp},
		{id: "p5", camp: model.BlueCamp},
		{id: "p6", camp: model.BlueCamp},
	}

	for seed := int64(1); seed <= 64; seed++ {
		lineup := buildInterleavedLineup(base, rand.New(rand.NewSource(seed)))
		if got := maxCampRun(lineup); got >= 3 {
			t.Fatalf("seed=%d got max same-camp run=%d, lineup=%+v", seed, got, lineup)
		}
	}
}

func TestBuildInterleavedLineup_ThreeVsOneStillAvoidsTriple(t *testing.T) {
	base := []lineupPlayer{
		{id: "p1", camp: model.RedCamp},
		{id: "p2", camp: model.RedCamp},
		{id: "p3", camp: model.RedCamp},
		{id: "p4", camp: model.BlueCamp},
	}

	for seed := int64(1); seed <= 64; seed++ {
		lineup := buildInterleavedLineup(base, rand.New(rand.NewSource(seed)))
		if got := maxCampRun(lineup); got >= 3 {
			t.Fatalf("seed=%d got max same-camp run=%d, lineup=%+v", seed, got, lineup)
		}
	}
}

func TestOrderedClientIDsLocked_SeatOrderFirst(t *testing.T) {
	room := NewRoom("seat-order")
	room.Clients = map[string]*Client{
		"p1": {PlayerID: "p1"},
		"p2": {PlayerID: "p2"},
		"p3": {PlayerID: "p3"},
	}
	room.SeatOrder = []string{"p3", "p1"}

	ids := room.orderedClientIDsLocked()
	want := []string{"p3", "p1", "p2"}
	if len(ids) != len(want) {
		t.Fatalf("ids len mismatch: got=%d want=%d ids=%v", len(ids), len(want), ids)
	}
	for i := range want {
		if ids[i] != want[i] {
			t.Fatalf("index %d mismatch: got=%s want=%s (ids=%v)", i, ids[i], want[i], ids)
		}
	}
}
