package timeutil

import (
	"testing"
	"time"
)

func TestNowUTC(t *testing.T) {
	now := NowUTC()
	if now.Location() != time.UTC {
		t.Errorf("NowUTC() should return UTC time, got %v", now.Location())
	}
}

func TestToUTC(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	localTime := time.Date(2025, 1, 8, 14, 30, 0, 0, loc)
	utcTime := ToUTC(localTime)

	if utcTime.Location() != time.UTC {
		t.Errorf("ToUTC() should return UTC time")
	}
}

func TestFormatForMessage(t *testing.T) {
	testTime := time.Date(2025, 1, 8, 14, 30, 0, 0, time.UTC)
	formatted := FormatForMessage(testTime)
	expected := "2025-01-08 14:30:00 UTC"
	if formatted != expected {
		t.Errorf("FormatForMessage() = %s, want %s", formatted, expected)
	}
}

func TestFormatShort(t *testing.T) {
	testTime := time.Date(2025, 1, 8, 14, 30, 0, 0, time.UTC)
	formatted := FormatShort(testTime)
	expected := "Jan 8 14:30 UTC"
	if formatted != expected {
		t.Errorf("FormatShort() = %s, want %s", formatted, expected)
	}
}

func TestIsExpired(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour)
	future := time.Now().Add(1 * time.Hour)

	if !IsExpired(past) {
		t.Errorf("IsExpired(past) should return true")
	}
	if IsExpired(future) {
		t.Errorf("IsExpired(future) should return false")
	}
}

func TestMinutesUntil(t *testing.T) {
	future := time.Now().Add(30 * time.Minute)
	minutes := MinutesUntil(future)
	if minutes < 29 || minutes > 31 {
		t.Errorf("MinutesUntil(future) = %d, want ~30", minutes)
	}

	past := time.Now().Add(-30 * time.Minute)
	minutes = MinutesUntil(past)
	if minutes > -30 || minutes < -31 {
		t.Errorf("MinutesUntil(past) = %d, want ~-30", minutes)
	}
}

func TestAddMinutes(t *testing.T) {
	base := time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC)
	result := AddMinutes(base, 30)
	expected := time.Date(2025, 1, 8, 14, 30, 0, 0, time.UTC)
	if !result.Equal(expected) {
		t.Errorf("AddMinutes() = %v, want %v", result, expected)
	}
}

func TestSubtractMinutes(t *testing.T) {
	base := time.Date(2025, 1, 8, 14, 30, 0, 0, time.UTC)
	result := SubtractMinutes(base, 30)
	expected := time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC)
	if !result.Equal(expected) {
		t.Errorf("SubtractMinutes() = %v, want %v", result, expected)
	}
}

func TestDurationInMinutes(t *testing.T) {
	duration := 90 * time.Minute
	minutes := DurationInMinutes(duration)
	if minutes != 90 {
		t.Errorf("DurationInMinutes() = %d, want 90", minutes)
	}
}

func TestToUnixMillis(t *testing.T) {
	testTime := time.Date(2025, 1, 8, 14, 30, 0, 0, time.UTC)
	ms := ToUnixMillis(testTime)
	if ms <= 0 {
		t.Errorf("ToUnixMillis() returned non-positive value: %d", ms)
	}
}

func TestToUnixSeconds(t *testing.T) {
	testTime := time.Date(2025, 1, 8, 14, 30, 0, 0, time.UTC)
	s := ToUnixSeconds(testTime)
	if s <= 0 {
		t.Errorf("ToUnixSeconds() returned non-positive value: %d", s)
	}
}

func TestFromUnixMillis(t *testing.T) {
	ms := int64(1736340600000) // 2025-01-08 14:30:00 UTC
	tm := FromUnixMillis(ms)
	if tm.Location() != time.UTC {
		t.Errorf("FromUnixMillis() should return UTC time")
	}
}

func TestFromUnixSeconds(t *testing.T) {
	s := int64(1736340600) // 2025-01-08 14:30:00 UTC
	tm := FromUnixSeconds(s)
	if tm.Location() != time.UTC {
		t.Errorf("FromUnixSeconds() should return UTC time")
	}
}

func TestCalculateDecisionDeadline(t *testing.T) {
	reviewTime := time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC)
	deadline := CalculateDecisionDeadline(reviewTime, 20)
	expected := time.Date(2025, 1, 8, 13, 40, 0, 0, time.UTC)
	if !deadline.Equal(expected) {
		t.Errorf("CalculateDecisionDeadline() = %v, want %v", deadline, expected)
	}
}

func TestCalculateNonWhitelistCancelTime(t *testing.T) {
	// Freeze time for testing
	baseTime := time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC)

	// This test uses current time, so we just check it returns a future time
	cancelTime := CalculateNonWhitelistCancelTime(5)
	if time.Now().Add(4 * time.Minute).After(cancelTime) {
		t.Errorf("CalculateNonWhitelistCancelTime() should return time > 4 minutes from now")
	}
}

func TestShouldShiftSlot(t *testing.T) {
	// Slot 20 minutes from now with threshold of 25
	nearFuture := time.Now().Add(20 * time.Minute)
	if !ShouldShiftSlot(nearFuture, 25) {
		t.Errorf("ShouldShiftSlot(nearFuture, 25) should return true")
	}

	// Slot 30 minutes from now with threshold of 25
	farFuture := time.Now().Add(30 * time.Minute)
	if ShouldShiftSlot(farFuture, 25) {
		t.Errorf("ShouldShiftSlot(farFuture, 25) should return false")
	}
}

func TestCalculateSlotDuration(t *testing.T) {
	start := time.Date(2025, 1, 8, 14, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 8, 15, 30, 0, 0, time.UTC)
	duration := CalculateSlotDuration(start, end)
	if duration != 90 {
		t.Errorf("CalculateSlotDuration() = %d, want 90", duration)
	}
}
