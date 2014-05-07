package main

import (
	"testing"
	"time"
)

func AssertEquals(t *testing.T, value interface{}, expected interface{}) {
	if value != expected {
		if _, ok := value.(int64); ok {
			t.Fatalf("got %d, expected %d", value, expected)
		} else {
			// Fall back to strings.
			t.Fatalf("got %s, expected %s", value, expected)
		}
	}
}

func TestParseTime(t *testing.T) {

	// Basic RFC3339 value at UTC.
	value := ParseTime("2014-04-30T21:54:09Z", "")
	if value == nil {
		t.Fatalf("unexpected nil")
	}
	if value.Unix() != 1398894849 {
		t.Fatalf("expected 1398894849, got %d", value.Unix())
	}

	// With timezone offsets.
	value = ParseTime("2014-04-30T15:56:42-06:00", "")
	if value == nil {
		t.Fatalf("unexpected nil")
	}
	if value.Unix() != 1398895002 {
		t.Fatalf("expected 1398895002, got %d", value.Unix())
	}
}

// Test durations the UI may present, or we provide in the
// documentation.
func TestParseDuration(t *testing.T) {

	var value time.Duration
	var err error

	value, err = time.ParseDuration("1m")
	if err != nil {
		t.Fatal(err)
	}
	AssertEquals(t, int64(value.Seconds()), int64(60))

	value, err = time.ParseDuration("15s")
	if err != nil {
		t.Fatal(err)
	}
	AssertEquals(t, int64(value.Seconds()), int64(15))

	value, err = time.ParseDuration("1h")
	if err != nil {
		t.Fatal(err)
	}
	AssertEquals(t, int64(value.Seconds()), int64(3600))
}
