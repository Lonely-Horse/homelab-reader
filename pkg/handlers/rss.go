package handlers

import (
	"encoding/json"
	"homelab-reader/pkg/database"
	"homelab-reader/pkg/models"
	"homelab-reader/pkg/rss"
	"net/http"
	"strconv"
	"strings"
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

func CreateRssHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("The Method isn't Post"))
		return
	}

	var rssfeed models.RSSFeed

	err := json.NewDecoder(r.Body).Decode(&rssfeed)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("The body decode failed"))
		return
	}
	if rssfeed.Title == "" || !rss.ValidRssUrl(rssfeed.URL, "https://", "http://") {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("The rss get failed"))
		return
	}

	query := "INSERT INTO rss_feeds(title,url,category,created_at) VALUES(?,?,?,?)"
	_, err = database.DB.Exec(query, rssfeed.Title, rssfeed.URL, rssfeed.Category, time.Now())
	if err != nil {
		if strings.Contains(err.Error(), "rss_feeds.url") {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("The rss url is already exists"))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("The insert failed"))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rssfeed)

}

func DeleteRssHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("The method isn't delete"))
		return
	}

	idstr := r.PathValue("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("The string switch to int failed"))
		return
	}

	query := "DELETE FROM rss_feeds WHERE id=?"
	result, err := database.DB.Exec(query, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("The delete rss failed"))
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}

func RssHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		CreateRssHandler(w, r)
	case http.MethodGet:
		GetRssHandler(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("The method isn't allowed"))
		return
	}
}

func FetchRssHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("The method isn't post"))
		return
	}

	var id int64
	var title, url, category string
	var created_at time.Time
	query := "SELECT id,title,url,category,created_at FROM rss_feeds"
	rows, err := database.DB.Query(query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("The select rss failed"))
		return
	}

	var result models.RssFetchResult
	var feeds []models.FeedRec
	for rows.Next() {
		err := rows.Scan(&id, &title, &url, &category, &created_at)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("The scan rows failed"))
			return
		}

		feeds = append(feeds, models.FeedRec{ID: id, Title: title, URL: url, Category: category})

	}

	if err = rows.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("The read rows failed"))
		return
	}
	rows.Close()

	for _, feed := range feeds {
		items, err := rss.FetchAndParseRSS(feed.URL)
		if err != nil {
			result.Failures = append(result.Failures, models.RssFailure{
				URL:   feed.URL,
				Error: err.Error(),
			})
			continue
		}

		result.Feeds = append(result.Feeds, models.RssFeedOutput{
			ID:       feed.ID,
			Title:    feed.Title,
			URL:      feed.URL,
			Category: feed.Category,
			Items:    items,
		})
	}

	json.NewEncoder(w).Encode(result)

}
