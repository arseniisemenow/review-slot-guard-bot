package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/arseniisemenow/review-slot-guard-bot/common/pkg/models"
	"github.com/arseniisemenow/review-slot-guard-bot/common/pkg/telegram"
	"github.com/arseniisemenow/review-slot-guard-bot/common/pkg/timeutil"
	"github.com/arseniisemenow/review-slot-guard-bot/common/pkg/ydb"
	"github.com/arseniisemenow/review-slot-guard-bot/functions/periodic_job/internal/logic"
)

// Handler is the Yandex Cloud Function entry point
func Handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.New(os.Stdout, "[PERIODIC_JOB] ", log.LstdFlags)

	logger.Println("Starting periodic job execution")

	// 1. Get all active users
	users, err := ydb.GetActiveUsers(ctx)
	if err != nil {
		logger.Printf("Failed to get active users: %v", err)
		http.Error(w, fmt.Sprintf("Failed to get active users: %v", err), http.StatusInternalServerError)
		return
	}

	logger.Printf("Found %d active users", len(users))

	// Process each user independently
	for _, user := range users {
		if err := processUser(ctx, user, logger); err != nil {
			logger.Printf("Error processing user %s: %v", user.ReviewerLogin, err)
			// Continue processing other users
		}
	}

	logger.Println("Periodic job completed successfully")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// processUser handles all logic for a single user
func processUser(ctx context.Context, user *models.User, logger *log.Logger) error {
	logger.Printf("Processing user: %s", user.ReviewerLogin)

	// 1. Get user settings
	settings, err := ydb.GetUserSettings(ctx, user.ReviewerLogin)
	if err != nil {
		return fmt.Errorf("failed to get user settings: %w", err)
	}

	// 2. Get existing review requests in intermediate states
	intermediateRequests, err := ydb.GetReviewRequestsByUserAndStatus(ctx, user.ReviewerLogin, []string{
		models.StatusUnknownProjectReview,
		models.StatusKnownProjectReview,
		models.StatusWhitelisted,
		models.StatusNotWhitelisted,
		models.StatusNeedToApprove,
		models.StatusWaitingForApprove,
	})
	if err != nil {
		return fmt.Errorf("failed to get review requests: %w", err)
	}

	logger.Printf("User %s has %d intermediate review requests", user.ReviewerLogin, len(intermediateRequests))

	// 3. Process each review request through the state machine
	for _, req := range intermediateRequests {
		if err := processReviewRequest(ctx, req, user, settings, logger); err != nil {
			logger.Printf("Error processing review request %s: %v", req.ID, err)
		}
	}

	// 4. Check for new bookings from calendar
	if err := checkNewBookings(ctx, user, settings, logger); err != nil {
		logger.Printf("Error checking new bookings for user %s: %v", user.ReviewerLogin, err)
	}

	return nil
}

// processReviewRequest processes a single review request through the state machine
func processReviewRequest(ctx context.Context, req *models.ReviewRequest, user *models.User, settings *models.UserSettings, logger *log.Logger) error {
	logger.Printf("Processing review request %s (status: %s)", req.ID, req.Status)

	switch req.Status {
	case models.StatusUnknownProjectReview:
		return processUnknownProjectReview(ctx, req, user, settings, logger)

	case models.StatusKnownProjectReview:
		return processKnownProjectReview(ctx, req, user, settings, logger)

	case models.StatusWhitelisted:
		return processWhitelisted(ctx, req, user, settings, logger)

	case models.StatusNotWhitelisted:
		return processNotWhitelisted(ctx, req, user, settings, logger)

	case models.StatusNeedToApprove:
		return processNeedToApprove(ctx, req, user, settings, logger)

	case models.StatusWaitingForApprove:
		return processWaitingForApprove(ctx, req, user, settings, logger)

	default:
		return fmt.Errorf("unexpected status: %s", req.Status)
	}
}

// processUnknownProjectReview: Resolve project from notification
func processUnknownProjectReview(ctx context.Context, req *models.ReviewRequest, user *models.User, settings *models.UserSettings, logger *log.Logger) error {
	// Step 3a: Extract project name from notification
	projectName, err := logic.ExtractProjectNameFromNotification(ctx, user.ReviewerLogin, req.NotificationID)
	if err != nil {
		return fmt.Errorf("failed to extract project name: %w", err)
	}

	// Step 3b: Look up family label
	familyLabel, err := ydb.GetFamilyLabelForProject(ctx, projectName)
	if err != nil {
		// Project not in project_families - lazy load
		logger.Printf("Project %s not found in project_families, triggering lazy load", projectName)
		if err := logic.PopulateProjectFamilies(ctx, user.ReviewerLogin); err != nil {
			return fmt.Errorf("failed to populate project families: %w", err)
		}

		// Retry lookup
		familyLabel, err = ydb.GetFamilyLabelForProject(ctx, projectName)
		if err != nil {
			return fmt.Errorf("project %s not found after population", projectName)
		}
	}

	// Step 3c: Update review request with project info and transition to KNOWN_PROJECT_REVIEW
	err = ydb.UpdateReviewRequestWithProjectInfo(ctx, req.ID, projectName, familyLabel, *req.NotificationID)
	if err != nil {
		return fmt.Errorf("failed to update review request: %w", err)
	}

	logger.Printf("Review request %s: UNKNOWN_PROJECT_REVIEW -> KNOWN_PROJECT_REVIEW", req.ID)
	return nil
}

// processKnownProjectReview: Check whitelist and time proximity
func processKnownProjectReview(ctx context.Context, req *models.ReviewRequest, user *models.User, settings *models.UserSettings, logger *log.Logger) error {
	projectName := ""
	if req.ProjectName != nil {
		projectName = *req.ProjectName
	}

	familyLabel := ""
	if req.FamilyLabel != nil {
		familyLabel = *req.FamilyLabel
	}

	// Step 4: Check if in whitelist
	inWhitelist, err := ydb.IsInWhitelist(ctx, user.ReviewerLogin, projectName, familyLabel)
	if err != nil {
		return fmt.Errorf("failed to check whitelist: %w", err)
	}

	reviewStartTime := timeutil.FromUnixSeconds(req.ReviewStartTime)
	now := time.Now()

	// Step 5: Check if review is within decision threshold
	deadline := timeutil.CalculateDecisionDeadline(reviewStartTime, int(settings.ResponseDeadlineShiftMinutes))
	minutesUntilDeadline := timeutil.MinutesUntil(deadline)

	// Check if we need to ask user for decision NOW
	needToAskNow := minutesUntilDeadline <= 0 || timeutil.ShouldShiftSlot(reviewStartTime, int(settings.SlotShiftThresholdMinutes))

	if needToAskNow {
		// Step 5b: Transition to NEED_TO_APPROVE
		err = ydb.UpdateReviewRequestStatus(ctx, req.ID, models.StatusNeedToApprove, nil)
		if err != nil {
			return fmt.Errorf("failed to update status: %w", err)
		}
		logger.Printf("Review request %s: KNOWN_PROJECT_REVIEW -> NEED_TO_APPROVE (deadline approaching)", req.ID)
		return nil
	}

	if inWhitelist {
		// Step 5a: Transition to WHITELISTED
		err = ydb.UpdateReviewRequestStatus(ctx, req.ID, models.StatusWhitelisted, nil)
		if err != nil {
			return fmt.Errorf("failed to update status: %w", err)
		}
		logger.Printf("Review request %s: KNOWN_PROJECT_REVIEW -> WHITELISTED", req.ID)
	} else {
		// Step 5a: Transition to NOT_WHITELISTED
		cancelTime := timeutil.CalculateNonWhitelistCancelTime(int(settings.NonWhitelistCancelDelayMinutes))
		err = ydb.UpdateReviewRequestToNotWhitelisted(ctx, req.ID, cancelTime.Unix())
		if err != nil {
			return fmt.Errorf("failed to update status: %w", err)
		}
		logger.Printf("Review request %s: KNOWN_PROJECT_REVIEW -> NOT_WHITELISTED", req.ID)
	}

	return nil
}

// processWhitelisted: Check if slot needs shifting
func processWhitelisted(ctx context.Context, req *models.ReviewRequest, user *models.User, settings *models.UserSettings, logger *log.Logger) error {
	reviewStartTime := timeutil.FromUnixSeconds(req.ReviewStartTime)

	// Step 6: Check if slot should be shifted
	if timeutil.ShouldShiftSlot(reviewStartTime, int(settings.SlotShiftThresholdMinutes)) {
		slotDuration := timeutil.CalculateSlotDuration(reviewStartTime, reviewStartTime.Add(time.Duration(req.ReviewStartTime)*time.Second))

		// Step 6a: Check if slot duration should be cleaned up
		if slotDuration <= int(settings.CleanupDurationsMinutes) {
			// Cancel the slot
			if err := logic.CancelCalendarSlot(ctx, user.ReviewerLogin, req.CalendarSlotID); err != nil {
				logger.Printf("Failed to cancel slot %s: %v", req.CalendarSlotID, err)
			}

			// Transition to AUTO_CANCELLED
			now := time.Now().Unix()
			err := ydb.UpdateReviewRequestStatus(ctx, req.ID, models.StatusAutoCancelled, &now)
			if err != nil {
				return fmt.Errorf("failed to update status: %w", err)
			}
			logger.Printf("Review request %s: WHITELISTED -> AUTO_CANCELLED (short slot)", req.ID)
			return nil
		}

		// Step 6b: Shift the slot
		newStartTime := reviewStartTime.Add(-time.Duration(settings.SlotShiftDurationMinutes) * time.Minute)
		newEndTime := newStartTime.Add(time.Duration(slotDuration) * time.Minute)

		if err := logic.ChangeCalendarSlot(ctx, user.ReviewerLogin, req.CalendarSlotID, newStartTime, newEndTime); err != nil {
			logger.Printf("Failed to shift slot %s: %v", req.CalendarSlotID, err)
			// If shift fails, cancel the slot
			if err := logic.CancelCalendarSlot(ctx, user.ReviewerLogin, req.CalendarSlotID); err != nil {
				logger.Printf("Failed to cancel slot %s: %v", req.CalendarSlotID, err)
			}

			now := time.Now().Unix()
			err := ydb.UpdateReviewRequestStatus(ctx, req.ID, models.StatusAutoCancelled, &now)
			if err != nil {
				return fmt.Errorf("failed to update status: %w", err)
			}
			logger.Printf("Review request %s: WHITELISTED -> AUTO_CANCELLED (shift failed)", req.ID)
			return nil
		}

		logger.Printf("Review request %s: Slot shifted from %s to %s", req.ID,
			timeutil.FormatShort(reviewStartTime), timeutil.FormatShort(newStartTime))
	}

	return nil
}

// processNotWhitelisted: Check cancel timeout
func processNotWhitelisted(ctx context.Context, req *models.ReviewRequest, user *models.User, settings *models.UserSettings, logger *log.Logger) error {
	if req.NonWhitelistCancelAt == nil {
		return fmt.Errorf("non_whitelist_cancel_at is nil for NOT_WHITELISTED review")
	}

	cancelTime := timeutil.FromUnixSeconds(*req.NonWhitelistCancelAt)

	// Check if cancel time has passed
	if time.Now().After(cancelTime) {
		// Send notification if enabled
		if settings.NotifyNonWhitelistCancel {
			if err := logic.SendNonWhitelistCancelNotification(ctx, user, req); err != nil {
				logger.Printf("Failed to send cancel notification: %v", err)
			}
		}

		// Cancel the slot
		if err := logic.CancelCalendarSlot(ctx, user.ReviewerLogin, req.CalendarSlotID); err != nil {
			logger.Printf("Failed to cancel slot %s: %v", req.CalendarSlotID, err)
		}

		// Transition to AUTO_CANCELLED_NOT_WHITELISTED
		now := time.Now().Unix()
		err := ydb.UpdateReviewRequestStatus(ctx, req.ID, models.StatusAutoCancelledNotWhitelisted, &now)
		if err != nil {
			return fmt.Errorf("failed to update status: %w", err)
		}
		logger.Printf("Review request %s: NOT_WHITELISTED -> AUTO_CANCELLED_NOT_WHITELISTED", req.ID)
	}

	return nil
}

// processNeedToApprove: Send Telegram message with buttons
func processNeedToApprove(ctx context.Context, req *models.ReviewRequest, user *models.User, settings *models.UserSettings, logger *log.Logger) error {
	projectName := "Unknown Project"
	if req.ProjectName != nil {
		projectName = *req.ProjectName
	}

	reviewStartTime := timeutil.FromUnixSeconds(req.ReviewStartTime)
	deadline := timeutil.CalculateDecisionDeadline(reviewStartTime, int(settings.ResponseDeadlineShiftMinutes))

	// Create Telegram message
	message := logic.FormatReviewRequestMessage(projectName, reviewStartTime, deadline)

	// Send message with buttons
	telegramClient, err := telegram.NewBotClientFromEnv()
	if err != nil {
		return fmt.Errorf("failed to create Telegram client: %w", err)
	}

	approveData := fmt.Sprintf("APPROVE:%s", req.ID)
	declineData := fmt.Sprintf("DECLINE:%s", req.ID)

	messageID, err := telegramClient.SendTwoButtonKeyboard(user.TelegramChatID, message, approveData, declineData)
	if err != nil {
		return fmt.Errorf("failed to send Telegram message: %w", err)
	}

	// Update review request with decision deadline and message ID
	err = ydb.UpdateReviewRequestToWaitingForApprove(ctx, req.ID, deadline.Unix(), fmt.Sprintf("%d", messageID))
	if err != nil {
		return fmt.Errorf("failed to update review request: %w", err)
	}

	logger.Printf("Review request %s: NEED_TO_APPROVE -> WAITING_FOR_APPROVE", req.ID)
	return nil
}

// processWaitingForApprove: Check if deadline has passed
func processWaitingForApprove(ctx context.Context, req *models.ReviewRequest, user *models.User, settings *models.UserSettings, logger *log.Logger) error {
	if req.DecisionDeadline == nil {
		return fmt.Errorf("decision_deadline is nil for WAITING_FOR_APPROVE review")
	}

	deadline := timeutil.FromUnixSeconds(*req.DecisionDeadline)

	// Check if deadline has passed
	if time.Now().After(deadline) {
		// Send timeout notification if enabled
		if settings.NotifyWhitelistTimeout {
			if err := logic.SendWhitelistTimeoutNotification(ctx, user, req); err != nil {
				logger.Printf("Failed to send timeout notification: %v", err)
			}
		}

		// Cancel the slot
		if err := logic.CancelCalendarSlot(ctx, user.ReviewerLogin, req.CalendarSlotID); err != nil {
			logger.Printf("Failed to cancel slot %s: %v", req.CalendarSlotID, err)
		}

		// Transition to AUTO_CANCELLED
		now := time.Now().Unix()
		err := ydb.UpdateReviewRequestStatus(ctx, req.ID, models.StatusAutoCancelled, &now)
		if err != nil {
			return fmt.Errorf("failed to update status: %w", err)
		}
		logger.Printf("Review request %s: WAITING_FOR_APPROVE -> AUTO_CANCELLED (deadline passed)", req.ID)
	}

	return nil
}

// checkNewBookings looks for new bookings in the calendar and creates review requests
func checkNewBookings(ctx context.Context, user *models.User, settings *models.UserSettings, logger *log.Logger) error {
	// Step 1: Fetch calendar events
	from := time.Now().Add(-2 * time.Hour)
	to := time.Now().Add(24 * time.Hour)

	events, err := logic.GetCalendarEvents(ctx, user.ReviewerLogin, from, to)
	if err != nil {
		return fmt.Errorf("failed to get calendar events: %w", err)
	}

	// Step 2: Extract bookings
	bookings := logic.ExtractBookings(events)

	// Step 3: Check for new bookings
	for _, booking := range bookings {
		// Check if review request already exists for this slot
		existing, err := ydb.GetReviewRequestByCalendarSlotID(ctx, booking.EventSlotID)
		if err == nil && existing != nil {
			// Review request already exists, skip
			continue
		}

		// Create new review request
		reviewID := uuid.New().String()

		req := &models.ReviewRequest{
			ID:              reviewID,
			ReviewerLogin:   user.ReviewerLogin,
			ReviewStartTime: booking.Start.Unix(),
			CalendarSlotID:  booking.EventSlotID,
			Status:          models.StatusUnknownProjectReview,
			CreatedAt:       time.Now().Unix(),
		}

		// Extract notification ID from booking
		notificationID := booking.ID
		req.NotificationID = &notificationID

		if err := ydb.CreateReviewRequest(ctx, req); err != nil {
			logger.Printf("Failed to create review request for slot %s: %v", booking.EventSlotID, err)
			continue
		}

		logger.Printf("Created new review request %s for slot %s", reviewID, booking.EventSlotID)
	}

	return nil
}
