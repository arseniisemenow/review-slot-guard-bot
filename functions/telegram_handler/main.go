package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	tba "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/arseniisemenow/review-slot-guard-bot/common/pkg/telegram"
	"github.com/arseniisemenow/review-slot-guard-bot/common/pkg/ydb"
	"github.com/arseniisemenow/review-slot-guard-bot/functions/telegram_handler/internal/handlers"
)

var (
	deps *handlers.Dependencies
)

// init initializes dependencies
func init() {
	ctx := context.Background()
	var err error

	deps, err = handlers.NewDependencies(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize dependencies: %v", err)
	}
}

// main function for local testing
func main() {
	http.HandleFunc("/", Handler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// Handler is the Yandex Cloud Function entry point for Telegram webhooks
func Handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := log.New(os.Stdout, "[TELEGRAM_HANDLER] ", log.LstdFlags)

	// Parse incoming update
	var update tba.Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		logger.Printf("Failed to decode update: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Process the update
	if err := processUpdate(ctx, &update, logger); err != nil {
		logger.Printf("Error processing update: %v", err)
		// Still return OK to Telegram to avoid retries
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// processUpdate handles an incoming Telegram update
func processUpdate(ctx context.Context, update *tba.Update, logger *log.Logger) error {
	// Handle callback queries (button clicks)
	if update.CallbackQuery != nil {
		return handleCallbackQuery(ctx, update.CallbackQuery, logger)
	}

	// Handle messages
	if update.Message != nil {
		return handleMessage(ctx, update.Message, logger)
	}

	return nil
}

// handleCallbackQuery handles button callback queries
func handleCallbackQuery(ctx context.Context, callback *tba.CallbackQuery, logger *log.Logger) error {
	logger.Printf("Received callback query from user %d", callback.From.ID)

	// Get user by telegram_chat_id
	user, err := ydb.GetUserByTelegramChatID(ctx, callback.From.ID)
	if err != nil {
		logger.Printf("User not found for telegram_chat_id %d: %v", callback.From.ID, err)
		// Answer the callback anyway
		deps.Bot.AnswerCallbackQuery(callback.ID, "User not found. Please use /start to authenticate.")
		return nil
	}

	// Parse callback data
	action, reviewRequestID, err := telegram.ParseCallbackData(callback.Data)
	if err != nil {
		logger.Printf("Failed to parse callback data %s: %v", callback.Data, err)
		deps.Bot.AnswerCallbackQuery(callback.ID, "Invalid callback data")
		return nil
	}

	// Get review request
	req, err := ydb.GetReviewRequestByID(ctx, reviewRequestID)
	if err != nil {
		logger.Printf("Review request not found: %s", reviewRequestID)
		deps.Bot.AnswerCallbackQuery(callback.ID, "Review request not found")
		return nil
	}

	// Verify the review belongs to the user
	if req.ReviewerLogin != user.ReviewerLogin {
		logger.Printf("User %s attempted to access review %s belonging to %s", user.ReviewerLogin, reviewRequestID, req.ReviewerLogin)
		deps.Bot.AnswerCallbackQuery(callback.ID, "Access denied")
		return nil
	}

	// Handle the action
	switch action {
	case "APPROVE":
		return handlers.HandleApprove(ctx, deps, user, req, callback, logger)

	case "DECLINE":
		return handlers.HandleDecline(ctx, deps, user, req, callback, logger)

	default:
		logger.Printf("Unknown action: %s", action)
		deps.Bot.AnswerCallbackQuery(callback.ID, "Unknown action")
	}

	return nil
}

// handleMessage handles incoming messages
func handleMessage(ctx context.Context, message *tba.Message, logger *log.Logger) error {
	// Only process text messages
	if message.Text == "" {
		return nil
	}

	logger.Printf("Received message from user %d: %s", message.From.ID, message.Text)

	// Handle commands
	if message.IsCommand() {
		return handleCommand(ctx, message, logger)
	}

	// Handle non-command text messages (like login:password)
	return handlers.HandleAuthenticate(ctx, deps, message, logger)
}

// handleCommand handles Telegram bot commands
func handleCommand(ctx context.Context, message *tba.Message, logger *log.Logger) error {
	command := message.Command()

	switch command {
	case "start":
		return handlers.HandleStart(ctx, deps, message, logger)

	case "help":
		return handlers.HandleHelp(ctx, deps, message, logger)

	case "logout":
		return handlers.HandleLogout(ctx, deps, message, logger)

	case "settings":
		return handlers.HandleSettings(ctx, deps, message, logger)

	case "whitelist":
		return handlers.HandleWhitelist(ctx, deps, message, logger)

	case "whitelist_add":
		return handlers.HandleWhitelistAdd(ctx, deps, message, logger)

	case "whitelist_remove":
		return handlers.HandleWhitelistRemove(ctx, deps, message, logger)

	case "set_deadline_shift":
		return handlers.HandleSetDeadlineShift(ctx, deps, message, logger)

	case "set_cancel_delay":
		return handlers.HandleSetCancelDelay(ctx, deps, message, logger)

	case "set_slot_shift_threshold":
		return handlers.HandleSetSlotShiftThreshold(ctx, deps, message, logger)

	case "set_slot_shift_duration":
		return handlers.HandleSetSlotShiftDuration(ctx, deps, message, logger)

	case "set_cleanup_duration":
		return handlers.HandleSetCleanupDuration(ctx, deps, message, logger)

	case "set_notify_whitelist_timeout":
		return handlers.HandleSetNotifyWhitelistTimeout(ctx, deps, message, logger)

	case "set_notify_non_whitelist_cancel":
		return handlers.HandleSetNotifyNonWhitelistCancel(ctx, deps, message, logger)

	case "status":
		return handlers.HandleStatus(ctx, deps, message, logger)

	default:
		return handlers.HandleUnknownCommand(ctx, deps, message, logger)
	}
}
