package game

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"fmt"
)

// CryptoIntn returns a uniform random integer in [0, n).
func CryptoIntn(n int) (int, error) {
	if n <= 0 {
		return 0, fmt.Errorf("invalid upper bound: %d", n)
	}

	if n == 1 {
		return 0, nil
	}

	max := uint64(n)
	limit := ^uint64(0) - (^uint64(0) % max)

	for {
		var buf [8]byte
		if _, err := cryptorand.Read(buf[:]); err != nil {
			return 0, err
		}

		v := binary.LittleEndian.Uint64(buf[:])
		if v < limit {
			return int(v % max), nil
		}
	}
}

// ShuffleCards shuffles cards in place with Fisher-Yates using crypto randomness.
func ShuffleCards(cards []Card) error {
	for i := len(cards) - 1; i > 0; i-- {
		j, err := CryptoIntn(i + 1)
		if err != nil {
			return err
		}
		cards[i], cards[j] = cards[j], cards[i]
	}
	return nil
}

// ShuffleDeckCards reshuffles deck order and synchronizes deck positions.
func ShuffleDeckCards(cards []Card, location string) error {
	if err := ShuffleCards(cards); err != nil {
		return err
	}

	for i := range cards {
		cards[i].Location = location
		cards[i].LocationNumber = i
	}

	return nil
}
