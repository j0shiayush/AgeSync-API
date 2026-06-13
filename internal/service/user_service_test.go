package service_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"AgeSync-API/internal/service"
)

// fixed returns a UTC time.Time for the given date, making test expectations deterministic.
func fixed(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func TestCalculateAge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		dob      time.Time
		now      time.Time
		wantAge  int
	}{
		{
			name:    "exact birthday today returns correct age",
			dob:     fixed(1990, 5, 10),
			now:     fixed(2025, 5, 10),
			wantAge: 35,
		},
		{
			name:    "birthday has already passed this year",
			dob:     fixed(1990, 1, 1),
			now:     fixed(2025, 6, 15),
			wantAge: 35,
		},
		{
			name:    "birthday has not yet occurred this year",
			dob:     fixed(1990, 12, 31),
			now:     fixed(2025, 6, 15),
			wantAge: 34,
		},
		{
			name:    "leap year birthday on 29 Feb — non-leap now year",
			dob:     fixed(1992, 2, 29),
			now:     fixed(2025, 2, 28), // before the 29th (which doesn't exist)
			wantAge: 32,
		},
		{
			name:    "leap year birthday on 29 Feb — non-leap now year after Mar 1",
			dob:     fixed(1992, 2, 29),
			now:     fixed(2025, 3, 1),
			wantAge: 33,
		},
		{
			name:    "leap year birthday on 29 Feb — leap now year exact day",
			dob:     fixed(1992, 2, 29),
			now:     fixed(2024, 2, 29),
			wantAge: 32,
		},
		{
			name:    "person born today (age 0)",
			dob:     fixed(2025, 6, 15),
			now:     fixed(2025, 6, 15),
			wantAge: 0,
		},
		{
			name:    "newborn — day before birthday",
			dob:     fixed(2025, 6, 16),
			now:     fixed(2025, 6, 15),
			wantAge: -1, // dob is in the future relative to now
		},
		{
			name:    "century age",
			dob:     fixed(1920, 3, 20),
			now:     fixed(2025, 3, 20),
			wantAge: 105,
		},
		{
			name:    "birthday tomorrow — has not occurred",
			dob:     fixed(1985, 6, 16),
			now:     fixed(2025, 6, 15),
			wantAge: 39,
		},
		{
			name:    "birthday yesterday — has occurred",
			dob:     fixed(1985, 6, 14),
			now:     fixed(2025, 6, 15),
			wantAge: 40,
		},
		{
			name:    "year boundary — born Dec 31, now Jan 1",
			dob:     fixed(2000, 12, 31),
			now:     fixed(2025, 1, 1),
			wantAge: 24,
		},
		{
			name:    "year boundary — born Jan 1, now Dec 31",
			dob:     fixed(2000, 1, 1),
			now:     fixed(2025, 12, 31),
			wantAge: 25,
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := service.CalculateAge(tc.dob, tc.now)
			assert.Equal(t, tc.wantAge, got,
				"CalculateAge(dob=%s, now=%s) = %d; want %d",
				tc.dob.Format("2006-01-02"), tc.now.Format("2006-01-02"), got, tc.wantAge,
			)
		})
	}
}

// TestCalculateAge_Idempotent ensures the function is pure — calling it twice
// with the same inputs always returns the same result.
func TestCalculateAge_Idempotent(t *testing.T) {
	t.Parallel()

	dob := fixed(1990, 5, 10)
	now := fixed(2025, 6, 12)

	first := service.CalculateAge(dob, now)
	second := service.CalculateAge(dob, now)

	assert.Equal(t, first, second, "CalculateAge must be a pure function")
}

// TestCalculateAge_DoesNotMutateInputs verifies the time.Time values are not
// modified inside the function (they are value types in Go, but good to assert).
func TestCalculateAge_DoesNotMutateInputs(t *testing.T) {
	t.Parallel()

	dob := fixed(1990, 5, 10)
	now := fixed(2025, 6, 12)

	dobBefore := dob
	nowBefore := now

	_ = service.CalculateAge(dob, now)

	assert.Equal(t, dobBefore, dob)
	assert.Equal(t, nowBefore, now)
}