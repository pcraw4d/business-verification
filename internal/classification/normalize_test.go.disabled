package classification

import "testing"

func TestNormalizeText(t *testing.T) {
	got := normalizeText("Acme, Inc. — Software & Consulting!")
	want := "acme inc software consulting"
	if got != want {
		t.Fatalf("normalizeText() = %q, want %q", got, want)
	}
}

func TestTokenizeRemovesStopwordsAndDedups(t *testing.T) {
	norm := "the acme company group acme software"
	toks := tokenize(norm)
	// stopwords: the, company, group are removed; unique: acme, software
	if len(toks) != 2 {
		t.Fatalf("expected 2 tokens, got %d: %v", len(toks), toks)
	}
	if toks[0] != "acme" || toks[1] != "software" {
		t.Fatalf("unexpected tokens: %v", toks)
	}
}

func TestNormalizeBusinessFields(t *testing.T) {
	norm, toks := normalizeBusinessFields("Acme, Inc.", "Custom software development", "software, consulting")
	if norm == "" || len(toks) == 0 {
		t.Fatalf("expected normalized text and tokens, got %q / %v", norm, toks)
	}
}
