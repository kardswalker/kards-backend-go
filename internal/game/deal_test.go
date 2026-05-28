package game

import "testing"

func TestDealStartingCardsAssignsLocations(t *testing.T) {
	cards := []Card{
		{CardID: 1, Location: "deck_left", LocationNumber: 0},
		{CardID: 2, Location: "deck_left", LocationNumber: 1},
		{CardID: 3, Location: "deck_left", LocationNumber: 2},
	}

	hand, deck := dealStartingCards(cards, 2, "hand_left", "deck_left")

	if len(hand) != 2 {
		t.Fatalf("len(hand) = %d, want 2", len(hand))
	}
	if len(deck) != 1 {
		t.Fatalf("len(deck) = %d, want 1", len(deck))
	}
	if hand[0].Location != "hand_left" || hand[0].LocationNumber != 0 {
		t.Fatalf("hand[0] = %+v, want hand_left location 0", hand[0])
	}
	if hand[1].Location != "hand_left" || hand[1].LocationNumber != 1 {
		t.Fatalf("hand[1] = %+v, want hand_left location 1", hand[1])
	}
	if deck[0].Location != "deck_left" || deck[0].LocationNumber != 0 {
		t.Fatalf("deck[0] = %+v, want deck_left location 0", deck[0])
	}
}
