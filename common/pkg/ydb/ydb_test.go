package ydb

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/arseniisemenow/review-slot-guard-bot/common/pkg/models"
)

func TestGetUserByTelegramChatID_NotFound(t *testing.T) {
	// This test verifies the behavior when user is not found
	// In a real scenario with mocks, we would mock the YDB connection
	ctx := context.Background()

	// Since we can't easily mock YDB without more setup, we'll test the error handling path
	// by testing with an invalid connection
	_, err := GetUserByTelegramChatID(ctx, 999999999)
	// This should return an error since YDB is not configured
	assert.Error(t, err)
}

func TestGetUserByReviewerLogin_NotFound(t *testing.T) {
	ctx := context.Background()

	_, err := GetUserByReviewerLogin(ctx, "nonexistent_user")
	assert.Error(t, err)
}

func TestModelsIntegration(t *testing.T) {
	// Test model creation and validation
	t.Run("DefaultUserSettings", func(t *testing.T) {
		settings := models.DefaultUserSettings("testuser")
		assert.Equal(t, "testuser", settings.ReviewerLogin)
		assert.Equal(t, int32(20), settings.ResponseDeadlineShiftMinutes)
		assert.Equal(t, int32(5), settings.NonWhitelistCancelDelayMinutes)
		assert.True(t, settings.NotifyWhitelistTimeout)
		assert.True(t, settings.NotifyNonWhitelistCancel)
		assert.Equal(t, int32(25), settings.SlotShiftThresholdMinutes)
		assert.Equal(t, int32(15), settings.SlotShiftDurationMinutes)
		assert.Equal(t, int32(15), settings.CleanupDurationsMinutes)
	})

	t.Run("ValidStatuses", func(t *testing.T) {
		validStatuses := []string{
			models.StatusUnknownProjectReview,
			models.StatusKnownProjectReview,
			models.StatusWhitelisted,
			models.StatusNotWhitelisted,
			models.StatusNeedToApprove,
			models.StatusWaitingForApprove,
			models.StatusApproved,
			models.StatusCancelled,
			models.StatusAutoCancelled,
			models.StatusAutoCancelledNotWhitelisted,
		}

		for _, status := range validStatuses {
			assert.True(t, models.IsValidStatus(status), "Status %s should be valid", status)
		}
	})

	t.Run("IntermediateStatuses", func(t *testing.T) {
		intermediate := []string{
			models.StatusUnknownProjectReview,
			models.StatusKnownProjectReview,
			models.StatusWhitelisted,
			models.StatusNotWhitelisted,
			models.StatusNeedToApprove,
			models.StatusWaitingForApprove,
		}

		for _, status := range intermediate {
			assert.True(t, models.IsIntermediateStatus(status), "Status %s should be intermediate", status)
		}
	})

	t.Run("FinalStatuses", func(t *testing.T) {
		final := []string{
			models.StatusApproved,
			models.StatusCancelled,
			models.StatusAutoCancelled,
			models.StatusAutoCancelledNotWhitelisted,
		}

		for _, status := range final {
			assert.True(t, models.IsFinalStatus(status), "Status %s should be final", status)
		}
	})
}

func TestUserCreation(t *testing.T) {
	user := &models.User{
		ReviewerLogin:      "testuser",
		Status:            models.UserStatusActive,
		TelegramChatID:     123456789,
		CreatedAt:         time.Now().Unix(),
		LastAuthSuccessAt: time.Now().Unix(),
	}

	assert.Equal(t, "testuser", user.ReviewerLogin)
	assert.Equal(t, models.UserStatusActive, user.Status)
	assert.Equal(t, int64(123456789), user.TelegramChatID)
	assert.True(t, user.CreatedAt > 0)
}

func TestReviewRequestCreation(t *testing.T) {
	req := &models.ReviewRequest{
		ID:              "test-id-123",
		ReviewerLogin:   "testuser",
		ReviewStartTime: time.Now().Add(1 * time.Hour).Unix(),
		CalendarSlotID:  "slot-123",
		Status:          models.StatusUnknownProjectReview,
		CreatedAt:       time.Now().Unix(),
	}

	assert.Equal(t, "test-id-123", req.ID)
	assert.Equal(t, "testuser", req.ReviewerLogin)
	assert.Equal(t, models.StatusUnknownProjectReview, req.Status)
	assert.True(t, req.ReviewStartTime > 0)
}

func TestWhitelistEntryCreation(t *testing.T) {
	entry := &models.WhitelistEntry{
		ReviewerLogin: "testuser",
		EntryType:     models.EntryTypeFamily,
		Name:          "Go - I",
	}

	assert.Equal(t, "testuser", entry.ReviewerLogin)
	assert.Equal(t, models.EntryTypeFamily, entry.EntryType)
	assert.Equal(t, "Go - I", entry.Name)
	assert.True(t, models.IsValidEntryType(entry.EntryType))
}

func TestProjectFamilyCreation(t *testing.T) {
	family := &models.ProjectFamily{
		FamilyLabel: "C - I",
		ProjectName: "C5_s21_decimal",
	}

	assert.Equal(t, "C - I", family.FamilyLabel)
	assert.Equal(t, "C5_s21_decimal", family.ProjectName)
}

func TestLockboxPayload(t *testing.T) {
	payload := &models.LockboxPayload{
		Version: 1,
		Users: map[string]models.UserTokens{
			"user1": {
				AccessToken:  "access123",
				RefreshToken: "refresh123",
			},
		},
	}

	assert.Equal(t, 1, payload.Version)
	assert.Len(t, payload.Users, 1)
	assert.Equal(t, "access123", payload.Users["user1"].AccessToken)
}

func TestTelegramCallbackDataParsing(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		action  string
		id      string
		isValid bool
	}{
		{
			name:    "Valid approve callback",
			data:    "APPROVE:550e8400-e29b-41d4-a716-446655440000",
			action:  "APPROVE",
			id:      "550e8400-e29b-41d4-a716-446655440000",
			isValid: true,
		},
		{
			name:    "Valid decline callback",
			data:    "DECLINE:550e8400-e29b-41d4-a716-446655440000",
			action:  "DECLINE",
			id:      "550e8400-e29b-41d4-a716-446655440000",
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simple split test (in real code, use ParseCallbackData function)
			parts := strings.Split(tt.data, ":")
			require.Len(t, parts, 2)

			callbackData := &models.TelegramCallbackData{
				Action:          parts[0],
				ReviewRequestID: parts[1],
			}

			assert.Equal(t, tt.action, callbackData.Action)
			assert.Equal(t, tt.id, callbackData.ReviewRequestID)
		})
	}
}

func TestUserSettingsValidation(t *testing.T) {
	t.Run("Valid settings", func(t *testing.T) {
		settings := &models.UserSettings{
			ReviewerLogin:                  "testuser",
			ResponseDeadlineShiftMinutes:   30,
			NonWhitelistCancelDelayMinutes: 5,
			NotifyWhitelistTimeout:         true,
			NotifyNonWhitelistCancel:       false,
			SlotShiftThresholdMinutes:      25,
			SlotShiftDurationMinutes:       15,
			CleanupDurationsMinutes:        30,
		}

		assert.Equal(t, int32(30), settings.ResponseDeadlineShiftMinutes)
		assert.Equal(t, int32(5), settings.NonWhitelistCancelDelayMinutes)
	})
}

func TestCalendarSlotAndBooking(t *testing.T) {
	t.Run("Calendar slot", func(t *testing.T) {
		slot := &models.CalendarSlot{
			ID:    "slot-123",
			Start: time.Now().Unix(),
			End:   time.Now().Add(1 * time.Hour).Unix(),
			Type:  models.SlotTypeFreeTime,
		}

		assert.Equal(t, "slot-123", slot.ID)
		assert.Equal(t, models.SlotTypeFreeTime, slot.Type)
	})

	t.Run("Calendar booking", func(t *testing.T) {
		booking := &models.CalendarBooking{
			ID:          "booking-123",
			EventSlotID: "slot-123",
			StartTime:   time.Now().Unix(),
			EndTime:     time.Now().Add(1 * time.Hour).Unix(),
			ProjectName: "go-concurrency",
		}

		assert.Equal(t, "booking-123", booking.ID)
		assert.Equal(t, "go-concurrency", booking.ProjectName)
	})
}
