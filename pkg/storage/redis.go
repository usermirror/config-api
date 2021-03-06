package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	redis "github.com/segmentio/redis-go"
)

// Redis backed persistence for arbitrary key/values.
type Redis struct{}

// implements Store interface
var _ Store = Redis{}

func (r Redis) Init() error {
	return nil
}

// Get ...
func (Redis) Get(input GetInput) ([]byte, error) {
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
func (Redis) Set(input SetInput) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := redis.Exec(ctx, "SET", input.Key, input.Value); err != nil {
		fmt.Println("redis.set.fail: server error")
		fmt.Println(err)

		return err
	}

	return nil
}

func (Redis) Scan(input ScanInput) (KeyList, error) {
	return KeyList{}, nil
}

func (Redis) CheckAuth(input AuthInput) error {
	return errors.New("operation not supported by this provider")
}
