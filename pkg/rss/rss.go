package rss

import (
	"encoding/xml"
	"homelab-reader/pkg/models"
	"io"
	"net/http"
	"time"
)

func FetchAndParseRSS(feedURL string) ([]models.RSSItem, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(feedURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	limitedReader := io.LimitReader(resp.Body, 5*1024*1024)

	var feed models.RssFeed
	decoder := xml.NewDecoder(limitedReader)
	err = decoder.Decode(&feed)
	if err != nil {
		return nil, err
	}

	return feed.Channel.Items, nil
}
