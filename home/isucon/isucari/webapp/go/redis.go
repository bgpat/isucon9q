package main

import (
	"github.com/go-redis/redis"
)

var redisCli = redis.NewClient(&redis.Options{
	Addr:     "isucon9q-2:6379",
	Password: "", // no password set
	DB:       0,  // use default DB
})
