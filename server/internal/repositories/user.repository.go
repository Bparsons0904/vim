package repositories

import (
	"context"
	"server/config"
	"server/internal/database"
	"server/internal/logger"
	. "server/internal/models"
	"server/internal/services"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	USER_CACHE_EXPIRY = 7 * 24 * time.Hour // 7 days
)

type UserRepository interface {
	GetByID(ctx context.Context, id string) (*User, error)
	GetByLogin(ctx context.Context, login string) (*User, error)
	Create(ctx context.Context, user *User, config config.Config) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
}

type userRepository struct {
	db  database.DB
	log logger.Logger
}

func New(db database.DB) UserRepository {
	return &userRepository{
		db:  db,
		log: logger.New("userRepository"),
	}
}

func (r *userRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := services.GetTransaction(ctx); ok {
		return tx
	}
	return r.db.SQLWithContext(ctx)
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*User, error) {
	log := r.log.Function("GetByID")

	var user User
	if err := r.getCacheByID(ctx, id, &user); err == nil {
		return &user, nil
	}

	if err := r.getDBByID(ctx, id, &user); err != nil {
		return nil, err
	}

	if err := r.addUserToCache(ctx, &user); err != nil {
		log.Warn("failed to add user to cache", "userID", id, "error", err)
	}

	return &user, nil
}

func (r *userRepository) GetByLogin(ctx context.Context, login string) (*User, error) {
	log := r.log.Function("GetByLogin")

	var user User
	if err := r.getDBByLogin(ctx, login, &user); err != nil {
		return nil, log.Err("failed to get user by login", err, "login", login)
	}

	if err := r.addUserToCache(ctx, &user); err != nil {
		log.Warn("failed to add user to cache", "userID", user.ID, "error", err)
	}

	return &user, nil
}

func (r *userRepository) Create(
	ctx context.Context,
	user *User,
	config config.Config,
) error {
	log := r.log.Function("Create")

	if err := r.getDB(ctx).Create(user).Error; err != nil {
		return log.Err("failed to create user", err, "user", user)
	}

	return nil
}

func (r *userRepository) Update(ctx context.Context, user *User) error {
	log := r.log.Function("Update")

	if err := r.getDB(ctx).Save(user).Error; err != nil {
		return log.Err("failed to update user", err, "user", user)
	}

	if err := r.addUserToCache(ctx, user); err != nil {
		log.Warn("failed to update user in cache", "userID", user.ID, "error", err)
	}

	return nil
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	log := r.log.Function("Delete")

	if err := r.getDB(ctx).Delete(&User{}, "id = ?", id).Error; err != nil {
		return log.Err("failed to delete user", err, "id", id)
	}

	if err := database.NewCacheBuilder(r.db.Cache.User, id).Delete(); err != nil {
		log.Warn("failed to remove user from cache", "userID", id, "error", err)
	}

	return nil
}

func (r *userRepository) getCacheByID(ctx context.Context, userID string, user *User) error {
	found, err := database.NewCacheBuilder(r.db.Cache.User, userID).Get(user)
	if err != nil {
		return r.log.Function("getCacheByID").
			Err("failed to get user from cache", err, "userID", userID)
	}

	if !found {
		return r.log.Function("getCacheByID").
			Error("user not found in cache", "userID", userID)
	}

	return nil
}

func (r *userRepository) addUserToCache(ctx context.Context, user *User) error {
	if err := database.NewCacheBuilder(r.db.Cache.User, user.ID).
		WithStruct(user).
		WithTTL(USER_CACHE_EXPIRY).
		WithContext(ctx).
		Set(); err != nil {
		return r.log.Function("addUserToCache").
			Err("failed to add user to cache", err, "user", user)
	}
	return nil
}

func (r *userRepository) getDBByID(ctx context.Context, userID string, user *User) error {
	log := r.log.Function("getDBByID")

	id, err := uuid.Parse(userID)
	if err != nil {
		return log.Err("failed to parse userID", err, "userID", userID)
	}

	if err := r.getDB(ctx).First(user, "id = ?", id).Error; err != nil {
		return log.Err("failed to get user by id", err, "id", userID)
	}

	return nil
}

func (r *userRepository) getDBByLogin(ctx context.Context, login string, user *User) error {
	if err := r.getDB(ctx).First(user, "login = ?", login).Error; err != nil {
		return r.log.Function("getDBByLogin").
			Err("failed to get user by login", err, "login", login)
	}
	return nil
}
