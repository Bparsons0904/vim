package database

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/valkey-io/valkey-go"
)

type CacheItem[T any] struct {
	Cache       valkey.Client
	HashPattern *string // "hash:%s"
	Key         any
	Value       T
	Expiry      *time.Duration
}

type DeleteCacheItem[T any] struct {
	Cache       valkey.Client
	HashPattern *string
	Key         any
}

type KeyType interface {
	string | []string | uuid.UUID
}

type CacheBuilder struct {
	cache      valkey.Client
	key        string
	keys       []string
	value      string
	ttl        time.Duration
	ctx        context.Context
	ctxTimeout time.Duration
	member     string
	err        error
}

func NewCacheBuilder[K KeyType](cache valkey.Client, key K) *CacheBuilder {
	cacheBuilder := CacheBuilder{
		cache:      cache,
		ttl:        1 * time.Hour,
		ctxTimeout: 5 * time.Second,
		ctx:        context.Background(),
	}

	switch any(key).(type) {
	case string:
		cacheBuilder.key = any(key).(string)
	case uuid.UUID:
		cacheBuilder.key = any(key).(uuid.UUID).String()
	case []string:
		cacheBuilder.keys = any(key).([]string)
	}

	return &cacheBuilder
}

func (c *CacheBuilder) WithValue(value string) *CacheBuilder {
	c.value = value
	return c
}

func (c *CacheBuilder) WithStruct(value any) *CacheBuilder {
	bytes, err := json.Marshal(value)
	if err != nil {
		c.err = fmt.Errorf("failed to marshal value to json: %w", err)
		return c
	}

	c.value = string(bytes)
	return c
}

func (cb *CacheBuilder) WithHashPattern(hashPattern string) *CacheBuilder {
	if hashPattern != "" {
		cb.key = fmt.Sprintf(hashPattern, cb.key)
	}

	return cb
}

func (cb *CacheBuilder) WithTTL(ttl time.Duration) *CacheBuilder {
	cb.ttl = ttl
	return cb
}

func (cb *CacheBuilder) WithContext(ctx context.Context) *CacheBuilder {
	cb.ctx = ctx
	return cb
}

func (cb *CacheBuilder) WithTimeout(timeout time.Duration) *CacheBuilder {
	cb.ctxTimeout = timeout
	return cb
}

func (cb *CacheBuilder) WithKeys(keys []string) *CacheBuilder {
	cb.keys = keys
	return cb
}

func (cb *CacheBuilder) Set() error {
	if cb.err != nil {
		return cb.err
	}

	ctx, cancel := cb.createTimeoutContext()
	defer cancel()

	if cb.key == "" {
		return fmt.Errorf("key is required")
	}

	if cb.value == "" {
		return fmt.Errorf("value is required")
	}

	return cb.cache.Do(ctx, cb.cache.B().Set().Key(cb.key).Value(cb.value).Ex(cb.ttl).Build()).
		Error()
}

func (cb *CacheBuilder) Get(result any) (bool, error) {
	if cb.err != nil {
		return false, cb.err
	}

	if cb.key == "" {
		return false, fmt.Errorf("key is required")
	}

	ctx, cancel := cb.createTimeoutContext()
	defer cancel()

	data, err := cb.cache.Do(ctx, cb.cache.B().Get().Key(cb.key).Build()).ToString()
	if err != nil {
		if isKeyNotFoundError(err) {
			return false, nil
		}
		return false, err
	}

	if data == "" {
		return false, nil
	}

	err = json.Unmarshal([]byte(data), result)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (cb *CacheBuilder) MGet(results any) (bool, error) {
	if cb.keys == nil {
		return false, fmt.Errorf("keys is required")
	}

	ctx, cancel := cb.createTimeoutContext()
	defer cancel()

	values, err := cb.cache.Do(ctx, cb.cache.B().Mget().Key(cb.keys...).Build()).ToArray()
	if err != nil {
		return false, err
	}

	for _, value := range values {
		if value.IsNil() {
			return false, nil
		}
	}

	var jsonStructs []any
	for _, value := range values {
		str, err := value.ToString()
		if err != nil {
			return false, err
		}
		var jsonStr any
		err = json.Unmarshal([]byte(str), &jsonStr)
		if err != nil {
			return false, err
		}

		jsonStructs = append(jsonStructs, jsonStr)
	}

	jsonBytes, err := json.Marshal(jsonStructs)
	if err != nil {
		return false, err
	}

	err = json.Unmarshal(jsonBytes, results)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (cb *CacheBuilder) Delete() error {
	if cb.err != nil {
		return cb.err
	}

	ctx, cancel := cb.createTimeoutContext()
	defer cancel()

	return cb.cache.Do(ctx, cb.cache.B().Del().Key(cb.key).Build()).Error()
}

// SADD

func (cb *CacheBuilder) WithMember(id string) *CacheBuilder {
	cb.member = id
	return cb
}

func (cb *CacheBuilder) WithMemberUUID(id uuid.UUID) *CacheBuilder {
	cb.member = id.String()
	return cb
}

func (cb *CacheBuilder) SetSadd() error {
	if cb.err != nil {
		return cb.err
	}

	if cb.member == "" {
		return fmt.Errorf("member is required")
	}

	ctx, cancel := cb.createTimeoutContext()
	defer cancel()

	return cb.cache.Do(ctx,
		cb.cache.B().Sadd().
			Key(cb.key).
			Member(cb.member).
			Build()).Error()
}

func (cb *CacheBuilder) RemoveSetMember() error {
	if cb.err != nil {
		return cb.err
	}

	if cb.member == "" {
		return fmt.Errorf("member is required")
	}

	ctx, cancel := cb.createTimeoutContext()
	defer cancel()

	return cb.cache.Do(ctx,
		cb.cache.B().Srem().
			Key(cb.key).
			Member(cb.member).
			Build()).Error()
}

func (cb *CacheBuilder) GetSetMembers() ([]string, error) {
	if cb.err != nil {
		return nil, cb.err
	}

	ctx, cancel := cb.createTimeoutContext()
	defer cancel()

	result, err := cb.cache.Do(ctx, cb.cache.B().Smembers().Key(cb.key).Build()).AsStrSlice()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (cb *CacheBuilder) createTimeoutContext() (context.Context, context.CancelFunc) {
	if deadline, ok := cb.ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return context.WithCancel(cb.ctx)
		}
		if remaining < cb.ctxTimeout {
			return context.WithCancel(cb.ctx)
		}
	}
	return context.WithTimeout(cb.ctx, cb.ctxTimeout)
}

// isKeyNotFoundError checks if the error is a "key not found" error from Valkey/Redis
func isKeyNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	// Check for Valkey/Redis specific "key not found" errors
	errStr := err.Error()
	return strings.Contains(errStr, "key not found") ||
		strings.Contains(errStr, "nil") ||
		valkey.IsValkeyNil(err)
}
