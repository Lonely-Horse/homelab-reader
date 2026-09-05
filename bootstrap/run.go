package bootstrap

import (
	"embed"
	"homelab-reader/pkg/database"
	"homelab-reader/pkg/handlers"
	"log"
	"os"
	"path/filepath"
)

func Run(s *handlers.AppServer, tmplFS embed.FS) error {
	dbPath := "./data/data.db"
	err := os.MkdirAll(filepath.Dir(dbPath), 0o755)
	_, err = database.InitDB(dbPath)
	if err != nil {
		log.Printf("The database init failed.detail: %s", err)
		return err
	}

	InitServer(s, tmplFS)

	return nil

}
