package handlers

import (
	"encoding/json"
	"homelab-reader/pkg/database"
	"homelab-reader/pkg/models"
	"net/http"
	"time"
)

func GetRssHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("The method isn't Get"))
		return
	}

	var rss_feeds []models.RSSFeed
	var title, url, category string
	var id int64
	var created_at time.Time

	query := "SELECT id,title,url,category,created_at FROM rss_feeds ORDER BY id ASC"
	rows, err := database.DB.Query(query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("The select data failed"))
		return
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&id, &title, &url, &category, &created_at)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("The row scan failed"))
			return
		}
		rss_feeds = append(rss_feeds, models.RSSFeed{
			ID:        id,
			Title:     title,
			URL:       url,
			Category:  category,
			CreatedAt: created_at,
		})
	}

	err = rows.Err()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("The row scan failed"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&rss_feeds)

}
