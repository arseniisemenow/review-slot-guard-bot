package logic

import (
	"context"
	"fmt"
	"time"

	"github.com/arseniisemenow/review-slot-guard-bot/common/pkg/external"
	"github.com/arseniisemenow/review-slot-guard-bot/common/pkg/lockbox"
	"github.com/arseniisemenow/review-slot-guard-bot/common/pkg/models"
	"github.com/arseniisemenow/review-slot-guard-bot/common/pkg/telegram"
	"github.com/arseniisemenow/review-slot-guard-bot/common/pkg/timeutil"
	"github.com/arseniisemenow/review-slot-guard-bot/common/pkg/ydb"
	"github.com/arseniisemenow/s21auto-client-go/requests"
)

// ExtractProjectNameFromNotification extracts project name from a notification
func ExtractProjectNameFromNotification(ctx context.Context, reviewerLogin, notificationID string) (string, error) {
	// Get user tokens from Lockbox
	tokens, err := lockbox.GetUserTokens(ctx, reviewerLogin)
	if err != nil {
		return "", fmt.Errorf("failed to get user tokens: %w", err)
	}

	// Create s21 client
	client := external.NewS21Client(tokens.AccessToken, tokens.RefreshToken)

	// Get notifications
	notificationsResp, err := client.GetNotifications(ctx, 0, 100)
	if err != nil {
		return "", fmt.Errorf("failed to get notifications: %w", err)
	}

	// Find the matching notification
	notifications := external.ExtractNotifications(notificationsResp)
	for _, notif := range notifications {
		if notif.ID == notificationID {
			// Extract project name from message
			// The notification message contains the project name
			return external.ExtractProjectNameFromMessage(notif.Message), nil
		}
	}

	return "", fmt.Errorf("notification not found: %s", notificationID)
}

// PopulateProjectFamilies fetches and stores all project families
func PopulateProjectFamilies(ctx context.Context, reviewerLogin string) error {
	// Get user tokens from Lockbox
	tokens, err := lockbox.GetUserTokens(ctx, reviewerLogin)
	if err != nil {
		return fmt.Errorf("failed to get user tokens: %w", err)
	}

	// Create s21 client
	client := external.NewS21Client(tokens.AccessToken, tokens.RefreshToken)

	// Get project graph
	graph, err := client.GetProjectGraph(ctx, reviewerLogin)
	if err != nil {
		return fmt.Errorf("failed to get project graph: %w", err)
	}

	// Extract families
	families, err := external.ExtractFamilies(graph)
	if err != nil {
		return fmt.Errorf("failed to extract families: %w", err)
	}

	// Store in YDB
	err = ydb.UpsertProjectFamilies(ctx, families)
	if err != nil {
		return fmt.Errorf("failed to store project families: %w", err)
	}

	return nil
}

// CancelCalendarSlot cancels a calendar slot via s21 API
func CancelCalendarSlot(ctx context.Context, reviewerLogin, slotID string) error {
	// Get user tokens from Lockbox
	tokens, err := lockbox.GetUserTokens(ctx, reviewerLogin)
	if err != nil {
		return fmt.Errorf("failed to get user tokens: %w", err)
	}

	// Create s21 client
	client := external.NewS21Client(tokens.AccessToken, tokens.RefreshToken)

	// Cancel the slot
	return client.CancelSlot(ctx, slotID)
}

// ChangeCalendarSlot changes the timing of a calendar slot
func ChangeCalendarSlot(ctx context.Context, reviewerLogin, slotID string, newStart, newEnd time.Time) error {
	// Get user tokens from Lockbox
	tokens, err := lockbox.GetUserTokens(ctx, reviewerLogin)
	if err != nil {
		return fmt.Errorf("failed to get user tokens: %w", err)
	}

	// Create s21 client
	client := external.NewS21Client(tokens.AccessToken, tokens.RefreshToken)

	// Change the slot
	return client.ChangeEventSlot(ctx, slotID, newStart, newEnd)
}

// SendNonWhitelistCancelNotification sends a notification about non-whitelist cancellation
func SendNonWhitelistCancelNotification(ctx context.Context, user interface{}, req interface{}) error {
	// Type assert to get the actual types
	u, ok := user.(*models.User)
	if !ok {
		return fmt.Errorf("invalid user type")
	}
	r, ok := req.(*models.ReviewRequest)
	if !ok {
		return fmt.Errorf("invalid review request type")
	}

	projectName := "Unknown Project"
	if r.ProjectName != nil {
		projectName = *r.ProjectName
	}

	bot, err := telegram.NewBotClientFromEnv()
	if err != nil {
		return fmt.Errorf("failed to create telegram client: %w", err)
	}

	message := fmt.Sprintf("❌ *Review Auto-Cancelled*\n\n"+
		"Project: %s\n"+
		"Time: %s\n\n"+
		"This project is not in your whitelist and was automatically cancelled.",
		projectName,
		timeutil.FormatShort(timeutil.FromUnixSeconds(r.ReviewStartTime)))

	bot.SendPlainMessage(u.TelegramChatID, message)
	return nil
}

// SendWhitelistTimeoutNotification sends a notification about whitelist timeout
func SendWhitelistTimeoutNotification(ctx context.Context, user interface{}, req interface{}) error {
	// Type assert to get the actual types
	u, ok := user.(*models.User)
	if !ok {
		return fmt.Errorf("invalid user type")
	}
	r, ok := req.(*models.ReviewRequest)
	if !ok {
		return fmt.Errorf("invalid review request type")
	}

	projectName := "Unknown Project"
	if r.ProjectName != nil {
		projectName = *r.ProjectName
	}

	bot, err := telegram.NewBotClientFromEnv()
	if err != nil {
		return fmt.Errorf("failed to create telegram client: %w", err)
	}

	message := fmt.Sprintf("⏰ *Review Timeout*\n\n"+
		"Project: %s\n"+
		"Time: %s\n\n"+
		"You did not respond in time and this review was automatically cancelled.",
		projectName,
		timeutil.FormatShort(timeutil.FromUnixSeconds(r.ReviewStartTime)))

	bot.SendPlainMessage(u.TelegramChatID, message)
	return nil
}

// FormatReviewRequestMessage creates the Telegram message for review request
func FormatReviewRequestMessage(projectName string, reviewStartTime, deadline time.Time) string {
	return fmt.Sprintf("*Review Request*\n\n"+
		"Project: %s\n"+
		"Time: %s\n\n"+
		"Please respond by %s.\n\n"+
		"Use the buttons below to approve or decline.",
		projectName,
		timeutil.FormatShort(reviewStartTime),
		timeutil.FormatShort(deadline))
}

// NewTelegramClient creates a new Telegram bot client
func NewTelegramClient() *telegram.BotClient {
	bot, err := telegram.NewBotClientFromEnv()
	if err != nil {
		// In a real scenario, we would handle this error properly
		// For now, return nil to maintain compatibility
		return nil
	}
	return bot
}

// GetCalendarEvents fetches calendar events for a user
func GetCalendarEvents(ctx context.Context, reviewerLogin string, from, to time.Time) (*requests.CalendarGetEvents_Data, error) {
	// Get user tokens from Lockbox
	tokens, err := lockbox.GetUserTokens(ctx, reviewerLogin)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tokens: %w", err)
	}

	// Create s21 client
	client := external.NewS21Client(tokens.AccessToken, tokens.RefreshToken)

	// Get calendar events
	return client.GetCalendarEvents(ctx, from, to)
}

// ExtractBookings extracts bookings from calendar events
func ExtractBookings(data *requests.CalendarGetEvents_Data) []external.CalendarBooking {
	return external.ExtractBookings(data)
}
