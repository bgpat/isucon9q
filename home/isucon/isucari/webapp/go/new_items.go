package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

func getNewItems(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	itemIDStr := query.Get("item_id")
	var itemID int64
	var err error
	if itemIDStr != "" {
		itemID, err = strconv.ParseInt(itemIDStr, 10, 64)
		if err != nil || itemID <= 0 {
			outputErrorMsg(w, http.StatusBadRequest, "item_id param error")
			return
		}
	}

	createdAtStr := query.Get("created_at")
	var createdAtInt64 int64
	var createdAt time.Time = time.Now()
	if createdAtStr != "" {
		createdAtInt64, err = strconv.ParseInt(createdAtStr, 10, 64)
		if err != nil || createdAtInt64 <= 0 {
			outputErrorMsg(w, http.StatusBadRequest, "created_at param error")
			return
		}
		createdAt=time.Unix(createdAtInt64, 0)
	}

	items := []Item{}
	if itemID > 0 && createdAtInt64 > 0 {
		// paging
		items, err = getItems([]string{ItemStatusOnSale, ItemStatusSoldOut}, createdAt, ItemsPerPage+1, itemID)
		if err != nil {
			log.Print(err)
			outputErrorMsg(w, http.StatusInternalServerError, "db error")
			return
		}
	} else {
		items, err = getItems([]string{ItemStatusOnSale, ItemStatusSoldOut}, createdAt, ItemsPerPage+1, 999999)
		// 1st page
		if err != nil {
			log.Print(err)
			outputErrorMsg(w, http.StatusInternalServerError, "db error")
			return
		}
	}

	itemSimples := []ItemSimple{}
	for _, item := range items {
		seller, err := getUserSimpleByID(dbx, item.SellerID)
		if err != nil {
			outputErrorMsg(w, http.StatusNotFound, "seller not found")
			return
		}
		category, err := getCategoryByID(dbx, item.CategoryID)
		if err != nil {
			outputErrorMsg(w, http.StatusNotFound, "category not found")
			return
		}
		itemSimples = append(itemSimples, ItemSimple{
			ID:         item.ID,
			SellerID:   item.SellerID,
			Seller:     &seller,
			Status:     item.Status,
			Name:       item.Name,
			Price:      item.Price,
			ImageURL:   getImageURL(item.ImageName),
			CategoryID: item.CategoryID,
			Category:   &category,
			CreatedAt:  item.CreatedAt.Unix(),
		})
	}

	hasNext := false
	if len(itemSimples) > ItemsPerPage {
		hasNext = true
		itemSimples = itemSimples[0:ItemsPerPage]
	}

	rni := resNewItems{
		Items:   itemSimples,
		HasNext: hasNext,
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	json.NewEncoder(w).Encode(rni)
}
