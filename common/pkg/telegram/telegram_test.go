package telegram

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInlineKeyboardButton(t *testing.T) {
	button := InlineKeyboardButton{
		Text: "✅ Approve",
		Data: "APPROVE:123",
	}

	assert.Equal(t, "✅ Approve", button.Text)
	assert.Equal(t, "APPROVE:123", button.Data)
}

func TestFormatCallbackData(t *testing.T) {
	tests := []struct {
		name            string
		action          string
		reviewRequestID string
		expected        string
	}{
		{
			name:            "Approve callback",
			action:          "APPROVE",
			reviewRequestID: "550e8400-e29b-41d4-a716-446655440000",
			expected:        "APPROVE:550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:            "Decline callback",
			action:          "DECLINE",
			reviewRequestID: "550e8400-e29b-41d4-a716-446655440000",
			expected:        "DECLINE:550e8400-e29b-41d4-a716-446655440000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatCallbackData(tt.action, tt.reviewRequestID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseCallbackData(t *testing.T) {
	tests := []struct {
		name            string
		data            string
		expectedAction  string
		expectedID      string
		expectError     bool
	}{
		{
			name:            "Valid approve callback",
			data:            "APPROVE:550e8400-e29b-41d4-a716-446655440000",
			expectedAction:  "APPROVE",
			expectedID:      "550e8400-e29b-41d4-a716-446655440000",
			expectError:     false,
		},
		{
			name:            "Valid decline callback",
			data:            "DECLINE:550e8400-e29b-41d4-a716-446655440000",
			expectedAction:  "DECLINE",
			expectedID:      "550e8400-e29b-41d4-a716-446655440000",
			expectError:     false,
		},
		{
			name:        "Invalid format - missing action",
			data:        "550e8400-e29b-41d4-a716-446655440000",
			expectError: true,
		},
		{
			name:        "Invalid format - missing ID",
			data:        "APPROVE:",
			expectError: true,
		},
		{
			name:        "Invalid action",
			data:        "INVALID:550e8400-e29b-41d4-a716-446655440000",
			expectError: true,
		},
		{
			name:        "Empty string",
			data:        "",
			expectError: true,
		},
		{
			name:        "Multiple colons in ID",
			data:        "APPROVE:550e8400:e29b-41d4-a716-446655440000",
			expectedAction: "APPROVE",
			expectedID:   "550e8400:e29b-41d4-a716-446655440000",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, id, err := ParseCallbackData(tt.data)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedAction, action)
				assert.Equal(t, tt.expectedID, id)
			}
		})
	}
}

func TestSplitData(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		n        int
		expected []string
	}{
		{
			name:     "Split by colon",
			s:        "APPROVE:123",
			n:        2,
			expected: []string{"APPROVE", "123"},
		},
		{
			name:     "Split with multiple colons",
			s:        "APPROVE:123:extra",
			n:        2,
			expected: []string{"APPROVE", "123:extra"},
		},
		{
			name:     "No colon",
			s:        "APPROVE",
			n:        2,
			expected: []string{"APPROVE"},
		},
		{
			name:     "Empty string",
			s:        "",
			n:        2,
			expected: []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitData(tt.s, tt.n)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMessageConfig(t *testing.T) {
	config := MessageConfig{
		ChatID:    123456789,
		Text:      "Test message",
		ParseMode: "Markdown",
	}

	assert.Equal(t, int64(123456789), config.ChatID)
	assert.Equal(t, "Test message", config.Text)
	assert.Equal(t, "Markdown", config.ParseMode)
}

func TestEditMessageConfig(t *testing.T) {
	config := EditMessageConfig{
		ChatID:    123456789,
		MessageID: 123,
		Text:      "Edited message",
		ParseMode: "Markdown",
	}

	assert.Equal(t, int64(123456789), config.ChatID)
	assert.Equal(t, 123, config.MessageID)
	assert.Equal(t, "Edited message", config.Text)
}

func TestCallbackConfig(t *testing.T) {
	config := CallbackConfig{
		CallbackQueryID: "callback_123",
		Text:            "Review approved",
		ShowAlert:       false,
	}

	assert.Equal(t, "callback_123", config.CallbackQueryID)
	assert.Equal(t, "Review approved", config.Text)
	assert.False(t, config.ShowAlert)
}
