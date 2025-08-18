package sectionController

import (
	"context"
	"server/internal/logger"
	. "server/internal/models"
	"server/internal/repositories"
	"server/internal/services"
	"time"
)

type SectionController struct {
	sectionRepo              repositories.SectionRepository
	storyRepo                repositories.StoryRepository
	transactionService       *services.TransactionService
	cacheInvalidationService *services.CacheInvalidationService
	log                      logger.Logger
}

func New(
	sectionRepo repositories.SectionRepository,
	storyRepo repositories.StoryRepository,
	transactionService *services.TransactionService,
	cacheInvalidationService *services.CacheInvalidationService,
) *SectionController {
	return &SectionController{
		sectionRepo:              sectionRepo,
		storyRepo:                storyRepo,
		transactionService:       transactionService,
		cacheInvalidationService: cacheInvalidationService,
		log:                      logger.New("SectionController"),
	}
}

func (sc *SectionController) Create(
	ctx context.Context,
	userID string,
	storyID string,
	content string,
) (Story, error) {
	log := sc.log.Function("Create")

	var result Story

	err := sc.transactionService.Execute(ctx, func(txCtx context.Context) error {
		story, err := sc.storyRepo.GetStory(txCtx, storyID)
		if err != nil {
			return log.Err("story not found", err, "storyID", storyID)
		}

		if userID == "" || storyID == "" || content == "" {
			return log.Error("invalid section data", "userID", userID, "storyID", storyID)
		}

		section := &StoryContent{
			Content:  content,
			AuthorID: userID,
			StoryID:  storyID,
		}

		if err := sc.sectionRepo.Create(txCtx, section); err != nil {
			return log.Err("failed to create section", err)
		}

		// TODO: Update the story, feels like this should be a hook on the section model
		story.UpdatedAt = time.Now()

		// if err := sc.storyRepo.Update(txCtx, story); err != nil {
		//   return log.Err("failed to update story", err)
		// }

		updatedStory, err := sc.storyRepo.GetStoryFromDB(txCtx, storyID)
		if err != nil {
			return log.Err("failed to get updated story", err)
		}

		result = updatedStory
		return nil
	})
	if err != nil {
		return Story{}, err
	}

	// Invalidate cache and notify clients
	if _, err := sc.cacheInvalidationService.InvalidateStoryCache(ctx, storyID); err != nil {
		log.Warn("Failed to invalidate story cache after section create", "storyID", storyID, "error", err)
	}

	return result, nil
}


func (sc *SectionController) Delete(
	ctx context.Context,
	sectionID string,
	storyID string,
) (Story, error) {
	log := sc.log.Function("Delete")

	var story Story
	err := sc.transactionService.Execute(ctx, func(txCtx context.Context) error {
		if err := sc.sectionRepo.Delete(txCtx, sectionID); err != nil {
			return log.Err("failed to delete section", err, "sectionID", sectionID)
		}

		// Get story
		var err error
		story, err = sc.storyRepo.GetStoryFromDB(txCtx, storyID)
		if err != nil {
			return log.Err("failed to get story", err, "storyID", storyID)
		}

		return nil
	})
	if err != nil {
		return Story{}, err
	}

	log.Info("Successfully deleted section", "sectionID", sectionID)

	// Invalidate cache and notify clients
	if _, err := sc.cacheInvalidationService.InvalidateStoryCache(ctx, storyID); err != nil {
		log.Warn("Failed to invalidate story cache after section delete", "storyID", storyID, "error", err)
	}

	return story, nil
}

func (sc *SectionController) Update(
	ctx context.Context,
	sectionID string,
	content string,
	userID string,
) (Story, error) {
	log := sc.log.Function("Update")

	if sectionID == "" || content == "" || userID == "" {
		return Story{}, log.Error(
			"invalid parameters",
			"sectionID",
			sectionID,
			"userID",
			userID,
			"contentEmpty",
			content == "",
		)
	}

	var story Story
	err := sc.transactionService.Execute(ctx, func(txCtx context.Context) error {
		existingSection, err := sc.sectionRepo.GetByIDAndAuthor(txCtx, sectionID, userID)
		if err != nil {
			return log.Err(
				"authorization check failed",
				err,
				"sectionID",
				sectionID,
				"userID",
				userID,
			)
		}

		if err := sc.sectionRepo.UpdateContent(txCtx, sectionID, content); err != nil {
			return log.Err("failed to update section", err, "sectionID", sectionID)
		}

		story, err = sc.storyRepo.GetStoryFromDB(txCtx, existingSection.StoryID)
		if err != nil {
			return log.Err("failed to get story", err, "storyID", existingSection.StoryID)
		}

		return nil
	})
	if err != nil {
		return Story{}, err
	}

	// Invalidate cache and notify clients
	if _, err := sc.cacheInvalidationService.InvalidateStoryCache(ctx, story.ID); err != nil {
		log.Warn("Failed to invalidate story cache after section update", "storyID", story.ID, "error", err)
	}

	return story, nil
}
