package main

import (
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

var redisCli = redis.NewClient(&redis.Options{
	Addr:     os.Getenv("REDIS_ADDR"),
	Password: "", // no password set
	DB:       0,  // use default DB
})

func itemsKey(status string) string {
	return "items_" + status
}

func calcScore(createdAt time.Time, id int64) float64 {
	return float64(createdAt.Unix()) + float64(id)*1e-6
}

func addItemStatus(id int64, createdAt time.Time, status string) error {
	z := redis.Z{
		Score:  calcScore(createdAt, id),
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

func getItems(statuses []string, createdAt time.Time, limit int64) ([]Item, error) {
	var eg errgroup.Group
	var zs []redis.Z
	var mu sync.Mutex
	for _, status := range statuses {
		status := status
		eg.Go(func() error {
			z, err := redisCli.ZRevRangeByScoreWithScores(itemsKey(status), redis.ZRangeBy{
				Max:   strconv.FormatFloat(calcScore(createdAt, 0), 'f', 7, 64),
				Count: limit,
			}).Result()
			if err != nil {
				return errors.WithStack(err)
			}
			mu.Lock()
			zs = append(zs, z...)
			mu.Unlock()
			//log.Println("zrange", itemsKey(status), status, len(z), scoreByCreatedAt(createdAt), strconv.FormatFloat(scoreByCreatedAt(createdAt), 'f', -1, 64))
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, errors.WithStack(err)
	}

	if len(zs) == 0 {
		return []Item{}, nil
	}

	sort.Slice(zs, func(i, j int) bool {
		return zs[i].Score > zs[j].Score
	})
	limitsize := limit
	if int64(len(zs)) < limitsize {
		limitsize = int64(len(zs))
	}
	ids := make([]string, 0, limitsize)

	for _, z := range zs[:limitsize] {
		id, ok := z.Member.(string)
		if !ok {
			return nil, errors.Errorf("failed to cast z.Member: %T", z.Member)
		}
		ids = append(ids, id)
	}
	query := "SELECT * FROM `items` WHERE `id` IN (" + strings.Join(ids, ",") + ") ORDER BY created_at DESC"
	var items []Item
	err := dbx.Select(&items, query)
	if err != nil {
		return nil, errors.Wrap(err, query)
	}
	return items, nil
}
