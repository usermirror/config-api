package config

import (
	"context"
	"fmt"
	"time"

	redis "github.com/segmentio/redis-go"
)

// GetInput ...
type GetInput struct {
	Key     string
	Timeout int
}

// SetInput ...
type SetInput struct {
	Key     string
	Value   []byte
	Timeout int
}

// Get ...
func Get(input GetInput) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var args = redis.Query(ctx, "GET", input.Key)
	var value []byte

	if args.Next(&value) {
		return value, nil
	}

	return nil, args.Close()
}

// Set ...
func Set(input SetInput) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := redis.Exec(ctx, "SET", input.Key, input.Value); err != nil {
		fmt.Println("models.config.redis: set failed")
		fmt.Println(err)

		return err
	}

	return nil
}
