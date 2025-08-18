package inviteController

import (
	"context"
	"server/internal/logger"
	. "server/internal/models"
	"server/internal/repositories"
	"server/internal/services"
)

type InviteController struct {
	inviteRepo               repositories.InviteRepository
	storyRepo                repositories.StoryRepository
	transactionService       *services.TransactionService
	cacheInvalidationService *services.CacheInvalidationService
	log                      logger.Logger
}

func New(
	inviteRepo repositories.InviteRepository,
	storyRepo repositories.StoryRepository,
	transactionService *services.TransactionService,
	cacheInvalidationService *services.CacheInvalidationService,
) *InviteController {
	return &InviteController{
		inviteRepo:               inviteRepo,
		storyRepo:                storyRepo,
		transactionService:       transactionService,
		cacheInvalidationService: cacheInvalidationService,
		log:                      logger.New("InviteController"),
	}
}

func (ic *InviteController) GetKnownContacts(ctx context.Context, userID string) ([]User, error) {
	contacts, err := ic.inviteRepo.GetKnownContacts(ctx, userID)
	if err != nil {
		return nil, ic.log.Function("GetKnownContacts").
			Err("failed to get known contacts", err, "userID", userID)
	}

	return contacts, nil
}

func (ic *InviteController) CreateInvitesByUserIDs(
	ctx context.Context,
	storyID, inviterID string,
	userIDs []string,
) error {
	log := ic.log.Function("CreateInvitesByUserIDs")

	err := ic.transactionService.Execute(ctx, func(txCtx context.Context) error {
		var invites []StoryInvite
		for _, userID := range userIDs {
			invites = append(invites, StoryInvite{
				StoryID:   storyID,
				UserID:    &userID,
				InviterID: inviterID,
				Status:    "pending",
			})
		}

		if err := ic.inviteRepo.CreateStoryInvitesBatch(txCtx, invites); err != nil {
			return log.Err("failed to create story invites batch", err, "invites", invites)
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Invalidate cache and notify clients
	if _, err := ic.cacheInvalidationService.InvalidateStoryCache(ctx, storyID); err != nil {
		// Log error but don't fail the operation - invites were created successfully
		log.Warn(
			"Failed to invalidate story cache after creating invites",
			"storyID",
			storyID,
			"error",
			err,
		)
	}

	return nil
}

func (ic *InviteController) CreateInvites(
	ctx context.Context,
	storyID, inviterID string,
	emails []string,
) error {
	log := ic.log.Function("CreateInvites")

	var invitedUserIDs []string

	err := ic.transactionService.Execute(ctx, func(txCtx context.Context) error {
		users, err := ic.inviteRepo.GetUsersByEmail(ctx, emails)
		if err != nil {
			return log.Err("failed to get users by email", err, "emails", emails)
		}

		var invites []StoryInvite
		for _, user := range users {
			invites = append(invites, StoryInvite{
				StoryID:   storyID,
				UserID:    &user.ID,
				Email:     user.Email,
				InviterID: inviterID,
				Status:    "pending",
			})
			invitedUserIDs = append(invitedUserIDs, user.ID)
		}

		if err := ic.inviteRepo.CreateStoryInvitesBatch(txCtx, invites); err != nil {
			return log.Err("failed to create email invites batch", err, "invites", invites)
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Invalidate cache and notify clients
	if _, err := ic.cacheInvalidationService.InvalidateStoryCache(ctx, storyID); err != nil {
		// Log error but don't fail the operation - invites were created successfully
		log.Warn(
			"Failed to invalidate story cache after creating invites",
			"storyID",
			storyID,
			"error",
			err,
		)
	}

	return nil
}

// TODO: Validate this functionality
func (ic *InviteController) AcceptInvite(ctx context.Context, inviteID, userID string) error {
	log := ic.log.Function("AcceptInvite")

	invite, err := ic.inviteRepo.GetStoryInvite(ctx, inviteID)
	if err != nil {
		return log.Err("failed to get story invite", err, "inviteID", inviteID)
	}

	if invite.UserID == nil || *invite.UserID != userID {
		return log.Error("invitation not for this user", "inviteID", inviteID, "userID", userID)
	}

	if invite.Status != "pending" {
		return log.Error("invitation is not pending", "inviteID", inviteID, "status", invite.Status)
	}

	err = ic.transactionService.Execute(ctx, func(txCtx context.Context) error {
		// Create story author
		storyAuthor := &StoryAuthor{
			StoryID:  invite.StoryID,
			AuthorID: userID,
		}
		if err := ic.storyRepo.CreateStoryAuthor(txCtx, storyAuthor); err != nil {
			return log.Err("failed to create story author", err, "storyAuthor", storyAuthor)
		}

		// Update invite status
		invite.Status = "accepted"
		if err := ic.inviteRepo.UpdateStoryInvite(txCtx, invite); err != nil {
			return log.Err("failed to update story invite", err, "invite", invite)
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Invalidate cache for all users with access to the story (including the newly added author)
	if _, err := ic.cacheInvalidationService.InvalidateStoryCache(ctx, invite.StoryID); err != nil {
		// Log error but don't fail the operation - invite was accepted successfully
		log.Warn(
			"Failed to invalidate story cache after accepting invite",
			"storyID",
			invite.StoryID,
			"error",
			err,
		)
	}

	return nil
}

// TODO: Validate this functionality
func (ic *InviteController) DeclineInvite(ctx context.Context, inviteID, userID string) error {
	log := ic.log.Function("DeclineInvite")

	invite, err := ic.inviteRepo.GetStoryInvite(ctx, inviteID)
	if err != nil {
		return log.Err("failed to get story invite", err, "inviteID", inviteID)
	}

	if invite.UserID == nil || *invite.UserID != userID {
		return log.Error("invitation not for this user", "inviteID", inviteID, "userID", userID)
	}

	if invite.Status != "pending" {
		return log.Error("invitation is not pending", "inviteID", inviteID, "status", invite.Status)
	}

	invite.Status = "declined"
	err = ic.inviteRepo.UpdateStoryInvite(ctx, invite)
	if err != nil {
		return err
	}

	// Invalidate cache for all users with access to the story AND the declining user
	if _, err := ic.cacheInvalidationService.InvalidateStoryCache(ctx, invite.StoryID, userID); err != nil {
		// Log error but don't fail the operation - invite was declined successfully
		log.Warn(
			"Failed to invalidate story cache after declining invite",
			"storyID",
			invite.StoryID,
			"error",
			err,
		)
	}

	return nil
}

func (ic *InviteController) RevokeInvite(ctx context.Context, inviteID, revokerID string) error {
	log := ic.log.Function("RevokeInvite")

	invite, err := ic.inviteRepo.GetStoryInvite(ctx, inviteID)
	if err != nil {
		return log.Err("failed to get story invite", err, "inviteID", inviteID)
	}

	story, err := ic.storyRepo.GetStory(ctx, invite.StoryID)
	if err != nil {
		return log.Err("failed to get story", err, "storyID", invite.StoryID)
	}

	if story.OwnerID != revokerID && invite.InviterID != revokerID {
		return log.Error(
			"user not authorized to revoke invite",
			"inviteID",
			inviteID,
			"revokerID",
			revokerID,
		)
	}

	var revokedUserID string
	if invite.UserID != nil && *invite.UserID != "" {
		revokedUserID = *invite.UserID
	}

	err = ic.inviteRepo.DeleteStoryInvite(ctx, inviteID)
	if err != nil {
		return log.Err("failed to delete story invite", err, "inviteID", inviteID)
	}

	// Only include revokedUserID in cache invalidation if it's not empty
	if revokedUserID != "" {
		if _, err := ic.cacheInvalidationService.InvalidateStoryCache(ctx, invite.StoryID, revokedUserID); err != nil {
			log.Warn(
				"Failed to invalidate story cache after revoking invite",
				"storyID",
				invite.StoryID,
				"error",
				err,
			)
		}
	} else {
		// If no specific user to notify, just invalidate for story users
		if _, err := ic.cacheInvalidationService.InvalidateStoryCache(ctx, invite.StoryID); err != nil {
			log.Warn(
				"Failed to invalidate story cache after revoking invite",
				"storyID",
				invite.StoryID,
				"error",
				err,
			)
		}
	}

	return nil
}
