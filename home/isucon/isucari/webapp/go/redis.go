package main

import (
	"os"

	"github.com/go-redis/redis"
)

var redisCli = redis.NewClient(&redis.Options{
	Addr:     os.Getenv("REDIS_ADDR"),
	Password: "", // no password set
	DB:       0,  // use default DB
})
