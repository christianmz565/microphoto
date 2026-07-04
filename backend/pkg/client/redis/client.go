// Package redis provides a Redis client for task queuing and pub/sub.
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/christianmz565/microphoto/pkg/model"
	jobs "github.com/christianmz565/microphoto/proto/jobs/v1"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
)

// Client wraps a Redis universal client for task management operations.
type Client struct {
	rdb redis.UniversalClient
}

// NewClient creates a new Redis client connected to the specified address.
func NewClient(addr string) (*Client, error) {
	rdb := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs: []string{addr},
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &Client{rdb: rdb}, nil
}

// PushTask serializes a job to Protobuf and pushes it to {"global"}:queue.
func (c *Client) PushTask(ctx context.Context, job *jobs.Job) error {
	data, err := proto.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal job: %w", err)
	}

	queueKey := `{"global"}:queue`

	return c.rdb.LPush(ctx, queueKey, data).Err()
}

// PushTasksPipeline serializes multiple jobs to Protobuf and pushes them to {"global"}:queue in a pipeline.
func (c *Client) PushTasksPipeline(ctx context.Context, jobs []*jobs.Job) error {
	pipe := c.rdb.Pipeline()
	queueKey := `{"global"}:queue`

	for _, job := range jobs {
		data, err := proto.Marshal(job)
		if err != nil {
			return fmt.Errorf("marshal job %s: %w", job.Id, err)
		}

		pipe.LPush(ctx, queueKey, data)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("pipeline exec: %w", err)
	}

	return nil
}

// PopTaskReliable moves a task from {"global"}:queue to {"global"}:in_progress:{taskID} using BLMOVE
// and updates the job.Timestamp to the current time to reset the reaper timeout.
func (c *Client) PopTaskReliable(ctx context.Context, taskID string) (*jobs.Job, []byte, error) {
	queueKey := `{"global"}:queue`
	progressKey := "{\"global\"}:in_progress:" + taskID

	data, err := c.rdb.BLMove(ctx, queueKey, progressKey, "RIGHT", "LEFT", 0).Result()
	if err != nil {
		return nil, nil, fmt.Errorf("blmove: %w", err)
	}

	job := &jobs.Job{}
	if err := proto.Unmarshal([]byte(data), job); err != nil {
		return nil, nil, fmt.Errorf("unmarshal job: %w", err)
	}

	job.Timestamp = time.Now().Unix()

	newData, err := proto.Marshal(job)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal job: %w", err)
	}

	if err := c.rdb.LSet(ctx, progressKey, 0, string(newData)).Err(); err != nil {
		return nil, nil, fmt.Errorf("lset: %w", err)
	}

	return job, newData, nil
}

// SetNX sets a key if it doesn't exist.
func (c *Client) SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	return c.rdb.SetNX(ctx, key, value, expiration).Result()
}

// DecrementCounter decrements the {"global"}:subtasks:{taskID} counter.
func (c *Client) DecrementCounter(ctx context.Context, taskID string) (int64, error) {
	counterKey := "{\"global\"}:subtasks:" + taskID
	return c.rdb.Decr(ctx, counterKey).Result()
}

// PublishProgress serializes progress to JSON, appends it to the history list, and publishes it to the {"global"}:progress:{taskID} channel.
func (c *Client) PublishProgress(ctx context.Context, taskID string, payload model.ProgressPayload) error {
	if payload.Timestamp == 0 {
		payload.Timestamp = time.Now().UnixNano()
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal progress: %w", err)
	}

	listKey := "{\"global\"}:events:" + taskID
	channel := "{\"global\"}:progress:" + taskID

	pipe := c.rdb.Pipeline()
	pipe.RPush(ctx, listKey, data)
	pipe.Publish(ctx, channel, data)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("publish progress: %w", err)
	}

	return nil
}

// GetProgressEvents returns all stored progress events for a task from the {"global"}:events:{taskID} list.
func (c *Client) GetProgressEvents(ctx context.Context, taskID string) ([]model.ProgressPayload, error) {
	listKey := "{\"global\"}:events:" + taskID

	items, err := c.rdb.LRange(ctx, listKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("lrange events: %w", err)
	}

	events := make([]model.ProgressPayload, 0, len(items))
	for _, item := range items {
		var payload model.ProgressPayload
		if err := json.Unmarshal([]byte(item), &payload); err != nil {
			continue
		}

		events = append(events, payload)
	}

	return events, nil
}

// SubscribeProgress subscribes to the {"global"}:progress:{taskID} Pub/Sub channel.
func (c *Client) SubscribeProgress(ctx context.Context, taskID string) (*redis.PubSub, <-chan *redis.Message) {
	channel := "{\"global\"}:progress:" + taskID
	pubsub := c.rdb.Subscribe(ctx, channel)
	ch := pubsub.Channel()

	return pubsub, ch
}

// InitializeTask sets the {"global"}:subtasks:{taskID} and {"global"}:attempts:{taskID} counters.
func (c *Client) InitializeTask(ctx context.Context, taskID string, subtasks, attempts int) error {
	subtasksKey := "{\"global\"}:subtasks:" + taskID
	attemptsKey := "{\"global\"}:attempts:" + taskID

	pipe := c.rdb.Pipeline()
	pipe.Set(ctx, subtasksKey, subtasks, 0)
	pipe.Set(ctx, attemptsKey, attempts, 0)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("pipeline exec: %w", err)
	}

	return nil
}

// Close closes the redis client.
func (c *Client) Close() error {
	return c.rdb.Close()
}

// ScanInProgressKeys returns keys matching the pattern using SCAN.
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

// GetListItems returns all items in a list.
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

// GetAttempts returns the value of {"global"}:attempts:{taskID}.
func (c *Client) GetAttempts(ctx context.Context, taskID string) (int, error) {
	key := "{\"global\"}:attempts:" + taskID

	val, err := c.rdb.Get(ctx, key).Int()
	if err != nil {
		return 0, err
	}

	return val, nil
}

// RescheduleTask moves a task from in_progress back to queue atomically.
func (c *Client) RescheduleTask(ctx context.Context, progressID, taskID string, oldData, newData []byte) error {
	progressKey := "{\"global\"}:in_progress:" + progressID
	queueKey := `{"global"}:queue`
	attemptsKey := "{\"global\"}:attempts:" + taskID

	_, err := c.rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Decr(ctx, attemptsKey)
		pipe.LRem(ctx, progressKey, 1, oldData)
		pipe.LPush(ctx, queueKey, newData)

		return nil
	})

	return err
}

// CleanupFailedTask removes a task from in_progress and deletes metadata.
func (c *Client) CleanupFailedTask(ctx context.Context, progressID, taskID string, data []byte) error {
	progressKey := "{\"global\"}:in_progress:" + progressID
	subtasksKey := "{\"global\"}:subtasks:" + taskID
	attemptsKey := "{\"global\"}:attempts:" + taskID

	_, err := c.rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.LRem(ctx, progressKey, 1, data)
		pipe.Del(ctx, subtasksKey)
		pipe.Del(ctx, attemptsKey)

		return nil
	})

	return err
}

// CompleteTask removes a task from the in_progress list after successful processing.
func (c *Client) CompleteTask(ctx context.Context, progressID string, data []byte) error {
	progressKey := "{\"global\"}:in_progress:" + progressID
	return c.rdb.LRem(ctx, progressKey, 1, data).Err()
}
