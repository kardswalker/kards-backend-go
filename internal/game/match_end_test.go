package game

import (
	"testing"
)

func TestEndMatchUsesReportedWinner(t *testing.T) {
	gm := &GameManager{}
	match := &Match{
		MatchID:           1,
		PlayerLeft:        10,
		PlayerRight:       20,
		Status:            "running",
		ActionIDSess:      77,
		ActionsData:       map[int]string{},
		PlayerStatusLeft:  "not_done",
		PlayerStatusRight: "not_done",
	}

	ended := gm.EndMatch(match, 20, MatchEndResult{
		WinnerSide: "right",
		WinnerID:   20,
		Reason:     "victory",
	})
	if !ended {
		t.Fatal("EndMatch returned false")
	}
	if match.WinnerSide != "right" {
		t.Fatalf("WinnerSide = %q, want right", match.WinnerSide)
	}
	if match.WinnerID != 20 {
		t.Fatalf("WinnerID = %d, want 20", match.WinnerID)
	}
	if len(match.Actions) != 1 || match.Actions[0] != 1 {
		t.Fatalf("Actions = %v, want [1]", match.Actions)
	}
	if match.ActionsData[1] == "" {
		t.Fatal("ActionsData[1] is empty")
	}
}

func TestEndMatchBySurrenderAwardsOpponent(t *testing.T) {
	gm := &GameManager{}
	match := &Match{
		MatchID:     1,
		PlayerLeft:  10,
		PlayerRight: 20,
		Status:      "running",
		ActionsData: map[int]string{},
	}
	gm.ActiveMatches.Store(match.MatchID, match)

	if !gm.EndMatchBySurrender(10, "surrender") {
		t.Fatal("EndMatchBySurrender returned false")
	}
	if match.WinnerSide != "right" {
		t.Fatalf("WinnerSide = %q, want right", match.WinnerSide)
	}
	if match.WinnerID != 20 {
		t.Fatalf("WinnerID = %d, want 20", match.WinnerID)
	}
}
