package storyController

import (
	"context"
	"server/internal/logger"
	. "server/internal/models"
	"server/internal/repositories"
	"server/internal/services"
)

type StoryController struct {
	storyRepo          repositories.StoryRepository
	inviteRepo         repositories.InviteRepository
	transactionService *services.TransactionService
	log                logger.Logger
}

func New(
	storyRepo repositories.StoryRepository,
	inviteRepo repositories.InviteRepository,
	transactionService *services.TransactionService,
) *StoryController {
	return &StoryController{
		storyRepo:          storyRepo,
		inviteRepo:         inviteRepo,
		transactionService: transactionService,
		log:                logger.New("StoryController"),
	}
}

func (sc *StoryController) GetStoriesByUserID(
	ctx context.Context,
	userID string,
) ([]Story, []StoryInvite, error) {
	log := sc.log.Function("GetStoriesByUserID")

	if userID == "" {
		return nil, nil, log.Error("invalid user ID", "userID", userID)
	}

	invites, err := sc.inviteRepo.GetUserInvites(ctx, userID)
	if err != nil {
		return nil, nil, log.Err("failed to get user invites", err, "userID", userID)
	}

	storyIDs, found, err := sc.storyRepo.GetUserStoryIDsFromCache(ctx, userID)
	if err != nil {
		return nil, nil, log.Err("failed to get user story IDs from cache", err, "userID", userID)
	}

	if !found || len(storyIDs) == 0 {
		log.Info("Story IDs not found in cache, fetching from DB", "userID", userID)
		stories, err := sc.storyRepo.GetStoriesFromDB(ctx, userID)
		if err != nil {
			return nil, nil, log.Err("failed to get stories from database", err, "userID", userID)
		}

		if err := sc.storyRepo.SetUserStoryIDsToCache(ctx, stories, userID); err != nil {
			log.Warn("failed to set user story IDs to cache", "userID", userID, "error", err)
		}

		return stories, invites, nil
	}

	stories, found, err := sc.storyRepo.GetStoriesFromCache(ctx, storyIDs)
	if err != nil {
		return nil, nil, log.Err("failed to get stories from cache", err, "userID", userID)
	}

	if found && len(stories) == len(storyIDs) {
		log.Debug("Found all stories in cache", "userID", userID, "count", len(stories))
		return stories, invites, nil
	}

	stories, err = sc.storyRepo.GetStoriesFromDB(ctx, userID)
	if err != nil {
		return nil, nil, log.Err("failed to get stories from database", err, "userID", userID)
	}

	sc.storyRepo.SetStoriesToCache(stories)

	return stories, invites, nil
}

func (sc *StoryController) GetCreateFormData(
	ctx context.Context,
) (*repositories.CreateFormData, error) {
	log := sc.log.Function("GetCreateFormData")

	formData, found, err := sc.storyRepo.GetCreateFormDataFromCache(ctx)
	if err != nil {
		log.Warn("failed to get create form data from cache", "error", err)
	}

	if found {
		log.Debug("Found create form data in cache")
		return formData, nil
	}

	log.Debug("Create form data not found in cache, fetching from DB")

	formData, err = sc.storyRepo.GetCreateFormDataFromDB(ctx)
	if err != nil {
		return nil, log.Err("failed to get create form data from database", err)
	}

	if err := sc.storyRepo.SetCreateFormDataToCache(ctx, formData); err != nil {
		log.Warn("failed to set create form data to cache", "error", err)
	}

	return formData, nil
}

func (sc *StoryController) GetStory(ctx context.Context, id string) (Story, error) {
	log := sc.log.Function("GetStory")

	story, err := sc.storyRepo.GetStory(ctx, id)
	if err != nil {
		return Story{}, log.Err("failed to get story", err, "id", id)
	}

	return story, nil
}

func (sc *StoryController) CreateStory(ctx context.Context, story *Story) error {
	log := sc.log.Function("CreateStory")

	if story.Title == "" {
		return log.Error("story title is required", "story", story)
	}

	// Set the owner as the initial NextUp
	story.NextUpID = &story.OwnerID

	return sc.transactionService.Execute(ctx, func(txCtx context.Context) error {
		if err := sc.storyRepo.CreateStory(txCtx, story); err != nil {
			return log.Err("failed to create story", err, "story", story)
		}

		storyAuthor := &StoryAuthor{
			StoryID:  story.ID,
			AuthorID: story.OwnerID,
			// Order will be set automatically by BeforeCreate hook
		}
		if err := sc.storyRepo.CreateStoryAuthor(txCtx, storyAuthor); err != nil {
			return log.Err(
				"failed to create story author for owner",
				err,
				"storyAuthor",
				storyAuthor,
			)
		}

		// Set the owner as the initial nextUp for all order types
		story.NextUpID = &story.OwnerID
		if err := sc.storyRepo.UpdateStory(txCtx, story); err != nil {
			return log.Err("failed to set initial nextUp", err, "story", story)
		}

		return nil
	})
}


func (sc *StoryController) GetStoryMembers(ctx context.Context, storyID string) ([]User, error) {
	log := sc.log.Function("GetStoryMembers")

	members, err := sc.storyRepo.GetStoryMembers(ctx, storyID)
	if err != nil {
		return nil, log.Err("failed to get story members", err, "storyID", storyID)
	}

	return members, nil
}

func (sc *StoryController) RemoveStoryMember(
	ctx context.Context,
	storyID, userID, removerID string,
) error {
	log := sc.log.Function("RemoveStoryMember")

	story, err := sc.storyRepo.GetStory(ctx, storyID)
	if err != nil {
		return log.Err("failed to get story", err, "storyID", storyID)
	}

	// Only owner can remove members
	if story.OwnerID != removerID {
		return log.Error(
			"only owner can remove members",
			"storyID",
			storyID,
			"removerID",
			removerID,
		)
	}

	// Cannot remove owner
	if story.OwnerID == userID {
		return log.Error("cannot remove story owner", "storyID", storyID, "userID", userID)
	}

	return sc.storyRepo.RemoveStoryMember(ctx, storyID, userID)
}
