package main

import (
	"testing"
	"time"
)

func TestEvent_GenerateKey(t *testing.T) {
	t.Run("with ID", func(t *testing.T) {
		e := &Event{
			Source: "TN",
			ID:     "123",
			Time:   time.Now(),
		}
		e.GenerateKey()
		if e.Key != "TN-123" {
			t.Errorf("expected key TN-123, got %s", e.Key)
		}
	})

	t.Run("without ID", func(t *testing.T) {
		tm, _ := time.Parse(timeLayout, "2024/01/01 12:00:00")
		e := &Event{
			Source:      "KH",
			Time:        tm,
			Category:    "Fire",
			Subcategory: "Big",
			Location:    "CityCenter",
		}
		e.GenerateKey()

		// Expected format: Source-Time-Category-Subcategory-Location
		expected := "KH-2024/01/01 12:00:00-Fire-Big-CityCenter"
		if e.Key != expected {
			t.Errorf("expected key %s, got %s", expected, e.Key)
		}
	})
}
