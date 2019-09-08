package main

import (
	"os"
	"time"

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

func addItemStatus(id int64, createdAt time.Time, status string) error {
	z := redis.Z{
		Score:  float64(createdAt.Unix()) + float64(createdAt.UnixNano())*1e-18,
		Member: id,
	}
	if err := redisCli.ZAdd(itemsKey(status), z).Err(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func updateItemStatus(item Item, newStatus string) error {
	if err := redisCli.ZRem(itemsKey(item.Status), item.ID).Err(); err != nil {
		return errors.WithStack(err)
	}
	if err := addItemStatus(item.ID, item.CreatedAt, newStatus); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
