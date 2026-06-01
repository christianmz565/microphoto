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
func (c *Client) PopTaskReliable(ctx context.Context, taskID string) (*jobs.Job, error) {
	queueKey := fmt.Sprintf("{%s}:queue", taskID)
	progressKey := fmt.Sprintf("{%s}:in_progress", taskID)

	data, err := c.rdb.BLMove(ctx, queueKey, progressKey, "RIGHT", "LEFT", 0).Result()
	if err != nil {
		return nil, fmt.Errorf("blmove: %w", err)
	}

	job := &jobs.Job{}
	if err := proto.Unmarshal([]byte(data), job); err != nil {
		return nil, fmt.Errorf("unmarshal job: %w", err)
	}

	return job, nil
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

// Close closes the redis client
func (c *Client) Close() error {
	return c.rdb.Close()
}
