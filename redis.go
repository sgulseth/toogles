package main

import (
    "os"
    "log"
    "gopkg.in/redis.v4"
)

var client *redis.Client
func connectToRedis() {
    redisAddr := os.Getenv("REDIS_ADDR")
    if redisAddr == "" {
        redisAddr = "127.0.0.1:6379"
    }

    redisPassword := os.Getenv("REDIS_PASSWORD")

    client = redis.NewClient(&redis.Options{
        Addr:     redisAddr,
        Password: redisPassword, // no password set
        DB:       0,  // use default DB
    })

    log.Println("Connected to redis")
}

func getRedisClient() *redis.Client {
    if client == nil {
        connectToRedis()
    }

    return client
}