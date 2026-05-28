package game

func dealStartingCards(cards []Card, handSize int, handLocation, deckLocation string) ([]Card, []Card) {
	hand := make([]Card, 0, min(handSize, len(cards)))
	deck := make([]Card, 0, max(0, len(cards)-handSize))

	for i := range cards {
		card := cards[i]
		if i < handSize {
			card.Location = handLocation
			card.LocationNumber = len(hand)
			hand = append(hand, card)
			continue
		}

		card.Location = deckLocation
		card.LocationNumber = len(deck)
		deck = append(deck, card)
	}

	return hand, deck
}
