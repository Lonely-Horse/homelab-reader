package rss

import (
	"encoding/xml"
	"fmt"
	"homelab-reader/pkg/models"
	"io"
	"net/http"
	"strings"
	"time"
)

func ValidRssUrl(url string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(url, prefix) {
			return true
		}
	}
	return false
}

func FetchAndParseRSS(feedURL string) ([]models.RSSItem, error) {
	if !ValidRssUrl(feedURL, "https://", "http://") {
		return nil, fmt.Errorf("The URL is illegal")
	}

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

	buf := make([]byte, 1)
	_, err = resp.Body.Read(buf)
	if err == nil {
		return nil, fmt.Errorf("feed too large")
	}
	if err != io.EOF {
		return nil, err
	}

	return feed.Channel.Items, nil
}
