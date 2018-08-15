package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/segmentio/conf"
	"github.com/segmentio/redis-go"

	"github.com/usermirror/config-api/pkg/server"
)

func main() {
	config := struct {
		Addr          string `conf:"addr" help:"Address where to bind the service, default = :8888"`
		RedisAddr     string `conf:"redis-addr" help:"Redis server address, default = localhost:6379"`
		RedisPassword string `conf:"redis-password" help:"Redis server password"`
	}{
		Addr:      ":8888",
		RedisAddr: "localhost:6379",
	}

	conf.Load(&config)

	envRedisAddr := os.Getenv("REDIS_ADDR")

	if envRedisAddr != "" {
		config.RedisAddr = envRedisAddr
	}

	// TODO: passthrough to store.Redis instead of using package global
	redis.DefaultClient = &redis.Client{
		Addr: config.RedisAddr,
	}

	if config.RedisPassword != "" {
		ctx := context.Background()
		err := redis.DefaultClient.Exec(ctx, "AUTH", config.RedisPassword)

		if err != nil {
			fmt.Println(fmt.Sprintf("redis.auth: %v", err))
		}
	}

	s := &server.Server{
		Addr: config.Addr,
	}

	fmt.Println(fmt.Sprintf("server.start: api ready on %s", config.Addr))
	fmt.Println(fmt.Sprintf("redis.connect: %s", config.RedisAddr))

	log.Fatal(s.Listen())
}
