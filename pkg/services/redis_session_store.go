package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/go-redis/redis/v8"
)

// RedisSessionStore is a Redis-based implementation of the SessionStore interface.
type RedisSessionStore struct {
	client *redis.Client
}

// NewRedisSessionStore creates a new RedisSessionStore.
func NewRedisSessionStore(addr, password string, db int) (*RedisSessionStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisSessionStore{
		client: client,
	}, nil
}

// Save saves the chat history for a given session ID to Redis.
func (s *RedisSessionStore) Save(sessionID string, history []*types.Content) error {
	ctx := context.Background()
	data, err := json.Marshal(history)
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	if err := s.client.Set(ctx, sessionID, data, 0).Err(); err != nil {
		return fmt.Errorf("failed to save session to redis: %w", err)
	}
	return nil
}

// Load loads the chat history for a given session ID from Redis.
func (s *RedisSessionStore) Load(sessionID string) ([]*types.Content, error) {
	ctx := context.Background()
	data, err := s.client.Get(ctx, sessionID).Bytes()
	if err != nil {
		if err == redis.Nil {
			return []*types.Content{}, nil
		}
		return nil, fmt.Errorf("failed to load session from redis: %w", err)
	}

	var history []*types.Content
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, fmt.Errorf("failed to unmarshal history: %w", err)
	}

	return history, nil
}

// List returns a sorted list of all saved session IDs from Redis.
func (s *RedisSessionStore) List() ([]string, error) {
	ctx := context.Background()
	keys, err := s.client.Keys(ctx, "*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions from redis: %w", err)
	}

	sort.Strings(keys)
	return keys, nil
}

// Delete deletes the session for a given session ID from Redis.
func (s *RedisSessionStore) Delete(sessionID string) error {
	ctx := context.Background()
	if err := s.client.Del(ctx, sessionID).Err(); err != nil {
		return fmt.Errorf("failed to delete session from redis: %w", err)
	}
	return nil
}
