package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	tba "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/arseniisemenow/review-slot-guard-bot-common/pkg/external"
	"github.com/arseniisemenow/review-slot-guard-bot-common/pkg/models"
	"github.com/arseniisemenow/review-slot-guard-bot-common/pkg/telegram"
	"github.com/arseniisemenow/review-slot-guard-bot-common/pkg/timeutil"
	"github.com/arseniisemenow/review-slot-guard-bot-common/pkg/ydb"
)

// HandleApprove handles the APPROVE button click
func HandleApprove(ctx context.Context, user *models.User, req *models.ReviewRequest, callback *tba.CallbackQuery, logger *log.Logger) error {
	logger.Printf("User %s approved review %s", user.ReviewerLogin, req.ID)

	// Get user tokens (for future API calls if needed)
	_, err := ydb.GetUserTokens(ctx, user.ReviewerLogin)
	if err != nil {
		return sendCallbackError(callback, fmt.Sprintf("Failed to get tokens: %v", err))
	}

	// The review is already whitelisted or was explicitly approved
	// Transition to APPROVED
	now := time.Now().Unix()
	err = ydb.UpdateReviewRequestStatus(ctx, req.ID, models.StatusApproved, &now)
	if err != nil {
		return sendCallbackError(callback, fmt.Sprintf("Failed to update status: %v", err))
	}

	// Update Telegram message
	bot, _ := telegram.NewBotClientFromEnv()
	messageText := fmt.Sprintf("✅ *Review Approved*\n\nProject: %s\nTime: %s",
		getProjectName(req),
		timeutil.FormatShort(timeutil.FromUnixSeconds(req.ReviewStartTime)))

	if req.TelegramMessageID != nil {
		msgID, _ := strconv.Atoi(*req.TelegramMessageID)
		bot.EditMessage(user.TelegramChatID, msgID, messageText)
	}

	// Answer callback
	bot.AnswerCallbackQuery(callback.ID, "Review approved!")

	return nil
}

// HandleDecline handles the DECLINE button click
func HandleDecline(ctx context.Context, user *models.User, req *models.ReviewRequest, callback *tba.CallbackQuery, logger *log.Logger) error {
	logger.Printf("User %s declined review %s", user.ReviewerLogin, req.ID)

	// Cancel the slot via s21 API
	tokens, err := ydb.GetUserTokens(ctx, user.ReviewerLogin)
	if err != nil {
		return sendCallbackError(callback, fmt.Sprintf("Failed to get tokens: %v", err))
	}

	client := external.NewS21Client(tokens.AccessToken, tokens.RefreshToken)
	err = client.CancelSlot(ctx, req.CalendarSlotID)
	if err != nil {
		logger.Printf("Failed to cancel slot %s: %v", req.CalendarSlotID, err)
		// Continue anyway - the user wants to decline
	}

	// Transition to CANCELLED
	now := time.Now().Unix()
	err = ydb.UpdateReviewRequestStatus(ctx, req.ID, models.StatusCancelled, &now)
	if err != nil {
		return sendCallbackError(callback, fmt.Sprintf("Failed to update status: %v", err))
	}

	// Update Telegram message
	bot, _ := telegram.NewBotClientFromEnv()
	messageText := fmt.Sprintf("❌ *Review Cancelled*\n\nProject: %s\nTime: %s",
		getProjectName(req),
		timeutil.FormatShort(timeutil.FromUnixSeconds(req.ReviewStartTime)))

	if req.TelegramMessageID != nil {
		msgID, _ := strconv.Atoi(*req.TelegramMessageID)
		bot.EditMessage(user.TelegramChatID, msgID, messageText)
	}

	// Answer callback
	bot.AnswerCallbackQuery(callback.ID, "Review cancelled")

	return nil
}

// sendCallbackError sends an error response via callback
func sendCallbackError(callback *tba.CallbackQuery, message string) error {
	bot, _ := telegram.NewBotClientFromEnv()
	bot.AnswerCallbackQuery(callback.ID, message)
	return fmt.Errorf("callback error: %s", message)
}

// getProjectName extracts project name from review request
func getProjectName(req *models.ReviewRequest) string {
	if req.ProjectName != nil {
		return *req.ProjectName
	}
	return "Unknown Project"
}
