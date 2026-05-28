package deckcode

import "testing"

func TestParseDeckCodeWithSeparateHQPart(t *testing.T) {
	parsed, err := ParseDeckCode("%%11|0Y;1A;;2B~0Y;1A;;2B|3v1i")
	if err != nil {
		t.Fatalf("ParseDeckCode returned error: %v", err)
	}

	if parsed.MainCountry != "Germany" {
		t.Fatalf("MainCountry = %q, want %q", parsed.MainCountry, "Germany")
	}
	if parsed.AllyCountry != "Germany" {
		t.Fatalf("AllyCountry = %q, want %q", parsed.AllyCountry, "Germany")
	}
	if parsed.HQ != "3v1i" {
		t.Fatalf("HQ = %q, want %q", parsed.HQ, "3v1i")
	}
	if parsed.Cards["0Y"] != 1 {
		t.Fatalf("Cards[0Y] = %d, want 1", parsed.Cards["0Y"])
	}
	if parsed.Cards["1A"] != 2 {
		t.Fatalf("Cards[1A] = %d, want 2", parsed.Cards["1A"])
	}
	if parsed.Cards["2B"] != 4 {
		t.Fatalf("Cards[2B] = %d, want 4", parsed.Cards["2B"])
	}
}

func TestParseDeckCodeWithEmbeddedHQPart(t *testing.T) {
	parsed, err := ParseDeckCode("%%11|3v0Y;1A;;2B")
	if err != nil {
		t.Fatalf("ParseDeckCode returned error: %v", err)
	}

	if parsed.HQ != "3v" {
		t.Fatalf("HQ = %q, want %q", parsed.HQ, "3v")
	}
	if parsed.Cards["0Y"] != 1 {
		t.Fatalf("Cards[0Y] = %d, want 1", parsed.Cards["0Y"])
	}
	if parsed.Cards["1A"] != 2 {
		t.Fatalf("Cards[1A] = %d, want 2", parsed.Cards["1A"])
	}
	if parsed.Cards["2B"] != 4 {
		t.Fatalf("Cards[2B] = %d, want 4", parsed.Cards["2B"])
	}
}

func TestBuildDefaultDeckCode(t *testing.T) {
	if got := BuildDefaultDeckCode("Germany", "USA"); got != "%%15|;;;|3v" {
		t.Fatalf("BuildDefaultDeckCode = %q, want %q", got, "%%15|;;;|3v")
	}
}

func TestEnsureDeckCodeHQConvertsEmbeddedFormat(t *testing.T) {
	got := EnsureDeckCodeHQ("%%11|3v0Y;1A;;2B", "Germany")
	want := "%%11|0Y;1A;;2B|3v"
	if got != want {
		t.Fatalf("EnsureDeckCodeHQ = %q, want %q", got, want)
	}
}

func TestEnsureDeckCodeHQAddsMissingHQPart(t *testing.T) {
	got := EnsureDeckCodeHQ("%%11|0Y;1A;;2B", "Germany")
	want := "%%11|0Y;1A;;2B|3v"
	if got != want {
		t.Fatalf("EnsureDeckCodeHQ = %q, want %q", got, want)
	}
}

func TestEnsureDeckCodeHQBuildsEmptyDeck(t *testing.T) {
	got := EnsureDeckCodeHQ("%%11|", "Germany")
	want := "%%11|;;;|3v"
	if got != want {
		t.Fatalf("EnsureDeckCodeHQ = %q, want %q", got, want)
	}
}
