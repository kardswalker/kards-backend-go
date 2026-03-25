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

type ParsedDeck struct {
	MainCountry string
	AllyCountry string
	Cards       map[string]int // cardCode -> count
	HQ          string
}

func ParseDeckCode(deckCode string) (*ParsedDeck, error) {
	if !strings.HasPrefix(deckCode, "%%") {
		return nil, fmt.Errorf("invalid prefix")
	}

	content := strings.TrimPrefix(deckCode, "%%")
	parts := strings.Split(content, "|")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid format")
	}

	// 1. 解析国家
	countryPart := parts[0]
	if len(countryPart) < 2 {
		return nil, fmt.Errorf("invalid country code")
	}

	deck := &ParsedDeck{
		MainCountry: CountryMap[string(countryPart[0])],
		AllyCountry: CountryMap[string(countryPart[1])],
		Cards:       make(map[string]int),
	}

	// 2. 解析卡牌与重复处理 (波浪号 ~)
	cardsPart := parts[1]
	if idx := strings.Index(cardsPart, "~"); idx != -1 {
		cardsPart = cardsPart[:idx]
	}

	// 分割 4 个数量组 (; 隔开)
	groups := strings.Split(cardsPart, ";")
	for i, group := range groups {
		multiplier := i + 1 // 第一组x1, 第二组x2...
		for j := 0; j < len(group); j += 2 {
			if j+1 < len(group) {
				id := group[j : j+2]
				deck.Cards[id] += multiplier
			}
		}
	}

	// 3. 解析总部 (如果有)
	if len(parts) == 3 {
		deck.HQ = parts[2]
	}

	return deck, nil
}
