package redisclient

import (
    "context"
    "os"

    "github.com/redis/go-redis/v9"
)

var Client *redis.Client
var Ctx = context.Background()

func Connect() error {
    redisURL := os.Getenv("REDIS_URL")
    var opts *redis.Options
    if redisURL != "" {
        var err error
        opts, err = redis.ParseURL(redisURL)
        if err != nil {
            return err
        }
    } else {
        // fallback for local development
        opts = &redis.Options{
            Addr:     os.Getenv("REDIS_ADDR"),
            Password: os.Getenv("REDIS_PASSWORD"),
            DB:       0,
        }
    }
    Client = redis.NewClient(opts)
    _, err := Client.Ping(Ctx).Result()
    return err
}
