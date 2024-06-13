package db

import (
	"log"
	"otp-service/internal/config"

	"github.com/go-redis/redis"
)

type Database struct {
	client *redis.Client
}

func InitDB(cfg *config.Config) (*Database, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Database.Address,
		Password: cfg.Database.Password,
		DB:       cfg.Database.DB,
	})

	_, err := client.Ping().Result()
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
		return nil, err
	}

	return &Database{client: client}, nil
}

func (d *Database) Close() error {
	if d.client != nil {
		return d.client.Close()
	}
	return nil
}

func (d *Database) GetClient() *redis.Client {
	return d.client
}
