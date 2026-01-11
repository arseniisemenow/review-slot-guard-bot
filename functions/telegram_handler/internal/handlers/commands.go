package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tba "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/arseniisemenow/review-slot-guard-bot/common/pkg/external"
	"github.com/arseniisemenow/review-slot-guard-bot/common/pkg/lockbox"
	"github.com/arseniisemenow/review-slot-guard-bot/common/pkg/models"
	"github.com/arseniisemenow/review-slot-guard-bot/common/pkg/telegram"
	"github.com/arseniisemenow/review-slot-guard-bot/common/pkg/timeutil"
	"github.com/arseniisemenow/review-slot-guard-bot/common/pkg/ydb"
)

// HandleStart handles the /start command - initiates authentication flow
func HandleStart(ctx context.Context, message *tba.Message, logger *log.Logger) error {
	chatID := message.From.ID

	// Check if user already exists
	user, err := ydb.GetUserByTelegramChatID(ctx, chatID)
	if err == nil && user != nil {
		sendMessage(chatID, fmt.Sprintf("Welcome back, %s! You are already authenticated.", user.ReviewerLogin))
		return nil
	}

	// Request login:password
	sendMessage(chatID, "Please authenticate by sending your School 21 credentials in the format:\n\n`login:password`\n\nYour credentials will be stored securely in Yandex Cloud Lockbox.")
	return nil
}

// HandleSettings handles the /settings command - shows current settings
func HandleSettings(ctx context.Context, message *tba.Message, logger *log.Logger) error {
	chatID := message.From.ID

	// Get user
	user, err := ydb.GetUserByTelegramChatID(ctx, chatID)
	if err != nil {
		sendMessage(chatID, "User not found. Please use /start to authenticate.")
		return nil
	}

	// Get settings
	settings, err := ydb.GetUserSettings(ctx, user.ReviewerLogin)
	if err != nil {
		sendMessage(chatID, "Failed to retrieve settings.")
		return nil
	}

	// Format settings message
	msg := fmt.Sprintf("*Your Settings*\n\n"+
		"📅 Response Deadline Shift: %d minutes\n"+
		"⏱️ Non-Whitelist Cancel Delay: %d minutes\n"+
		"🔔 Notify Whitelist Timeout: %s\n"+
		"🔔 Notify Non-Whitelist Cancel: %s\n"+
		"🔄 Slot Shift Threshold: %d minutes\n"+
		"⬇️ Slot Shift Duration: %d minutes\n"+
		"🧹 Cleanup Duration: %d minutes",
		settings.ResponseDeadlineShiftMinutes,
		settings.NonWhitelistCancelDelayMinutes,
		boolToYesNo(settings.NotifyWhitelistTimeout),
		boolToYesNo(settings.NotifyNonWhitelistCancel),
		settings.SlotShiftThresholdMinutes,
		settings.SlotShiftDurationMinutes,
		settings.CleanupDurationsMinutes)

	sendMessage(chatID, msg)
	return nil
}

// HandleWhitelist handles the /whitelist command - shows current whitelist
func HandleWhitelist(ctx context.Context, message *tba.Message, logger *log.Logger) error {
	chatID := message.From.ID

	// Get user
	user, err := ydb.GetUserByTelegramChatID(ctx, chatID)
	if err != nil {
		sendMessage(chatID, "User not found. Please use /start to authenticate.")
		return nil
	}

	// Get whitelist
	entries, err := ydb.GetUserWhitelist(ctx, user.ReviewerLogin)
	if err != nil {
		sendMessage(chatID, "Failed to retrieve whitelist.")
		return nil
	}

	if len(entries) == 0 {
		sendMessage(chatID, "Your whitelist is empty.\n\nUse /whitelist_add to add projects or families.")
		return nil
	}

	// Format whitelist
	var families []string
	var projects []string

	for _, entry := range entries {
		if entry.EntryType == models.EntryTypeFamily {
			families = append(families, entry.Name)
		} else {
			projects = append(projects, entry.Name)
		}
	}

	msg := "*Your Whitelist*\n\n"

	if len(families) > 0 {
		msg += "📁 Families:\n" + formatList(families)
	}

	if len(projects) > 0 {
		msg += "📦 Projects:\n" + formatList(projects)
	}

	sendMessage(chatID, msg)
	return nil
}

// HandleWhitelistAdd handles the /whitelist_add command
func HandleWhitelistAdd(ctx context.Context, message *tba.Message, logger *log.Logger) error {
	chatID := message.From.ID

	// Get user
	user, err := ydb.GetUserByTelegramChatID(ctx, chatID)
	if err != nil {
		sendMessage(chatID, "User not found. Please use /start to authenticate.")
		return nil
	}

	// Parse arguments
	args := strings.SplitN(message.CommandArguments(), " ", 2)
	if len(args) < 2 {
		sendMessage(chatID, "Usage: /whitelist_add <family|project> <name>\n\nExample:\n/whitelist_add family \"C - I\"\n/whitelist_add project \"go-concurrency\"")
		return nil
	}

	entryType := strings.ToUpper(args[0])
	name := args[1]

	if !models.IsValidEntryType(entryType) {
		sendMessage(chatID, "Invalid entry type. Use 'family' or 'project'.")
		return nil
	}

	// Add to whitelist
	entry := &models.WhitelistEntry{
		ReviewerLogin: user.ReviewerLogin,
		EntryType:     entryType,
		Name:          name,
	}

	err = ydb.AddToWhitelist(ctx, entry)
	if err != nil {
		sendMessage(chatID, fmt.Sprintf("Failed to add to whitelist: %v", err))
		return nil
	}

	sendMessage(chatID, fmt.Sprintf("✅ Added %s to your whitelist.", name))
	return nil
}

// HandleWhitelistRemove handles the /whitelist_remove command
func HandleWhitelistRemove(ctx context.Context, message *tba.Message, logger *log.Logger) error {
	chatID := message.From.ID

	// Get user
	user, err := ydb.GetUserByTelegramChatID(ctx, chatID)
	if err != nil {
		sendMessage(chatID, "User not found. Please use /start to authenticate.")
		return nil
	}

	name := strings.TrimSpace(message.CommandArguments())
	if name == "" {
		sendMessage(chatID, "Usage: /whitelist_remove <name>\n\nExample: /whitelist_remove \"C - I\"")
		return nil
	}

	err = ydb.RemoveFromWhitelist(ctx, user.ReviewerLogin, name)
	if err != nil {
		sendMessage(chatID, fmt.Sprintf("Failed to remove from whitelist: %v", err))
		return nil
	}

	sendMessage(chatID, fmt.Sprintf("✅ Removed %s from your whitelist.", name))
	return nil
}

// HandleSetDeadlineShift handles the /set_deadline_shift command
func HandleSetDeadlineShift(ctx context.Context, message *tba.Message, logger *log.Logger) error {
	return handleNumericSetting(ctx, message, "response_deadline_shift_minutes", 20, 60, 1)
}

// HandleSetCancelDelay handles the /set_cancel_delay command
func HandleSetCancelDelay(ctx context.Context, message *tba.Message, logger *log.Logger) error {
	return handleNumericSetting(ctx, message, "non_whitelist_cancel_delay_minutes", 5, 10, 1)
}

// HandleSetSlotShiftThreshold handles the /set_slot_shift_threshold command
func HandleSetSlotShiftThreshold(ctx context.Context, message *tba.Message, logger *log.Logger) error {
	return handleNumericSetting(ctx, message, "slot_shift_threshold_minutes", 20, 60, 5)
}

// HandleSetSlotShiftDuration handles the /set_slot_shift_duration command
func HandleSetSlotShiftDuration(ctx context.Context, message *tba.Message, logger *log.Logger) error {
	return handleNumericSetting(ctx, message, "slot_shift_duration_minutes", 15, 60, 15)
}

// HandleSetCleanupDuration handles the /set_cleanup_duration command
func HandleSetCleanupDuration(ctx context.Context, message *tba.Message, logger *log.Logger) error {
	chatID := message.From.ID

	// Get user
	user, err := ydb.GetUserByTelegramChatID(ctx, chatID)
	if err != nil {
		sendMessage(chatID, "User not found. Please use /start to authenticate.")
		return nil
	}

	arg := strings.TrimSpace(message.CommandArguments())
	value, err := strconv.Atoi(arg)
	if err != nil {
		sendMessage(chatID, "Usage: /set_cleanup_duration <minutes>\n\nAllowed values: 15, 30, 45, 60")
		return nil
	}

	// Validate: must be one of 15, 30, 45, 60
	validValues := []int{15, 30, 45, 60}
	isValid := false
	for _, v := range validValues {
		if value == v {
			isValid = true
			break
		}
	}

	if !isValid {
		sendMessage(chatID, "Invalid value. Allowed values: 15, 30, 45, 60")
		return nil
	}

	// Update setting
	err = ydb.UpdateUserSetting(ctx, user.ReviewerLogin, "cleanup_durations_minutes", value)
	if err != nil {
		sendMessage(chatID, fmt.Sprintf("Failed to update setting: %v", err))
		return nil
	}

	sendMessage(chatID, fmt.Sprintf("✅ Cleanup duration set to %d minutes", value))
	return nil
}

// HandleSetNotifyWhitelistTimeout handles the /set_notify_whitelist_timeout command
func HandleSetNotifyWhitelistTimeout(ctx context.Context, message *tba.Message, logger *log.Logger) error {
	return handleBooleanSetting(ctx, message, "notify_whitelist_timeout")
}

// HandleSetNotifyNonWhitelistCancel handles the /set_notify_non_whitelist_cancel command
func HandleSetNotifyNonWhitelistCancel(ctx context.Context, message *tba.Message, logger *log.Logger) error {
	return handleBooleanSetting(ctx, message, "notify_non_whitelist_cancel")
}

// HandleStatus handles the /status command - shows user status
func HandleStatus(ctx context.Context, message *tba.Message, logger *log.Logger) error {
	chatID := message.From.ID

	// Get user
	user, err := ydb.GetUserByTelegramChatID(ctx, chatID)
	if err != nil {
		sendMessage(chatID, "User not found. Please use /start to authenticate.")
		return nil
	}

	// Get recent review requests
	requests, err := ydb.GetReviewRequestsByUserAndStatus(ctx, user.ReviewerLogin, []string{
		models.StatusWaitingForApprove,
		models.StatusWhitelisted,
	})
	if err != nil {
		sendMessage(chatID, "Failed to retrieve status.")
		return nil
	}

	msg := fmt.Sprintf("*Status*\n\nUser: %s\nActive Reviews: %d",
		user.ReviewerLogin,
		len(requests))

	if len(requests) > 0 {
		msg += "\n\nRecent Reviews:"
		for _, req := range requests {
			projectName := "Unknown"
			if req.ProjectName != nil {
				projectName = *req.ProjectName
			}
			msg += fmt.Sprintf("\n- %s at %s", projectName, timeutil.FormatShort(timeutil.FromUnixSeconds(req.ReviewStartTime)))
		}
	}

	sendMessage(chatID, msg)
	return nil
}

// HandleUnknownCommand handles unrecognized commands
func HandleUnknownCommand(ctx context.Context, message *tba.Message, logger *log.Logger) error {
	sendMessage(message.Chat.ID, fmt.Sprintf("Unknown command: %s\n\nUse /help to see available commands.", message.Command()))
	return nil
}

// HandleAuthenticate handles login:password authentication
func HandleAuthenticate(ctx context.Context, message *tba.Message, logger *log.Logger) error {
	chatID := message.From.ID
	text := strings.TrimSpace(message.Text)

	// Parse login:password format
	parts := strings.SplitN(text, ":", 2)
	if len(parts) != 2 {
		sendMessage(chatID, "Invalid format. Please send your credentials in the format:\n\n`login:password`")
		return nil
	}

	login := strings.TrimSpace(parts[0])
	password := strings.TrimSpace(parts[1])

	// Check if user already exists
	existingUser, err := ydb.GetUserByTelegramChatID(ctx, chatID)
	if err == nil && existingUser != nil {
		sendMessage(chatID, fmt.Sprintf("You are already authenticated as %s.\n\nUse /logout first if you want to re-authenticate.", existingUser.ReviewerLogin))
		return nil
	}

	// Authenticate with s21 API
	tokenResp, err := external.Authenticate(ctx, login, password)
	if err != nil {
		logger.Printf("Authentication failed for user %d: %v", chatID, err)
		sendMessage(chatID, "Authentication failed. Please check your credentials and try again.")
		return nil
	}

	// Get user info from s21 to get the reviewer login
	// For now, use the login as reviewer_login
	// In production, you would fetch the actual username from the API
	reviewerLogin := login

	// Store tokens in Lockbox
	err = lockbox.StoreUserTokens(ctx, reviewerLogin, tokenResp.AccessToken, tokenResp.RefreshToken)
	if err != nil {
		logger.Printf("Failed to store tokens for %s: %v", reviewerLogin, err)
		sendMessage(chatID, "Authentication succeeded, but failed to store tokens. Please contact support.")
		return nil
	}

	// Create user record
	now := time.Now().Unix()
	user := &models.User{
		ReviewerLogin:     reviewerLogin,
		Status:            models.UserStatusActive,
		TelegramChatID:    chatID,
		CreatedAt:         now,
		LastAuthSuccessAt: now,
		LastAuthFailureAt: nil,
	}

	err = ydb.UpsertUser(ctx, user)
	if err != nil {
		logger.Printf("Failed to create user record for %s: %v", reviewerLogin, err)
		sendMessage(chatID, "Authentication succeeded, but failed to create user record. Please contact support.")
		return nil
	}

	// Create default settings
	err = ydb.CreateDefaultUserSettings(ctx, reviewerLogin)
	if err != nil {
		logger.Printf("Failed to create default settings for %s: %v", reviewerLogin, err)
		// Non-fatal, continue anyway
	}

	sendMessage(chatID, fmt.Sprintf("✅ Successfully authenticated as %s!\n\nYou can now use the bot. Use /help to see available commands.", reviewerLogin))
	return nil
}

// HandleLogout handles user logout
func HandleLogout(ctx context.Context, message *tba.Message, logger *log.Logger) error {
	chatID := message.From.ID

	// Get user
	user, err := ydb.GetUserByTelegramChatID(ctx, chatID)
	if err != nil {
		sendMessage(chatID, "You are not authenticated.")
		return nil
	}

	// Delete tokens from Lockbox
	err = lockbox.DeleteUserTokens(ctx, user.ReviewerLogin)
	if err != nil {
		logger.Printf("Failed to delete tokens for %s: %v", user.ReviewerLogin, err)
	}

	// Update user status to inactive
	err = ydb.UpdateUserStatus(ctx, user.ReviewerLogin, models.UserStatusInactive)
	if err != nil {
		logger.Printf("Failed to update user status for %s: %v", user.ReviewerLogin, err)
	}

	sendMessage(chatID, "✅ Logged out successfully. You can authenticate again with /start.")
	return nil
}

// HandleHelp displays help information
func HandleHelp(ctx context.Context, message *tba.Message, logger *log.Logger) error {
	chatID := message.From.ID

	helpText := `*Review Slot Guard Bot*

This bot helps you manage your review slots for School 21.

*Commands:*

/start - Start authentication
/logout - Log out from the bot
/status - Show your current status and active reviews
/settings - Display your current settings
/whitelist - Show your whitelisted projects and families

*Whitelist Management:*
/whitelist_add <family|project> <name> - Add to whitelist
/whitelist_remove <name> - Remove from whitelist

*Settings:*
/set_deadline_shift <minutes> - Response deadline shift (1-60)
/set_cancel_delay <minutes> - Non-whitelist cancel delay (1-10)
/set_slot_shift_threshold <minutes> - Slot shift threshold (5-60)
/set_slot_shift_duration <minutes> - Slot shift duration (5-60)
/set_cleanup_duration <minutes> - Cleanup duration (15, 30, 45, 60)
/set_notify_whitelist_timeout <true|false> - Notify on whitelist timeout
/set_notify_non_whitelist_cancel <true|false> - Notify on non-whitelist cancel`

	sendMessage(chatID, helpText)
	return nil
}

// Helper functions

func handleNumericSetting(ctx context.Context, message *tba.Message, field string, min, max, step int) error {
	chatID := message.From.ID

	// Get user
	user, err := ydb.GetUserByTelegramChatID(ctx, chatID)
	if err != nil {
		sendMessage(chatID, "User not found. Please use /start to authenticate.")
		return nil
	}

	arg := strings.TrimSpace(message.CommandArguments())
	value, err := strconv.Atoi(arg)
	if err != nil {
		sendMessage(chatID, fmt.Sprintf("Usage: /set_%s <value>\n\nValid range: %d - %d (step %d)", field, min, max, step))
		return nil
	}

	// Validate
	if value < min || value > max {
		sendMessage(chatID, fmt.Sprintf("Value must be between %d and %d", min, max))
		return nil
	}

	// Update setting
	err = ydb.UpdateUserSetting(ctx, user.ReviewerLogin, field, value)
	if err != nil {
		sendMessage(chatID, fmt.Sprintf("Failed to update setting: %v", err))
		return nil
	}

	sendMessage(chatID, fmt.Sprintf("✅ Setting updated to %d", value))
	return nil
}

func handleBooleanSetting(ctx context.Context, message *tba.Message, field string) error {
	chatID := message.From.ID

	// Get user
	user, err := ydb.GetUserByTelegramChatID(ctx, chatID)
	if err != nil {
		sendMessage(chatID, "User not found. Please use /start to authenticate.")
		return nil
	}

	arg := strings.ToLower(strings.TrimSpace(message.CommandArguments()))
	value := true

	if arg == "false" || arg == "no" || arg == "0" || arg == "off" {
		value = false
	}

	// Update setting
	err = ydb.UpdateUserSetting(ctx, user.ReviewerLogin, field, value)
	if err != nil {
		sendMessage(chatID, fmt.Sprintf("Failed to update setting: %v", err))
		return nil
	}

	sendMessage(chatID, fmt.Sprintf("✅ %s set to %t", field, value))
	return nil
}

func sendMessage(chatID int64, text string) {
	bot, _ := telegram.NewBotClientFromEnv()
	bot.SendPlainMessage(chatID, text)
}

func boolToYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func formatList(items []string) string {
	var result string
	for _, item := range items {
		result += fmt.Sprintf("  • %s\n", item)
	}
	return result
}
