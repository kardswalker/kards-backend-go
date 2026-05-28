package deckcode

import (
	"fmt"
	"strings"
)

var CountryMap = map[string]string{
	"1": "Germany", "2": "Britain", "3": "Japan",
	"4": "Soviet", "5": "USA", "6": "France",
	"7": "Italy", "8": "Poland", "9": "Finland",
}

var countryCodeByFaction = map[string]string{
	"germany": "1",
	"britain": "2",
	"japan":   "3",
	"soviet":  "4",
	"usa":     "5",
	"france":  "6",
	"italy":   "7",
	"poland":  "8",
	"finland": "9",
}

var defaultHQByFaction = map[string]string{
	"soviet":  "8v",
	"usa":     "ce",
	"britain": "0N",
	"japan":   "6l",
	"germany": "3v",
}

type ParsedDeck struct {
	MainCountry string
	AllyCountry string
	Cards       map[string]int
	HQ          string
}

func ParseDeckCode(deckCode string) (*ParsedDeck, error) {
	if !strings.HasPrefix(deckCode, "%%") {
		return nil, fmt.Errorf("invalid prefix")
	}

	countryPart, cardsPart, hqPart, err := splitDeckCode(deckCode)
	if err != nil {
		return nil, err
	}

	deck := &ParsedDeck{
		MainCountry: CountryMap[string(countryPart[0])],
		AllyCountry: CountryMap[string(countryPart[1])],
		Cards:       make(map[string]int),
		HQ:          hqPart,
	}

	groups := strings.Split(cardsPart, ";")
	if len(groups) != 4 {
		return nil, fmt.Errorf("invalid card group count")
	}

	if deck.HQ == "" {
		if defaultHQ, ok := DefaultHQCode(deck.MainCountry); ok && strings.HasPrefix(groups[0], defaultHQ) {
			deck.HQ = defaultHQ
			groups[0] = groups[0][len(defaultHQ):]
		}
	}

	if deck.HQ == "" {
		if defaultHQ, ok := DefaultHQCode(deck.MainCountry); ok {
			deck.HQ = defaultHQ
		}
	}

	for i, group := range groups {
		multiplier := i + 1
		for j := 0; j < len(group); j += 2 {
			if j+1 < len(group) {
				id := group[j : j+2]
				deck.Cards[id] += multiplier
			}
		}
	}

	return deck, nil
}

func CountryCodeFromFaction(faction string) (string, bool) {
	code, ok := countryCodeByFaction[strings.ToLower(strings.TrimSpace(faction))]
	return code, ok
}

func BuildDefaultDeckCode(mainFaction, allyFaction string) string {
	mainCode, ok := CountryCodeFromFaction(mainFaction)
	if !ok {
		mainCode = "1"
	}

	allyCode, ok := CountryCodeFromFaction(allyFaction)
	if !ok {
		allyCode = mainCode
	}

	hqCode, ok := DefaultHQCode(mainFaction)
	if !ok {
		return "%%" + mainCode + allyCode + "|;;;"
	}

	return "%%" + mainCode + allyCode + "|;;;|" + hqCode
}

func DefaultHQCode(mainFaction string) (string, bool) {
	hq, ok := defaultHQByFaction[strings.ToLower(strings.TrimSpace(mainFaction))]
	return hq, ok
}

func EnsureDeckCodeHQ(deckCode, mainFaction string) string {
	if deckCode == "" {
		return BuildDefaultDeckCode(mainFaction, mainFaction)
	}
	if !strings.HasPrefix(deckCode, "%%") {
		return deckCode
	}

	hqCode, ok := DefaultHQCode(mainFaction)
	if !ok {
		return deckCode
	}

	countryPart, cardsPart, existingHQ, err := splitDeckCode(deckCode)
	if err != nil {
		return deckCode
	}

	groups := strings.Split(normalizeCardsPart(cardsPart), ";")
	if len(groups) == 4 && existingHQ == "" && strings.HasPrefix(groups[0], hqCode) {
		groups[0] = groups[0][len(hqCode):]
	}

	if existingHQ == "" {
		existingHQ = hqCode
	}

	return "%%" + countryPart + "|" + strings.Join(groups, ";") + "|" + existingHQ
}

func splitDeckCode(deckCode string) (countryPart, cardsPart, hqPart string, err error) {
	content := strings.TrimPrefix(deckCode, "%%")
	parts := strings.Split(content, "|")
	if len(parts) < 2 || len(parts) > 3 {
		return "", "", "", fmt.Errorf("invalid format")
	}

	countryPart = parts[0]
	if len(countryPart) != 2 {
		return "", "", "", fmt.Errorf("invalid country code")
	}

	cardsPart = parts[1]
	if idx := strings.Index(cardsPart, "~"); idx != -1 {
		cardsPart = cardsPart[:idx]
	}

	if len(parts) == 3 {
		hqPart = parts[2]
	}

	return countryPart, cardsPart, hqPart, nil
}

func normalizeCardsPart(cardsPart string) string {
	groups := strings.Split(cardsPart, ";")
	for len(groups) < 4 {
		groups = append(groups, "")
	}
	if len(groups) > 4 {
		groups = groups[:4]
	}
	return strings.Join(groups, ";")
}
