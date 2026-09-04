package models

import (
	"encoding/xml"
	"time"
)

// 建表时候需要使用的结构体
type User struct {
	ID           int64     `json:"id"`
	Role         string    `json:"role"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Session struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Token     string    `json:"session_token"`
	ExpiresAt time.Time `json:"expres_at"`
}

type RSSFeed struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
}

type Book struct {
	ID        int64     `josn:"id"`
	Title     string    `json:"title"`
	FilePath  string    `json:"file_path"`
	Format    string    `json:"format"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

// 构建RSS的整体结构体,用于实际解析xml数据
type RSSItem struct {
	Title       string `xml:"title" json:"title"`
	Link        string `xml:"link" json:"link"`
	Description string `xml:"description" json:"description"`
	PubDate     string `xml:"pubDate" json:"pub_date"`
}

type RSSChannel struct {
	Title string    `xml:"title"`
	Items []RSSItem `xml:"item"`
}

type RssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Channel RSSChannel `xml:"channel"`
}

// 鉴权时，使用的结构体
type RegisterReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
