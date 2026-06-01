package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/christianmz565/microphoto/pkg/model"
	"github.com/christianmz565/microphoto/proto/jobs"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
)

type Client struct {
	rdb redis.UniversalClient
}

func NewClient(addr string) (*Client, error) {
	rdb := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs: []string{addr},
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &Client{rdb: rdb}, nil
}

// PushTask serializes a job to Protobuf and pushes it to {taskID}:queue
func (c *Client) PushTask(ctx context.Context, taskID string, job *jobs.Job) error {
	data, err := proto.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal job: %w", err)
	}

	queueKey := fmt.Sprintf("{%s}:queue", taskID)
	return c.rdb.LPush(ctx, queueKey, data).Err()
}

// PopTaskReliable moves a task from {taskID}:queue to {taskID}:in_progress using BLMOVE
func (c *Client) PopTaskReliable(ctx context.Context, taskID string) (*jobs.Job, []byte, error) {
	queueKey := fmt.Sprintf("{%s}:queue", taskID)
	progressKey := fmt.Sprintf("{%s}:in_progress", taskID)

	data, err := c.rdb.BLMove(ctx, queueKey, progressKey, "RIGHT", "LEFT", 0).Result()
	if err != nil {
		return nil, nil, fmt.Errorf("blmove: %w", err)
	}

	job := &jobs.Job{}
	if err := proto.Unmarshal([]byte(data), job); err != nil {
		return nil, nil, fmt.Errorf("unmarshal job: %w", err)
	}

	return job, []byte(data), nil
}

// DecrementCounter decrements the {taskID}:subtasks counter
func (c *Client) DecrementCounter(ctx context.Context, taskID string) (int64, error) {
	counterKey := fmt.Sprintf("{%s}:subtasks", taskID)
	return c.rdb.Decr(ctx, counterKey).Result()
}

// PublishProgress serializes progress to JSON and publishes it to the {taskID} channel
func (c *Client) PublishProgress(ctx context.Context, taskID string, payload model.ProgressPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal progress: %w", err)
	}

	channel := fmt.Sprintf("{%s}", taskID)
	return c.rdb.Publish(ctx, channel, data).Err()
}

// InitializeTask sets the {taskID}:subtasks and {taskID}:attempts counters
func (c *Client) InitializeTask(ctx context.Context, taskID string, subtasks int, attempts int) error {
	subtasksKey := fmt.Sprintf("{%s}:subtasks", taskID)
	attemptsKey := fmt.Sprintf("{%s}:attempts", taskID)

	pipe := c.rdb.Pipeline()
	pipe.Set(ctx, subtasksKey, subtasks, 0)
	pipe.Set(ctx, attemptsKey, attempts, 0)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("pipeline exec: %w", err)
	}

	return nil
}

// Close closes the redis client
func (c *Client) Close() error {
	return c.rdb.Close()
}

// ScanInProgressKeys returns keys matching the pattern using SCAN
func (c *Client) ScanInProgressKeys(ctx context.Context, pattern string) ([]string, error) {
	var keys []string
	iter := c.rdb.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return keys, nil
}

// GetListItems returns all items in a list
func (c *Client) GetListItems(ctx context.Context, key string) ([][]byte, error) {
	items, err := c.rdb.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	result := make([][]byte, len(items))
	for i, item := range items {
		result[i] = []byte(item)
	}
	return result, nil
}

// GetAttempts returns the value of {taskID}:attempts
func (c *Client) GetAttempts(ctx context.Context, taskID string) (int, error) {
	key := fmt.Sprintf("{%s}:attempts", taskID)
	val, err := c.rdb.Get(ctx, key).Int()
	if err != nil {
		return 0, err
	}
	return val, nil
}

// RescheduleTask moves a task from in_progress back to queue atomically
func (c *Client) RescheduleTask(ctx context.Context, taskID string, oldData []byte, newData []byte) error {
	progressKey := fmt.Sprintf("{%s}:in_progress", taskID)
	queueKey := fmt.Sprintf("{%s}:queue", taskID)
	attemptsKey := fmt.Sprintf("{%s}:attempts", taskID)

	_, err := c.rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Decr(ctx, attemptsKey)
		pipe.LRem(ctx, progressKey, 1, oldData)
		pipe.LPush(ctx, queueKey, newData)
		return nil
	})

	return err
}

// CleanupFailedTask removes a task from in_progress and deletes metadata
func (c *Client) CleanupFailedTask(ctx context.Context, taskID string, data []byte) error {
	progressKey := fmt.Sprintf("{%s}:in_progress", taskID)
	subtasksKey := fmt.Sprintf("{%s}:subtasks", taskID)
	attemptsKey := fmt.Sprintf("{%s}:attempts", taskID)

	_, err := c.rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.LRem(ctx, progressKey, 1, data)
		pipe.Del(ctx, subtasksKey)
		pipe.Del(ctx, attemptsKey)
		return nil
	})

	return err
}

// CompleteTask removes a task from the in_progress list after successful processing
func (c *Client) CompleteTask(ctx context.Context, taskID string, data []byte) error {
	progressKey := fmt.Sprintf("{%s}:in_progress", taskID)
	return c.rdb.LRem(ctx, progressKey, 1, data).Err()
}
