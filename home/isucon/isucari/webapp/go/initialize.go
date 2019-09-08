package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

func postInitialize(w http.ResponseWriter, r *http.Request) {
	ri := reqInitialize{}

	err := json.NewDecoder(r.Body).Decode(&ri)
	if err != nil {
		outputErrorMsg(w, http.StatusBadRequest, "json decode error")
		return
	}

	cmd := exec.Command("../sql/init.sh")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stderr
	cmd.Run()
	if err != nil {
		outputErrorMsg(w, http.StatusInternalServerError, "exec init.sh error")
		return
	}

	_, err = dbx.Exec(
		"INSERT INTO `configs` (`name`, `val`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `val` = VALUES(`val`)",
		"payment_service_url",
		ri.PaymentServiceURL,
	)
	if err != nil {
		log.Print(err)
		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		return
	}
	_, err = dbx.Exec(
		"INSERT INTO `configs` (`name`, `val`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `val` = VALUES(`val`)",
		"shipment_service_url",
		ri.ShipmentServiceURL,
	)
	if err != nil {
		log.Print(err)
		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		return
	}

	if err := redisCli.FlushDB().Err(); err != nil {
		log.Printf("%+v\n", err)
		outputErrorMsg(w, http.StatusInternalServerError, "redis error")
		return
	}
	if err := initializeItems(); err != nil {
		log.Printf("%+v\n", err)
		outputErrorMsg(w, http.StatusInternalServerError, "redis error")
		return
	}

	res := resInitialize{
		// キャンペーン実施時には還元率の設定を返す。詳しくはマニュアルを参照のこと。
		Campaign: 0,
		// 実装言語を返す
		Language: "Go",
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	json.NewEncoder(w).Encode(res)
}

func initializeItems() error {
	var eg errgroup.Group
	var items []Item
	for _, status := range []string{
		"on_sale",
	} {
		status := status
		eg.Go(func() error {
			err := dbx.Select(&items, "SELECT * FROM `items` WHERE `status` = ?", status)
			if err != nil {
				return errors.WithStack(err)
			}
			z := make([]redis.Z, 0, len(items))
			for _, item := range items {
				z = append(z, redis.Z{
					Score:  float64(item.CreatedAt.Unix()) + float64(item.CreatedAt.UnixNano())*1e-18,
					Member: item.ID,
				})
			}
			key := itemsKey(status)
			return errors.WithStack(redisCli.ZAdd(key, z...).Err())
		})
	}
	return errors.WithStack(eg.Wait())
}
