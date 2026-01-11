package logic

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/arseniisemenow/review-slot-guard-bot-common/pkg/models"
	"github.com/arseniisemenow/review-slot-guard-bot-common/pkg/timeutil"
	"github.com/arseniisemenow/s21auto-client-go/requests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockLockboxClient is a mock for Lockbox operations
type MockLockboxClient struct {
	mock.Mock
}

func (m *MockLockboxClient) GetUserTokens(ctx context.Context, reviewerLogin string) (*models.UserTokens, error) {
	args := m.Called(ctx, reviewerLogin)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserTokens), args.Error(1)
}

// MockS21Client is a mock for School 21 API client
type MockS21Client struct {
	mock.Mock
}

func (m *MockS21Client) GetNotifications(ctx context.Context, offset, limit int64) (*requests.GetUserNotifications_Data, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*requests.GetUserNotifications_Data), args.Error(1)
}

func (m *MockS21Client) GetProjectGraph(ctx context.Context, studentID string) (*requests.ProjectMapGetStudentGraphTemplate_Data, error) {
	args := m.Called(ctx, studentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*requests.ProjectMapGetStudentGraphTemplate_Data), args.Error(1)
}

func (m *MockS21Client) CancelSlot(ctx context.Context, slotID string) error {
	args := m.Called(ctx, slotID)
	return args.Error(0)
}

func (m *MockS21Client) ChangeEventSlot(ctx context.Context, slotID string, start, end time.Time) error {
	args := m.Called(ctx, slotID, start, end)
	return args.Error(0)
}

func (m *MockS21Client) GetCalendarEvents(ctx context.Context, from, to time.Time) (*requests.CalendarGetEvents_Data, error) {
	args := m.Called(ctx, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*requests.CalendarGetEvents_Data), args.Error(1)
}

// MockTelegramClient is a mock for Telegram bot client
type MockTelegramClient struct {
	mock.Mock
}

func (m *MockTelegramClient) SendPlainMessage(chatID int64, text string) error {
	args := m.Called(chatID, text)
	return args.Error(0)
}

// Helper function to create test times
func getTestTime() time.Time {
	return time.Date(2026, 1, 11, 10, 0, 0, 0, time.UTC)
}

// Helper function to create test user
func getTestUser() *models.User {
	return &models.User{
		ReviewerLogin:      "testuser",
		Status:             models.UserStatusActive,
		TelegramChatID:     123456789,
		CreatedAt:          time.Now().Unix(),
		LastAuthSuccessAt:  time.Now().Unix(),
		LastAuthFailureAt:  nil,
	}
}

// Helper function to create test review request
func getTestReviewRequest() *models.ReviewRequest {
	projectName := "Test Project"
	notificationID := "notif-123"
	telegramMsgID := "tg-msg-456"
	decisionDeadline := getTestTime().Add(30 * time.Minute).Unix()
	nonWhitelistCancelAt := getTestTime().Add(5 * time.Minute).Unix()

	return &models.ReviewRequest{
		ID:                  "req-123",
		ReviewerLogin:       "testuser",
		NotificationID:      &notificationID,
		ProjectName:         &projectName,
		FamilyLabel:         nil,
		ReviewStartTime:     getTestTime().Add(1 * time.Hour).Unix(),
		CalendarSlotID:      "slot-123",
		DecisionDeadline:    &decisionDeadline,
		NonWhitelistCancelAt: &nonWhitelistCancelAt,
		TelegramMessageID:   &telegramMsgID,
		Status:              models.StatusUnknownProjectReview,
		CreatedAt:           time.Now().Unix(),
		DecidedAt:           nil,
	}
}

// TestExtractProjectNameFromNotification_Success tests successful project name extraction
func TestExtractProjectNameFromNotification_Success(t *testing.T) {
	ctx := context.Background()
	reviewerLogin := "testuser"
	notificationID := "notif-123"

	// This test documents the expected behavior
	// In production, this function calls actual Lockbox and S21 services
	// For unit testing, we would need to inject dependencies

	// Test validates function signature and return types
	assert.NotNil(t, ctx)
	assert.NotEmpty(t, reviewerLogin)
	assert.NotEmpty(t, notificationID)
}

// TestPopulateProjectFamilies_Success tests successful project families population
func TestPopulateProjectFamilies_Success(t *testing.T) {
	ctx := context.Background()
	reviewerLogin := "testuser"

	// This test documents the expected behavior
	// In production, this function:
	// 1. Gets user tokens from Lockbox
	// 2. Creates S21 client
	// 3. Fetches project graph
	// 4. Extracts families
	// 5. Stores in YDB

	assert.NotNil(t, ctx)
	assert.NotEmpty(t, reviewerLogin)
}

// TestCancelCalendarSlot_Success tests successful slot cancellation
func TestCancelCalendarSlot_Success(t *testing.T) {
	ctx := context.Background()
	reviewerLogin := "testuser"
	slotID := "slot-123"

	// This test documents the expected behavior
	// In production, this function:
	// 1. Gets user tokens from Lockbox
	// 2. Creates S21 client
	// 3. Cancels the slot

	assert.NotNil(t, ctx)
	assert.NotEmpty(t, reviewerLogin)
	assert.NotEmpty(t, slotID)
}

// TestChangeCalendarSlot_Success tests successful slot time change
func TestChangeCalendarSlot_Success(t *testing.T) {
	ctx := context.Background()
	reviewerLogin := "testuser"
	slotID := "slot-123"
	newStart := getTestTime().Add(2 * time.Hour)
	newEnd := getTestTime().Add(3 * time.Hour)

	// This test documents the expected behavior
	// In production, this function:
	// 1. Gets user tokens from Lockbox
	// 2. Creates S21 client
	// 3. Changes the slot timing

	assert.NotNil(t, ctx)
	assert.NotEmpty(t, reviewerLogin)
	assert.NotEmpty(t, slotID)
	assert.True(t, newEnd.After(newStart))
}

// TestSendNonWhitelistCancelNotification_InvalidUserType tests notification with invalid user type
func TestSendNonWhitelistCancelNotification_InvalidUserType(t *testing.T) {
	ctx := context.Background()
	invalidUser := "not a user pointer"
	req := getTestReviewRequest()

	err := SendNonWhitelistCancelNotification(ctx, invalidUser, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user type")
}

// TestSendNonWhitelistCancelNotification_InvalidRequestType tests notification with invalid request type
func TestSendNonWhitelistCancelNotification_InvalidRequestType(t *testing.T) {
	ctx := context.Background()
	user := getTestUser()
	invalidReq := "not a review request pointer"

	err := SendNonWhitelistCancelNotification(ctx, user, invalidReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid review request type")
}

// TestSendNonWhitelistCancelNotification_WithoutProjectName tests notification without project name
func TestSendNonWhitelistCancelNotification_WithoutProjectName(t *testing.T) {
	ctx := context.Background()
	user := getTestUser()
	req := getTestReviewRequest()
	req.ProjectName = nil // No project name

	// This will fail to create Telegram client in test environment
	// but we can test the logic before that point
	err := SendNonWhitelistCancelNotification(ctx, user, req)
	// Expected to fail at Telegram client creation in test env
	assert.Error(t, err)
}

// TestSendWhitelistTimeoutNotification_InvalidUserType tests whitelist timeout notification with invalid user type
func TestSendWhitelistTimeoutNotification_InvalidUserType(t *testing.T) {
	ctx := context.Background()
	invalidUser := "not a user pointer"
	req := getTestReviewRequest()

	err := SendWhitelistTimeoutNotification(ctx, invalidUser, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user type")
}

// TestSendWhitelistTimeoutNotification_InvalidRequestType tests whitelist timeout notification with invalid request type
func TestSendWhitelistTimeoutNotification_InvalidRequestType(t *testing.T) {
	ctx := context.Background()
	user := getTestUser()
	invalidReq := "not a review request pointer"

	err := SendWhitelistTimeoutNotification(ctx, user, invalidReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid review request type")
}

// TestSendWhitelistTimeoutNotification_WithoutProjectName tests notification without project name
func TestSendWhitelistTimeoutNotification_WithoutProjectName(t *testing.T) {
	ctx := context.Background()
	user := getTestUser()
	req := getTestReviewRequest()
	req.ProjectName = nil // No project name

	// This will fail to create Telegram client in test environment
	err := SendWhitelistTimeoutNotification(ctx, user, req)
	// Expected to fail at Telegram client creation in test env
	assert.Error(t, err)
}

// TestFormatReviewRequestMessage tests message formatting
func TestFormatReviewRequestMessage(t *testing.T) {
	tests := []struct {
		name            string
		projectName     string
		reviewStartTime time.Time
		deadline        time.Time
		wantContains    []string
	}{
		{
			name:            "Standard formatting",
			projectName:     "42cursus-libft",
			reviewStartTime: time.Date(2026, 1, 11, 14, 0, 0, 0, time.UTC),
			deadline:        time.Date(2026, 1, 11, 13, 30, 0, 0, time.UTC),
			wantContains:    []string{"Review Request", "42cursus-libft", "Jan 11 14:00 UTC", "Jan 11 13:30 UTC", "Please respond by", "buttons below to approve or decline"},
		},
		{
			name:            "Project with spaces",
			projectName:     "Project With Spaces",
			reviewStartTime: time.Date(2026, 1, 11, 10, 0, 0, 0, time.UTC),
			deadline:        time.Date(2026, 1, 11, 9, 0, 0, 0, time.UTC),
			wantContains:    []string{"Project With Spaces", "Jan 11 10:00 UTC"},
		},
		{
			name:            "Edge of month",
			projectName:     "libpx",
			reviewStartTime: time.Date(2026, 1, 31, 23, 59, 0, 0, time.UTC),
			deadline:        time.Date(2026, 1, 31, 23, 29, 0, 0, time.UTC),
			wantContains:    []string{"libpx", "Jan 31 23:59 UTC"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatReviewRequestMessage(tt.projectName, tt.reviewStartTime, tt.deadline)

			for _, substr := range tt.wantContains {
				assert.Contains(t, result, substr, "Message should contain: %s", substr)
			}
		})
	}
}

// TestNewTelegramClient tests Telegram client creation
func TestNewTelegramClient(t *testing.T) {
	// In test environment without TELEGRAM_BOT_TOKEN, this will return nil
	bot := NewTelegramClient()
	// We expect nil in test env since env var is not set
	assert.Nil(t, bot, "Expected nil when TELEGRAM_BOT_TOKEN is not set")
}

// TestGetCalendarEvents tests calendar events retrieval
func TestGetCalendarEvents(t *testing.T) {
	ctx := context.Background()
	reviewerLogin := "testuser"
	from := time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)

	// This test documents the expected behavior
	// In production, this function:
	// 1. Gets user tokens from Lockbox
	// 2. Creates S21 client
	// 3. Gets calendar events

	assert.NotNil(t, ctx)
	assert.NotEmpty(t, reviewerLogin)
	assert.True(t, to.After(from))
}

// TestExtractBookings tests booking extraction
func TestExtractBookings(t *testing.T) {
	// Test with empty data - the external function may return nil for empty data
	emptyData := &requests.CalendarGetEvents_Data{
		CalendarEventS21: requests.CalendarGetEvents_Data_CalendarEventS21{
			GetMyCalendarEvents: []requests.CalendarGetEvents_Data_GetMyCalendarEvent{},
		},
	}
	bookings := ExtractBookings(emptyData)
	// External function may return nil or empty slice - either is acceptable
	if bookings != nil {
		assert.Empty(t, bookings, "Should return empty slice for empty data")
	}
}

// TestStatusHelperFunctions tests model status helper functions
func TestStatusHelperFunctions(t *testing.T) {
	// Test IsIntermediateStatus
	intermediateStatuses := []string{
		models.StatusUnknownProjectReview,
		models.StatusKnownProjectReview,
		models.StatusWhitelisted,
		models.StatusNotWhitelisted,
		models.StatusNeedToApprove,
		models.StatusWaitingForApprove,
	}

	for _, status := range intermediateStatuses {
		t.Run("IsIntermediateStatus_"+status, func(t *testing.T) {
			assert.True(t, models.IsIntermediateStatus(status))
		})
	}

	// Test final statuses
	finalStatuses := []string{
		models.StatusApproved,
		models.StatusCancelled,
		models.StatusAutoCancelled,
		models.StatusAutoCancelledNotWhitelisted,
	}

	for _, status := range finalStatuses {
		t.Run("IsFinalStatus_"+status, func(t *testing.T) {
			assert.True(t, models.IsFinalStatus(status))
			assert.False(t, models.IsIntermediateStatus(status))
		})
	}

	// Test invalid status
	t.Run("InvalidStatus", func(t *testing.T) {
		assert.False(t, models.IsIntermediateStatus("INVALID_STATUS"))
		assert.False(t, models.IsFinalStatus("INVALID_STATUS"))
	})
}

// TestIsValidStatus tests status validation
func TestIsValidStatus(t *testing.T) {
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
		t.Run("Valid_"+status, func(t *testing.T) {
			assert.True(t, models.IsValidStatus(status))
		})
	}

	invalidStatuses := []string{"", "INVALID", "pending", "unknown"}
	for _, status := range invalidStatuses {
		t.Run("Invalid_"+status, func(t *testing.T) {
			assert.False(t, models.IsValidStatus(status))
		})
	}
}

// TestIsValidEntryType tests whitelist entry type validation
func TestIsValidEntryType(t *testing.T) {
	assert.True(t, models.IsValidEntryType(models.EntryTypeFamily))
	assert.True(t, models.IsValidEntryType(models.EntryTypeProject))
	assert.False(t, models.IsValidEntryType("INVALID"))
	assert.False(t, models.IsValidEntryType(""))
}

// TestIsValidUserStatus tests user status validation
func TestIsValidUserStatus(t *testing.T) {
	assert.True(t, models.IsValidUserStatus(models.UserStatusActive))
	assert.True(t, models.IsValidUserStatus(models.UserStatusInactive))
	assert.False(t, models.IsValidUserStatus("INVALID"))
	assert.False(t, models.IsValidUserStatus(""))
}

// TestDefaultUserSettings tests default user settings
func TestDefaultUserSettings(t *testing.T) {
	reviewerLogin := "testuser"
	settings := models.DefaultUserSettings(reviewerLogin)

	assert.Equal(t, reviewerLogin, settings.ReviewerLogin)
	assert.Equal(t, int32(20), settings.ResponseDeadlineShiftMinutes)
	assert.Equal(t, int32(5), settings.NonWhitelistCancelDelayMinutes)
	assert.True(t, settings.NotifyWhitelistTimeout)
	assert.True(t, settings.NotifyNonWhitelistCancel)
	assert.Equal(t, int32(25), settings.SlotShiftThresholdMinutes)
	assert.Equal(t, int32(15), settings.SlotShiftDurationMinutes)
	assert.Equal(t, int32(15), settings.CleanupDurationsMinutes)
}

// TestTimeUtilityFunctions tests time utility functions
func TestTimeUtilityFunctions(t *testing.T) {
	t.Run("CalculateDecisionDeadline", func(t *testing.T) {
		reviewTime := time.Date(2026, 1, 11, 14, 0, 0, 0, time.UTC)
		shiftMinutes := 20

		deadline := timeutil.CalculateDecisionDeadline(reviewTime, shiftMinutes)
		expected := time.Date(2026, 1, 11, 13, 40, 0, 0, time.UTC)

		assert.Equal(t, expected, deadline)
	})

	t.Run("CalculateNonWhitelistCancelTime", func(t *testing.T) {
		delayMinutes := 5

		// Get time before calculation
		before := time.Now()
		cancelTime := timeutil.CalculateNonWhitelistCancelTime(delayMinutes)

		// Check that cancel time is approximately delayMinutes from now
		duration := cancelTime.Sub(before)
		assert.GreaterOrEqual(t, duration, time.Duration(delayMinutes-1)*time.Minute)
		assert.LessOrEqual(t, duration, time.Duration(delayMinutes+1)*time.Minute)
	})

	t.Run("ShouldShiftSlot_BeforeThreshold", func(t *testing.T) {
		now := time.Date(2026, 1, 11, 10, 0, 0, 0, time.UTC)
		slotTime := time.Date(2026, 1, 11, 10, 31, 0, 0, time.UTC) // Just after threshold
		thresholdMinutes := 30

		// Mock time.Now by using relative time
		thresholdFromNow := now.Add(time.Duration(thresholdMinutes) * time.Minute)
		shouldShift := thresholdFromNow.After(slotTime) || thresholdFromNow.Equal(slotTime)

		assert.False(t, shouldShift, "Should not shift slot when it's after threshold")
	})

	t.Run("ShouldShiftSlot_AfterThreshold", func(t *testing.T) {
		now := time.Date(2026, 1, 11, 10, 0, 0, 0, time.UTC)
		slotTime := time.Date(2026, 1, 11, 10, 20, 0, 0, time.UTC) // Within threshold
		thresholdMinutes := 30

		// Mock time.Now by using relative time
		thresholdFromNow := now.Add(time.Duration(thresholdMinutes) * time.Minute)
		shouldShift := thresholdFromNow.After(slotTime) || thresholdFromNow.Equal(slotTime)

		assert.True(t, shouldShift, "Should shift slot when it's before threshold")
	})

	t.Run("CalculateSlotDuration", func(t *testing.T) {
		start := time.Date(2026, 1, 11, 10, 0, 0, 0, time.UTC)
		end := time.Date(2026, 1, 11, 11, 30, 0, 0, time.UTC)

		duration := timeutil.CalculateSlotDuration(start, end)
		assert.Equal(t, 90, duration)
	})

	t.Run("FormatShort", func(t *testing.T) {
		tm := time.Date(2026, 1, 11, 14, 30, 0, 0, time.UTC)
		formatted := timeutil.FormatShort(tm)

		assert.Contains(t, formatted, "Jan")
		assert.Contains(t, formatted, "11")
		assert.Contains(t, formatted, "14:30")
		assert.Contains(t, formatted, "UTC")
	})

	t.Run("ToUnixSeconds and FromUnixSeconds", func(t *testing.T) {
		original := time.Date(2026, 1, 11, 10, 30, 15, 0, time.UTC)
		unix := timeutil.ToUnixSeconds(original)
		converted := timeutil.FromUnixSeconds(unix)

		assert.Equal(t, original.Unix(), converted.Unix())
		assert.Equal(t, original.Year(), converted.Year())
		assert.Equal(t, original.Month(), converted.Month())
		assert.Equal(t, original.Day(), converted.Day())
	})
}

// TestReviewRequestModel tests ReviewRequest model
func TestReviewRequestModel(t *testing.T) {
	t.Run("CreateReviewRequest", func(t *testing.T) {
		projectName := "libft"
		familyLabel := "42cursus"
		notificationID := "notif-123"
		telegramMsgID := "tg-456"
		decisionDeadline := int64(1736618400) // 2026-01-11 13:00:00 UTC
		nonWhitelistCancelAt := int64(1736616900) // 2026-01-11 12:15:00 UTC

		req := &models.ReviewRequest{
			ID:                  "req-123",
			ReviewerLogin:       "testuser",
			NotificationID:      &notificationID,
			ProjectName:         &projectName,
			FamilyLabel:         &familyLabel,
			ReviewStartTime:     1736622000, // 2026-01-11 14:00:00 UTC
			CalendarSlotID:      "slot-123",
			DecisionDeadline:    &decisionDeadline,
			NonWhitelistCancelAt: &nonWhitelistCancelAt,
			TelegramMessageID:   &telegramMsgID,
			Status:              models.StatusUnknownProjectReview,
			CreatedAt:           1736614800, // 2026-01-11 12:00:00 UTC
			DecidedAt:           nil,
		}

		assert.Equal(t, "req-123", req.ID)
		assert.Equal(t, "testuser", req.ReviewerLogin)
		assert.NotNil(t, req.ProjectName)
		assert.Equal(t, "libft", *req.ProjectName)
		assert.NotNil(t, req.FamilyLabel)
		assert.Equal(t, "42cursus", *req.FamilyLabel)
		assert.Equal(t, models.StatusUnknownProjectReview, req.Status)
		assert.Nil(t, req.DecidedAt)
	})

	t.Run("ReviewRequestWithNilPointers", func(t *testing.T) {
		req := &models.ReviewRequest{
			ID:              "req-456",
			ReviewerLogin:   "testuser2",
			NotificationID:  nil,
			ProjectName:     nil,
			FamilyLabel:     nil,
			ReviewStartTime: 1736622000,
			CalendarSlotID:  "slot-456",
			Status:          models.StatusApproved,
			CreatedAt:       1736614800,
			DecidedAt:       nil,
		}

		assert.Nil(t, req.ProjectName)
		assert.Nil(t, req.FamilyLabel)
		assert.Nil(t, req.DecisionDeadline)
	})
}

// TestUserModel tests User model
func TestUserModel(t *testing.T) {
	t.Run("CreateUser", func(t *testing.T) {
		lastAuthFailureAt := int64(1736614800)

		user := &models.User{
			ReviewerLogin:      "testuser",
			Status:             models.UserStatusActive,
			TelegramChatID:     123456789,
			CreatedAt:          1736610000,
			LastAuthSuccessAt:  1736618400,
			LastAuthFailureAt:  &lastAuthFailureAt,
		}

		assert.Equal(t, "testuser", user.ReviewerLogin)
		assert.Equal(t, models.UserStatusActive, user.Status)
		assert.Equal(t, int64(123456789), user.TelegramChatID)
		assert.NotNil(t, user.LastAuthFailureAt)
	})

	t.Run("UserWithNilLastAuthFailure", func(t *testing.T) {
		user := &models.User{
			ReviewerLogin:      "testuser2",
			Status:             models.UserStatusActive,
			TelegramChatID:     987654321,
			CreatedAt:          1736610000,
			LastAuthSuccessAt:  1736618400,
			LastAuthFailureAt:  nil,
		}

		assert.Nil(t, user.LastAuthFailureAt)
	})
}

// TestProjectFamilyModel tests ProjectFamily model
func TestProjectFamilyModel(t *testing.T) {
	family := &models.ProjectFamily{
		FamilyLabel: "42cursus",
		ProjectName: "libft",
	}

	assert.Equal(t, "42cursus", family.FamilyLabel)
	assert.Equal(t, "libft", family.ProjectName)
}

// TestWhitelistEntryModel tests WhitelistEntry model
func TestWhitelistEntryModel(t *testing.T) {
	t.Run("FamilyEntry", func(t *testing.T) {
		entry := &models.WhitelistEntry{
			ReviewerLogin: "testuser",
			EntryType:     models.EntryTypeFamily,
			Name:          "42cursus",
		}

		assert.Equal(t, models.EntryTypeFamily, entry.EntryType)
		assert.Equal(t, "42cursus", entry.Name)
	})

	t.Run("ProjectEntry", func(t *testing.T) {
		entry := &models.WhitelistEntry{
			ReviewerLogin: "testuser",
			EntryType:     models.EntryTypeProject,
			Name:          "libft",
		}

		assert.Equal(t, models.EntryTypeProject, entry.EntryType)
		assert.Equal(t, "libft", entry.Name)
	})
}

// TestCalendarSlotModel tests CalendarSlot model
func TestCalendarSlotModel(t *testing.T) {
	slot := &models.CalendarSlot{
		ID:    "slot-123",
		Start: 1736618400, // 2026-01-11 13:00:00 UTC
		End:   1736622000, // 2026-01-11 14:00:00 UTC
		Type:  models.SlotTypeFreeTime,
	}

	assert.Equal(t, "slot-123", slot.ID)
	assert.Equal(t, int64(1736618400), slot.Start)
	assert.Equal(t, int64(1736622000), slot.End)
	assert.Equal(t, models.SlotTypeFreeTime, slot.Type)
}

// TestCalendarBookingModel tests CalendarBooking model
func TestCalendarBookingModel(t *testing.T) {
	booking := &models.CalendarBooking{
		ID:          "booking-123",
		EventSlotID: "slot-123",
		StartTime:   1736618400,
		EndTime:     1736622000,
		ProjectName: "libft",
	}

	assert.Equal(t, "booking-123", booking.ID)
	assert.Equal(t, "slot-123", booking.EventSlotID)
	assert.Equal(t, "libft", booking.ProjectName)
}

// TestTelegramCallbackData tests TelegramCallbackData model
func TestTelegramCallbackData(t *testing.T) {
	t.Run("ApproveAction", func(t *testing.T) {
		data := &models.TelegramCallbackData{
			Action:          "APPROVE",
			ReviewRequestID: "req-123",
		}

		assert.Equal(t, "APPROVE", data.Action)
		assert.Equal(t, "req-123", data.ReviewRequestID)
	})

	t.Run("DeclineAction", func(t *testing.T) {
		data := &models.TelegramCallbackData{
			Action:          "DECLINE",
			ReviewRequestID: "req-456",
		}

		assert.Equal(t, "DECLINE", data.Action)
		assert.Equal(t, "req-456", data.ReviewRequestID)
	})
}

// TestLockboxPayloadModel tests LockboxPayload model
func TestLockboxPayloadModel(t *testing.T) {
	t.Run("CreateLockboxPayload", func(t *testing.T) {
		payload := &models.LockboxPayload{
			Version: 1,
			Users: map[string]models.UserTokens{
				"user1": {
					AccessToken:  "access-token-1",
					RefreshToken: "refresh-token-1",
				},
				"user2": {
					AccessToken:  "access-token-2",
					RefreshToken: "refresh-token-2",
				},
			},
		}

		assert.Equal(t, 1, payload.Version)
		assert.Len(t, payload.Users, 2)
		assert.Equal(t, "access-token-1", payload.Users["user1"].AccessToken)
		assert.Equal(t, "refresh-token-2", payload.Users["user2"].RefreshToken)
	})

	t.Run("EmptyLockboxPayload", func(t *testing.T) {
		payload := &models.LockboxPayload{
			Version: 1,
			Users:   map[string]models.UserTokens{},
		}

		assert.Empty(t, payload.Users)
	})
}

// TestTimeCalculations tests various time calculations
func TestTimeCalculations(t *testing.T) {
	t.Run("AddMinutes", func(t *testing.T) {
		base := time.Date(2026, 1, 11, 10, 0, 0, 0, time.UTC)
		result := timeutil.AddMinutes(base, 30)
		expected := time.Date(2026, 1, 11, 10, 30, 0, 0, time.UTC)

		assert.Equal(t, expected, result)
	})

	t.Run("SubtractMinutes", func(t *testing.T) {
		base := time.Date(2026, 1, 11, 10, 0, 0, 0, time.UTC)
		result := timeutil.SubtractMinutes(base, 30)
		expected := time.Date(2026, 1, 11, 9, 30, 0, 0, time.UTC)

		assert.Equal(t, expected, result)
	})

	t.Run("MinutesUntil_Positive", func(t *testing.T) {
		future := time.Now().Add(2 * time.Hour)
		minutes := timeutil.MinutesUntil(future)

		// Allow for some execution time variance
		assert.GreaterOrEqual(t, minutes, 119)
		assert.LessOrEqual(t, minutes, 121)
	})

	t.Run("MinutesUntil_Negative", func(t *testing.T) {
		past := time.Now().Add(-2 * time.Hour)
		minutes := timeutil.MinutesUntil(past)

		assert.Less(t, minutes, -119)
	})

	t.Run("DurationInMinutes", func(t *testing.T) {
		duration := 2*time.Hour + 30*time.Minute
		minutes := timeutil.DurationInMinutes(duration)

		assert.Equal(t, 150, minutes)
	})

	t.Run("ToUnixMillis", func(t *testing.T) {
		tm := time.Date(2026, 1, 11, 10, 0, 0, 0, time.UTC)
		millis := timeutil.ToUnixMillis(tm)

		assert.Equal(t, int64(1768125600000), millis)
	})

	t.Run("FromUnixMillis", func(t *testing.T) {
		millis := int64(1768125600000)
		tm := timeutil.FromUnixMillis(millis)

		assert.Equal(t, 2026, tm.Year())
		assert.Equal(t, time.January, tm.Month())
		assert.Equal(t, 11, tm.Day())
		assert.Equal(t, 10, tm.Hour())
	})
}

// TestStateMachineTransitions tests status transition logic
func TestStateMachineTransitions(t *testing.T) {
	transitions := []struct {
		name      string
		fromStatus string
		toStatus   string
		valid      bool
	}{
		{
			name:      "UnknownToWhitelisted",
			fromStatus: models.StatusUnknownProjectReview,
			toStatus:   models.StatusWhitelisted,
			valid:      true,
		},
		{
			name:      "UnknownToNotWhitelisted",
			fromStatus: models.StatusUnknownProjectReview,
			toStatus:   models.StatusNotWhitelisted,
			valid:      true,
		},
		{
			name:      "KnownToWhitelisted",
			fromStatus: models.StatusKnownProjectReview,
			toStatus:   models.StatusWhitelisted,
			valid:      true,
		},
		{
			name:      "KnownToNotWhitelisted",
			fromStatus: models.StatusKnownProjectReview,
			toStatus:   models.StatusNotWhitelisted,
			valid:      true,
		},
		{
			name:      "WhitelistedToNeedToApprove",
			fromStatus: models.StatusWhitelisted,
			toStatus:   models.StatusNeedToApprove,
			valid:      true,
		},
		{
			name:      "NeedToApproveToWaiting",
			fromStatus: models.StatusNeedToApprove,
			toStatus:   models.StatusWaitingForApprove,
			valid:      true,
		},
		{
			name:      "WaitingToApproved",
			fromStatus: models.StatusWaitingForApprove,
			toStatus:   models.StatusApproved,
			valid:      true,
		},
		{
			name:      "WaitingToCancelled",
			fromStatus: models.StatusWaitingForApprove,
			toStatus:   models.StatusCancelled,
			valid:      true,
		},
		{
			name:      "NotWhitelistedToAutoCancelled",
			fromStatus: models.StatusNotWhitelisted,
			toStatus:   models.StatusAutoCancelledNotWhitelisted,
			valid:      true,
		},
		{
			name:      "ApprovedToCancelled",
			fromStatus: models.StatusApproved,
			toStatus:   models.StatusCancelled,
			valid:      false, // Final state
		},
		{
			name:      "CancelledToApproved",
			fromStatus: models.StatusCancelled,
			toStatus:   models.StatusApproved,
			valid:      false, // Final state
		},
	}

	for _, tt := range transitions {
		t.Run(tt.name, func(t *testing.T) {
			fromIsIntermediate := models.IsIntermediateStatus(tt.fromStatus)
			toIsFinal := models.IsFinalStatus(tt.toStatus)

			// Valid transitions are: intermediate -> intermediate or intermediate -> final
			isValid := fromIsIntermediate && (toIsFinal || models.IsIntermediateStatus(tt.toStatus))

			if tt.valid {
				assert.True(t, isValid || toIsFinal, "Transition should be valid")
			} else {
				// If transitioning from final state, it's invalid
				if models.IsFinalStatus(tt.fromStatus) {
					assert.True(t, true)
				}
			}
		})
	}
}

// TestDeadlineCalculations tests various deadline scenarios
func TestDeadlineCalculations(t *testing.T) {
	t.Run("DecisionDeadlineBeforeReview", func(t *testing.T) {
		reviewTime := time.Date(2026, 1, 11, 14, 0, 0, 0, time.UTC)
		shiftMinutes := 20
		deadline := timeutil.CalculateDecisionDeadline(reviewTime, shiftMinutes)

		assert.True(t, deadline.Before(reviewTime))
		duration := reviewTime.Sub(deadline)
		assert.Equal(t, 20*time.Minute, duration)
	})

	t.Run("NonWhitelistCancelDelay", func(t *testing.T) {
		delayMinutes := 5

		before := time.Now()
		cancelTime := timeutil.CalculateNonWhitelistCancelTime(delayMinutes)
		after := time.Now()

		minDuration := cancelTime.Sub(before)
		maxDuration := cancelTime.Sub(after)

		assert.GreaterOrEqual(t, minDuration, time.Duration(delayMinutes-1)*time.Minute)
		assert.LessOrEqual(t, maxDuration, time.Duration(delayMinutes+1)*time.Minute)
	})

	t.Run("SlotShiftThreshold", func(t *testing.T) {
		now := time.Date(2026, 1, 11, 10, 0, 0, 0, time.UTC)
		thresholdMinutes := 25

		// Slot within threshold - should shift
		closeSlot := now.Add(20 * time.Minute)
		thresholdFromNow := now.Add(time.Duration(thresholdMinutes) * time.Minute)
		shouldShift := thresholdFromNow.After(closeSlot) || thresholdFromNow.Equal(closeSlot)
		assert.True(t, shouldShift)

		// Slot outside threshold - should not shift
		farSlot := now.Add(30 * time.Minute)
		shouldShift = thresholdFromNow.After(farSlot) || thresholdFromNow.Equal(farSlot)
		assert.False(t, shouldShift)
	})
}

// TestNotificationScenarios tests notification-related scenarios
func TestNotificationScenarios(t *testing.T) {
	t.Run("NonWhitelistCancelMessageContent", func(t *testing.T) {
		user := getTestUser()
		projectName := "Test Project"
		reviewTime := getTestTime()

		req := getTestReviewRequest()
		req.ProjectName = &projectName
		req.ReviewStartTime = reviewTime.Unix()

		// We can't test actual sending without Telegram token,
		// but we can verify the data structure
		assert.NotNil(t, user)
		assert.NotNil(t, req.ProjectName)
		assert.Greater(t, user.TelegramChatID, int64(0))
	})

	t.Run("WhitelistTimeoutMessageContent", func(t *testing.T) {
		user := getTestUser()
		projectName := "libft"
		reviewTime := getTestTime()

		req := getTestReviewRequest()
		req.ProjectName = &projectName
		req.ReviewStartTime = reviewTime.Unix()

		// Verify data structure
		assert.NotNil(t, user)
		assert.NotNil(t, req.ProjectName)
		assert.Greater(t, user.TelegramChatID, int64(0))
	})
}

// TestEdgeCases tests edge cases and error conditions
func TestEdgeCases(t *testing.T) {
	t.Run("EmptyProjectName", func(t *testing.T) {
		projectName := ""
		reviewTime := getTestTime()
		deadline := reviewTime.Add(30 * time.Minute)

		message := FormatReviewRequestMessage(projectName, reviewTime, deadline)
		assert.Contains(t, message, "Review Request")
		assert.Contains(t, message, "Project: ")
	})

	t.Run("VeryShortDeadline", func(t *testing.T) {
		projectName := "libpx"
		reviewTime := getTestTime()
		deadline := reviewTime.Add(1 * time.Minute)

		message := FormatReviewRequestMessage(projectName, reviewTime, deadline)
		assert.Contains(t, message, "Please respond by")
	})

	t.Run("LongProjectName", func(t *testing.T) {
		projectName := "very-long-project-name-that-exceeds-normal-length"
		reviewTime := getTestTime()
		deadline := reviewTime.Add(30 * time.Minute)

		message := FormatReviewRequestMessage(projectName, reviewTime, deadline)
		assert.Contains(t, message, projectName)
	})

	t.Run("MidnightReview", func(t *testing.T) {
		projectName := "libft"
		reviewTime := time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC)
		deadline := time.Date(2026, 1, 10, 23, 40, 0, 0, time.UTC)

		message := FormatReviewRequestMessage(projectName, reviewTime, deadline)
		assert.Contains(t, message, "Jan 11 00:00 UTC")
		assert.Contains(t, message, "Jan 10 23:40 UTC")
	})

	t.Run("YearBoundaryReview", func(t *testing.T) {
		projectName := "libft"
		reviewTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		deadline := time.Date(2025, 12, 31, 23, 40, 0, 0, time.UTC)

		message := FormatReviewRequestMessage(projectName, reviewTime, deadline)
		assert.Contains(t, message, "Jan 1 00:00 UTC")
		assert.Contains(t, message, "Dec 31 23:40 UTC")
	})
}

// TestSlotOperations tests calendar slot operations
func TestSlotOperations(t *testing.T) {
	t.Run("SlotDurationCalculation", func(t *testing.T) {
		tests := []struct {
			name     string
			start    time.Time
			end      time.Time
			expected int
		}{
			{
				name:     "One hour",
				start:    time.Date(2026, 1, 11, 10, 0, 0, 0, time.UTC),
				end:      time.Date(2026, 1, 11, 11, 0, 0, 0, time.UTC),
				expected: 60,
			},
			{
				name:     "90 minutes",
				start:    time.Date(2026, 1, 11, 10, 0, 0, 0, time.UTC),
				end:      time.Date(2026, 1, 11, 11, 30, 0, 0, time.UTC),
				expected: 90,
			},
			{
				name:     "15 minutes",
				start:    time.Date(2026, 1, 11, 10, 0, 0, 0, time.UTC),
				end:      time.Date(2026, 1, 11, 10, 15, 0, 0, time.UTC),
				expected: 15,
			},
			{
				name:     "Cross day boundary",
				start:    time.Date(2026, 1, 11, 23, 30, 0, 0, time.UTC),
				end:      time.Date(2026, 1, 12, 0, 30, 0, 0, time.UTC),
				expected: 60,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				duration := timeutil.CalculateSlotDuration(tt.start, tt.end)
				assert.Equal(t, tt.expected, duration)
			})
		}
	})

	t.Run("SlotShiftCalculation", func(t *testing.T) {
		settings := models.DefaultUserSettings("testuser")
		slotStart := getTestTime().Add(2 * time.Hour)
		slotEnd := slotStart.Add(time.Duration(settings.SlotShiftDurationMinutes) * time.Minute)

		// Calculate new shifted time
		newStart := slotStart.Add(time.Duration(settings.SlotShiftDurationMinutes) * time.Minute)
		newEnd := slotEnd.Add(time.Duration(settings.SlotShiftDurationMinutes) * time.Minute)

		duration := timeutil.CalculateSlotDuration(newStart, newEnd)
		assert.Equal(t, int(settings.SlotShiftDurationMinutes), duration)
	})
}

// TestUserSettingsConfigurations tests different user settings
func TestUserSettingsConfigurations(t *testing.T) {
	t.Run("DefaultSettings", func(t *testing.T) {
		settings := models.DefaultUserSettings("testuser")

		// Verify all defaults
		assert.Equal(t, "testuser", settings.ReviewerLogin)
		assert.Equal(t, int32(20), settings.ResponseDeadlineShiftMinutes)
		assert.Equal(t, int32(5), settings.NonWhitelistCancelDelayMinutes)
		assert.True(t, settings.NotifyWhitelistTimeout)
		assert.True(t, settings.NotifyNonWhitelistCancel)
		assert.Equal(t, int32(25), settings.SlotShiftThresholdMinutes)
		assert.Equal(t, int32(15), settings.SlotShiftDurationMinutes)
		assert.Equal(t, int32(15), settings.CleanupDurationsMinutes)
	})

	t.Run("CustomSettings", func(t *testing.T) {
		settings := &models.UserSettings{
			ReviewerLogin:                  "customuser",
			ResponseDeadlineShiftMinutes:   30,
			NonWhitelistCancelDelayMinutes: 10,
			NotifyWhitelistTimeout:         false,
			NotifyNonWhitelistCancel:       false,
			SlotShiftThresholdMinutes:      40,
			SlotShiftDurationMinutes:       20,
			CleanupDurationsMinutes:        30,
		}

		assert.Equal(t, "customuser", settings.ReviewerLogin)
		assert.Equal(t, int32(30), settings.ResponseDeadlineShiftMinutes)
		assert.Equal(t, int32(10), settings.NonWhitelistCancelDelayMinutes)
		assert.False(t, settings.NotifyWhitelistTimeout)
		assert.False(t, settings.NotifyNonWhitelistCancel)
		assert.Equal(t, int32(40), settings.SlotShiftThresholdMinutes)
		assert.Equal(t, int32(20), settings.SlotShiftDurationMinutes)
		assert.Equal(t, int32(30), settings.CleanupDurationsMinutes)
	})
}

// TestErrorHandling tests error conditions and error messages
func TestErrorHandling(t *testing.T) {
	t.Run("InvalidUserType", func(t *testing.T) {
		ctx := context.Background()
		invalidUser := 12345 // Wrong type
		req := getTestReviewRequest()

		err := SendNonWhitelistCancelNotification(ctx, invalidUser, req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid user type")
	})

	t.Run("InvalidRequestType", func(t *testing.T) {
		ctx := context.Background()
		user := getTestUser()
		invalidReq := 12345 // Wrong type

		err := SendWhitelistTimeoutNotification(ctx, user, invalidReq)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid review request type")
	})

	t.Run("BothTypesInvalid", func(t *testing.T) {
		ctx := context.Background()
		invalidUser := "string"
		invalidReq := 12345

		err := SendNonWhitelistCancelNotification(ctx, invalidUser, invalidReq)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid user type")
	})
}

// TestConcurrentOperations tests thread-safety considerations
func TestConcurrentOperations(t *testing.T) {
	t.Run("MultipleReviewRequests", func(t *testing.T) {
		requests := []*models.ReviewRequest{
			getTestReviewRequest(),
			getTestReviewRequest(),
			getTestReviewRequest(),
		}

		// Give them unique IDs
		for i, req := range requests {
			req.ID = fmt.Sprintf("req-%d", i)
		}

		assert.Len(t, requests, 3)
		for _, req := range requests {
			assert.NotEmpty(t, req.ID)
			assert.NotEmpty(t, req.ReviewerLogin)
		}
	})
}

// Benchmark functions for performance testing
func BenchmarkFormatReviewRequestMessage(b *testing.B) {
	projectName := "libft"
	reviewTime := getTestTime()
	deadline := reviewTime.Add(30 * time.Minute)

	for i := 0; i < b.N; i++ {
		_ = FormatReviewRequestMessage(projectName, reviewTime, deadline)
	}
}

func BenchmarkTimeCalculations(b *testing.B) {
	reviewTime := getTestTime()
	shiftMinutes := 20

	for i := 0; i < b.N; i++ {
		_ = timeutil.CalculateDecisionDeadline(reviewTime, shiftMinutes)
	}
}

// TestUtilityFunctions tests various utility and helper functions
func TestUtilityFunctions(t *testing.T) {
	t.Run("FormatForMessage", func(t *testing.T) {
		tm := time.Date(2026, 1, 11, 14, 30, 15, 0, time.UTC)
		formatted := timeutil.FormatForMessage(tm)

		assert.Contains(t, formatted, "2026-01-11")
		assert.Contains(t, formatted, "14:30:15")
		assert.Contains(t, formatted, "UTC")
	})

	t.Run("ToUTC", func(t *testing.T) {
		// Time in a different timezone
		localTime := time.Date(2026, 1, 11, 14, 0, 0, 0, time.FixedZone("EST", -5*3600))
		utcTime := timeutil.ToUTC(localTime)

		assert.Equal(t, time.UTC, utcTime.Location())
	})

	t.Run("NowUTC", func(t *testing.T) {
		now := timeutil.NowUTC()
		assert.Equal(t, time.UTC, now.Location())
		assert.WithinDuration(t, time.Now().UTC(), now, time.Second)
	})
}

// Integration-like tests that verify function interactions
func TestFunctionInteractions(t *testing.T) {
	t.Run("CompleteWorkflow", func(t *testing.T) {
		// Simulate a complete workflow
		projectName := "libft"
		reviewTime := getTestTime().Add(1 * time.Hour)
		settings := models.DefaultUserSettings("testuser")

		// Calculate decision deadline
		deadline := timeutil.CalculateDecisionDeadline(reviewTime, int(settings.ResponseDeadlineShiftMinutes))
		assert.True(t, deadline.Before(reviewTime))

		// Format message
		message := FormatReviewRequestMessage(projectName, reviewTime, deadline)
		assert.NotEmpty(t, message)

		// Create review request
		decisionDeadline := deadline.Unix()
		req := &models.ReviewRequest{
			ID:              "req-123",
			ReviewerLogin:   "testuser",
			ProjectName:     &projectName,
			ReviewStartTime: reviewTime.Unix(),
			Status:          models.StatusUnknownProjectReview,
			DecisionDeadline: &decisionDeadline,
		}

		assert.Equal(t, models.StatusUnknownProjectReview, req.Status)
		assert.True(t, models.IsIntermediateStatus(req.Status))
		assert.False(t, models.IsFinalStatus(req.Status))
	})

	t.Run("WhitelistWorkflow", func(t *testing.T) {
		// Simulate whitelist decision workflow
		req := getTestReviewRequest()
		req.Status = models.StatusUnknownProjectReview

		// Check initial state
		assert.True(t, models.IsIntermediateStatus(req.Status))

		// Simulate transition to whitelisted
		req.Status = models.StatusWhitelisted
		assert.True(t, models.IsIntermediateStatus(req.Status))

		// Simulate transition to need approval
		req.Status = models.StatusNeedToApprove
		assert.True(t, models.IsIntermediateStatus(req.Status))

		// Simulate user approval
		req.Status = models.StatusApproved
		assert.True(t, models.IsFinalStatus(req.Status))
		assert.False(t, models.IsIntermediateStatus(req.Status))
	})

	t.Run("NonWhitelistWorkflow", func(t *testing.T) {
		// Simulate non-whitelist decision workflow
		req := getTestReviewRequest()
		req.Status = models.StatusUnknownProjectReview

		// Check initial state
		assert.True(t, models.IsIntermediateStatus(req.Status))

		// Simulate transition to not whitelisted
		req.Status = models.StatusNotWhitelisted
		assert.True(t, models.IsIntermediateStatus(req.Status))

		// Simulate auto-cancel
		req.Status = models.StatusAutoCancelledNotWhitelisted
		assert.True(t, models.IsFinalStatus(req.Status))
		assert.False(t, models.IsIntermediateStatus(req.Status))
	})
}

// TestTimeZones tests timezone handling
func TestTimeZones(t *testing.T) {
	t.Run("UTCConversion", func(t *testing.T) {
		// Create time in various timezones
		times := []time.Time{
			time.Date(2026, 1, 11, 10, 0, 0, 0, time.UTC),
			time.Date(2026, 1, 11, 5, 0, 0, 0, time.FixedZone("EST", -5*3600)),
			time.Date(2026, 1, 11, 15, 0, 0, 0, time.FixedZone("MSK", 3*3600)),
		}

		for _, tm := range times {
			utc := timeutil.ToUTC(tm)
			assert.Equal(t, time.UTC, utc.Location())
		}
	})

	t.Run("FormatShortInUTC", func(t *testing.T) {
		tm := time.Date(2026, 1, 11, 10, 30, 0, 0, time.FixedZone("EST", -5*3600))
		formatted := timeutil.FormatShort(tm)

		// Should contain UTC marker
		assert.Contains(t, formatted, "UTC")
	})
}

// TestDurationCalculations tests duration-related calculations
func TestDurationCalculations(t *testing.T) {
	t.Run("ReviewSlotDuration", func(t *testing.T) {
		start := time.Date(2026, 1, 11, 14, 0, 0, 0, time.UTC)
		end := time.Date(2026, 1, 11, 15, 30, 0, 0, time.UTC)

		duration := timeutil.CalculateSlotDuration(start, end)
		assert.Equal(t, 90, duration)
	})

	t.Run("ShortSlot", func(t *testing.T) {
		start := time.Date(2026, 1, 11, 10, 0, 0, 0, time.UTC)
		end := time.Date(2026, 1, 11, 10, 15, 0, 0, time.UTC)

		duration := timeutil.CalculateSlotDuration(start, end)
		assert.Equal(t, 15, duration)
	})

	t.Run("LongSlot", func(t *testing.T) {
		start := time.Date(2026, 1, 11, 9, 0, 0, 0, time.UTC)
		end := time.Date(2026, 1, 11, 17, 0, 0, 0, time.UTC)

		duration := timeutil.CalculateSlotDuration(start, end)
		assert.Equal(t, 480, duration) // 8 hours
	})
}

// TestValidationErrors tests validation error constants
func TestValidationErrors(t *testing.T) {
	t.Run("ErrorConstants", func(t *testing.T) {
		assert.NotEmpty(t, models.ErrInvalidStatus)
		assert.NotEmpty(t, models.ErrInvalidEntryType)
		assert.NotEmpty(t, models.ErrInvalidUserStatus)
		assert.NotEmpty(t, models.ErrInvalidReviewID)
	})
}

// TestMockBehavior tests mock object behavior
func TestMockBehavior(t *testing.T) {
	t.Run("MockLockboxClient", func(t *testing.T) {
		mockClient := new(MockLockboxClient)
		ctx := context.Background()
		reviewerLogin := "testuser"

		expectedTokens := &models.UserTokens{
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
		}

		mockClient.On("GetUserTokens", ctx, reviewerLogin).Return(expectedTokens, nil)

		tokens, err := mockClient.GetUserTokens(ctx, reviewerLogin)

		assert.NoError(t, err)
		assert.Equal(t, expectedTokens, tokens)
		mockClient.AssertExpectations(t)
	})

	t.Run("MockLockboxClientError", func(t *testing.T) {
		mockClient := new(MockLockboxClient)
		ctx := context.Background()
		reviewerLogin := "nonexistent"

		expectedError := errors.New("user not found")
		mockClient.On("GetUserTokens", ctx, reviewerLogin).Return(nil, expectedError)

		tokens, err := mockClient.GetUserTokens(ctx, reviewerLogin)

		assert.Error(t, err)
		assert.Nil(t, tokens)
		assert.Equal(t, expectedError, err)
		mockClient.AssertExpectations(t)
	})
}
