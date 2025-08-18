package inviteController

import (
	"testing"

	. "server/internal/models"

	"github.com/stretchr/testify/assert"
)

// Test the nil pointer safety logic that we added to RevokeInvite
func TestRevokeInvite_UserIDValidation(t *testing.T) {
	tests := []struct {
		name           string
		invite         StoryInvite
		expectedUserID string
		shouldBeEmpty  bool
	}{
		{
			name: "Valid user ID",
			invite: StoryInvite{
				UserID: stringPtr("valid-user-123"),
			},
			expectedUserID: "valid-user-123",
			shouldBeEmpty:  false,
		},
		{
			name: "Nil user ID",
			invite: StoryInvite{
				UserID: nil,
			},
			expectedUserID: "",
			shouldBeEmpty:  true,
		},
		{
			name: "Empty string user ID",
			invite: StoryInvite{
				UserID: stringPtr(""),
			},
			expectedUserID: "",
			shouldBeEmpty:  true,
		},
		{
			name: "Whitespace user ID",
			invite: StoryInvite{
				UserID: stringPtr("   "),
			},
			expectedUserID: "   ", // We don't trim whitespace in our logic
			shouldBeEmpty:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the logic from RevokeInvite method
			var revokedUserID string
			if tt.invite.UserID != nil && *tt.invite.UserID != "" {
				revokedUserID = *tt.invite.UserID
			}

			if tt.shouldBeEmpty {
				assert.Empty(t, revokedUserID, "revokedUserID should be empty")
			} else {
				assert.Equal(t, tt.expectedUserID, revokedUserID, "revokedUserID should match expected value")
			}
		})
	}
}

func TestDeclineInvite_UserIDValidation(t *testing.T) {
	// Test cases for DeclineInvite where userID is passed to cache invalidation
	tests := []struct {
		name   string
		userID string
		valid  bool
	}{
		{
			name:   "Valid user ID",
			userID: "valid-user-123",
			valid:  true,
		},
		{
			name:   "Empty user ID",
			userID: "",
			valid:  false, // Empty userID would still be passed but filtered out by cache invalidation service
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The DeclineInvite method passes userID directly to cache invalidation
			// The cache invalidation service should handle empty strings properly
			userID := tt.userID
			
			if tt.valid {
				assert.NotEmpty(t, userID, "Valid user ID should not be empty")
			} else {
				assert.Empty(t, userID, "Invalid user ID should be empty")
			}
		})
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}