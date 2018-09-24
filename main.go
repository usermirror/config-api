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
		Addr           string `conf:"addr" help:"Address where to bind the service"`
		CheckAuth      bool   `conf:"check-auth" help:"Use store to verify that a write token is correct"`
		EtcdAddr       string `conf:"etcd-addr" help:"etcd client port"`
		RedisAddr      string `conf:"redis-addr" help:"Redis server address"`
		RedisPassword  string `conf:"redis-password" help:"Redis server password"`
		VaultAddr      string `conf:"vault-addr" help:"Vault server address"`
		VaultToken     string `conf:"vault-token" help:"Vault root token"`
		PostgresAddr   string `conf:"postgres-addr" help:"Postgres database address"`
		StorageBackend string `conf:"storage-backend" help:"Default storage backend for configs"`
	}{
		Addr:           ":8888",
		CheckAuth:      true,
		EtcdAddr:       "localhost:2379",
		RedisAddr:      "localhost:6379",
		VaultAddr:      "http://localhost:8200/",
		VaultToken:     "1e7d2b9b-de0e-67a6-9362-6b9b01bf4e89",
		PostgresAddr:   "postgres://postgres:@localhost:5432/postgres?sslmode=disable",
		StorageBackend: "vault",
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
		Addr:           config.Addr,
		CheckAuth:      config.CheckAuth,
		EtcdAddr:       config.EtcdAddr,
		RedisAddr:      config.RedisAddr,
		VaultAddr:      config.VaultAddr,
		VaultToken:     config.VaultToken,
		PostgresAddr:   config.PostgresAddr,
		StorageBackend: config.StorageBackend,
	}

	fmt.Println(fmt.Sprintf("server.start: api ready on %s", config.Addr))

	log.Fatal(s.Listen())
}
