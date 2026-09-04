package bootstrap

import (
	"database/sql"
	"log"
)

var DB *sql.DB

func createTables() error {
	// 直接制定建表的规则
	schema := `
	-- 1.用户表
	CREATE TABLE IF NOT EXISTS
	users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		role TEXT DEFAULT 'user' NOT NULL,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	--2.会话表 (session)
	CREATE TABLE IF NOT EXISTS
	sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		session_token TEXT UNIQUE NOT NULL,
		expires_at DATETIME NOT NULL,
		FOREIGN KEY (user_id)
		REFERENCES users(id) ON DELETE CASCADE
	);

	--3.RSS 订阅表
	CREATE TABLE IF NOT EXISTS
	rss_feeds (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		url TEXT UNIQUE NOT NULL,
		category TEXT NOT NULL,
		created_at DATETIME NOT NULL
	);

	--4.书籍元数据表
	CREATE TABLE IF NOT EXISTS
	books (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		filepath TEXT UNIQUE NOT NULL,
		format TEXT NOT NULL,
		size INTEGER NOT NULL,
		created_at DATETIME NOT NULL
	);
	`

	//开始建表
	_, err := DB.Exec(schema)
	if err != nil {
		//失败的话，必须要关闭数据库，防止内存泄露
		DB.Close()
		return err
	}

	return nil
}

func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "file:"+dbPath+"?_journal_mode=WAL")
	if err != nil {
		return nil, err
	}

	//规定数据库同时并发申请为1
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	DB = db

	err = createTables()
	if err != nil {
		db.Close()
		return nil, err
	}

	log.Println("[InitDB] Sqlite init successfully with WAL mode")
	return db, nil
}
