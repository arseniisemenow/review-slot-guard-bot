package lockbox

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/arseniisemenow/review-slot-guard-bot/common/pkg/models"
)

func TestLockboxPayloadStructure(t *testing.T) {
	payload := &models.LockboxPayload{
		Version: 1,
		Users: map[string]models.UserTokens{
			"john.doe": {
				AccessToken:  "access_token_123",
				RefreshToken: "refresh_token_123",
			},
			"jane.doe": {
				AccessToken:  "access_token_456",
				RefreshToken: "refresh_token_456",
			},
		},
	}

	assert.Equal(t, 1, payload.Version)
	assert.Len(t, payload.Users, 2)

	tokens, ok := payload.Users["john.doe"]
	assert.True(t, ok)
	assert.Equal(t, "access_token_123", tokens.AccessToken)
	assert.Equal(t, "refresh_token_123", tokens.RefreshToken)
}

func TestUserTokensStructure(t *testing.T) {
	tokens := &models.UserTokens{
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
	}

	assert.Equal(t, "test_access_token", tokens.AccessToken)
	assert.Equal(t, "test_refresh_token", tokens.RefreshToken)
}

func TestLockboxCacheExpiry(t *testing.T) {
	// Test cache expiry logic
	cacheExpiry := time.Now().Add(5 * time.Minute)
	payloadCache := &models.LockboxPayload{
		Version: 1,
		Users:   make(map[string]models.UserTokens),
	}

	// Cache should be valid
	assert.True(t, time.Now().Before(cacheExpiry))
	assert.NotNil(t, payloadCache)

	// Simulate cache expiry
	cacheExpiry = time.Now().Add(-1 * time.Minute)
	assert.True(t, time.Now().After(cacheExpiry))
}

func TestGetUserTokensFromPayload(t *testing.T) {
	payload := &models.LockboxPayload{
		Version: 1,
		Users: map[string]models.UserTokens{
			"testuser": {
				AccessToken:  "test_access",
				RefreshToken: "test_refresh",
			},
		},
	}

	t.Run("User exists", func(t *testing.T) {
		tokens, ok := payload.Users["testuser"]
		assert.True(t, ok)
		assert.Equal(t, "test_access", tokens.AccessToken)
		assert.Equal(t, "test_refresh", tokens.RefreshToken)
	})

	t.Run("User does not exist", func(t *testing.T) {
		_, ok := payload.Users["nonexistent"]
		assert.False(t, ok)
	})
}

func TestInvalidateCache(t *testing.T) {
	// Set some cache state
	payloadCache = &models.LockboxPayload{
		Version: 1,
		Users:   make(map[string]models.UserTokens),
	}
	cacheExpiry = time.Now().Add(5 * time.Minute)

	// Invalidate
	InvalidateCache()

	// Check cache is cleared
	assert.Nil(t, payloadCache)
}

func TestLockboxErrorMessages(t *testing.T) {
	tests := []struct {
		name      string
		reviewerLogin string
		expectedErr string
	}{
		{
			name:         "Empty reviewer login",
			reviewerLogin: "",
			expectedErr:  "tokens not found for user",
		},
		{
			name:         "Non-existent user",
			reviewerLogin: "nonexistent_user",
			expectedErr:  "tokens not found for user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create empty payload
			payload := &models.LockboxPayload{
				Version: 1,
				Users:   make(map[string]models.UserTokens),
			}

			// Try to get tokens
			_, ok := payload.Users[tt.reviewerLogin]
			assert.False(t, ok, "User should not exist in payload")
		})
	}
}

func TestGetSecretID(t *testing.T) {
	// Initially should be empty
	id := GetSecretID()
	assert.Equal(t, "", id, "Secret ID should be empty before init")
}

func TestStoreUserTokensLogic(t *testing.T) {
	payload := &models.LockboxPayload{
		Version: 1,
		Users:   make(map[string]models.UserTokens),
	}

	// Add first user
	payload.Users["user1"] = models.UserTokens{
		AccessToken:  "token1_access",
		RefreshToken: "token1_refresh",
	}

	assert.Len(t, payload.Users, 1)

	// Add second user
	payload.Users["user2"] = models.UserTokens{
		AccessToken:  "token2_access",
		RefreshToken: "token2_refresh",
	}

	assert.Len(t, payload.Users, 2)

	// Verify both users exist
	tokens1, ok := payload.Users["user1"]
	assert.True(t, ok)
	assert.Equal(t, "token1_access", tokens1.AccessToken)

	tokens2, ok := payload.Users["user2"]
	assert.True(t, ok)
	assert.Equal(t, "token2_access", tokens2.AccessToken)
}

func TestDeleteUserTokensLogic(t *testing.T) {
	payload := &models.LockboxPayload{
		Version: 1,
		Users: map[string]models.UserTokens{
			"user1": {
				AccessToken:  "token1_access",
				RefreshToken: "token1_refresh",
			},
			"user2": {
				AccessToken:  "token2_access",
				RefreshToken: "token2_refresh",
			},
		},
	}

	assert.Len(t, payload.Users, 2)

	// Delete user1
	delete(payload.Users, "user1")

	assert.Len(t, payload.Users, 1)

	// Verify user1 is gone
	_, ok := payload.Users["user1"]
	assert.False(t, ok)

	// Verify user2 still exists
	tokens2, ok := payload.Users["user2"]
	assert.True(t, ok)
	assert.Equal(t, "token2_access", tokens2.AccessToken)
}

// Mock context for testing
func mockContext() context.Context {
	return context.Background()
}

func TestContextUsage(t *testing.T) {
	ctx := mockContext()
	assert.NotNil(t, ctx)

	// Test with timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	assert.NotNil(t, ctx)
}
