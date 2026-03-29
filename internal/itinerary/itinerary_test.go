package itinerary

import (
	"context"
	"testing"
)

func TestGenerateItineraryRejectsInvalidDays(t *testing.T) {
	_, err := GenerateItinerary(context.Background(), nil, 0, 0, nil, 0)
	if err == nil {
		t.Fatal("expected error when days <= 0")
	}
}

func TestGenerateItineraryReturnsEmptyDaysForNoMatches(t *testing.T) {
	days := 3
	it, err := GenerateItinerary(context.Background(), nil, 0, 0, nil, days)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(it) != days {
		t.Fatalf("expected %d day buckets, got %d", days, len(it))
	}
	for i, day := range it {
		if len(day) != 0 {
			t.Fatalf("expected day %d to be empty, got %d places", i+1, len(day))
		}
	}
}
