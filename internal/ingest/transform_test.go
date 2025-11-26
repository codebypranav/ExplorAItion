package ingest

import (
	"strings"
	"testing"
)

func TestBuildTextForEmbedding(t *testing.T) {
	p := POI{Name: "Eiffel Tower", Description: "Tall tower in Paris", Kinds: "tourism,architecture", Country: "France"}
	text := BuildTextForEmbedding(p)
	if text == "" {
		t.Fatalf("expected non-empty text")
	}
	if !(strings.Contains(text, "Eiffel Tower") && strings.Contains(text, "France") && strings.Contains(text, "architecture")) {
		t.Fatalf("expected name, country, and kind in text: %s", text)
	}
}
