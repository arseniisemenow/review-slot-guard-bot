package handlers

import (
	"context"

	"github.com/arseniisemenow/review-slot-guard-bot-common/pkg/lockbox"
	"github.com/arseniisemenow/review-slot-guard-bot-common/pkg/telegram"
	"github.com/arseniisemenow/review-slot-guard-bot-common/pkg/ydb"
)

// Dependencies holds all external service interfaces for dependency injection
type Dependencies struct {
	Bot     telegram.BotSender
	DB      ydb.Database
	Lockbox lockbox.LockboxClient
}

// NewDependencies creates real dependencies for production use
func NewDependencies(ctx context.Context) (*Dependencies, error) {
	bot, err := telegram.NewBotClientFromEnv()
	if err != nil {
		return nil, err
	}

	db, err := ydb.NewYDBClient(ctx)
	if err != nil {
		return nil, err
	}

	lockboxClient := lockbox.NewClientAdapter()

	return &Dependencies{
		Bot:     bot,
		DB:      db,
		Lockbox: lockboxClient,
	}, nil
}

// NewTestDependencies creates mock dependencies for testing
func NewTestDependencies(mockBot *telegram.MockBotSender, mockDB *ydb.MockDatabase, mockLockbox *lockbox.MockLockboxClient) *Dependencies {
	return &Dependencies{
		Bot:     mockBot,
		DB:      mockDB,
		Lockbox: mockLockbox,
	}
}
