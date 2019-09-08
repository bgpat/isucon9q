package main

import (
	"os"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

var redisCli = redis.NewClient(&redis.Options{
	Addr:     os.Getenv("REDIS_ADDR"),
	Password: "", // no password set
	DB:       0,  // use default DB
})

func itemsKey(status string) string {
	return "items_" + status
}

func updateItemStatus(item Item, newStatus string) error {
	if err := redisCli.ZRem(itemsKey(item.Status), item.ID).Err(); err != nil {
		return errors.WithStack(err)
	}
	z := redis.Z{
		Score:  float64(item.CreatedAt.Unix()) + float64(item.CreatedAt.UnixNano())*1e-18,
		Member: item.ID,
	}
	if err := redisCli.ZAdd(itemsKey(item.Status), z).Err(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
