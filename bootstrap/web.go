package bootstrap

import (
	"embed"
	"flag"
	"homelab-reader/pkg/handlers"
	"homelab-reader/pkg/middleware"
	"log"
	"net/http"
	"time"
)

func ServerRoutes(addr string, mux *http.ServeMux) *http.Server {
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server
}

func MuxRoutes(s *handlers.AppServer, tmplFS embed.FS) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/dashboard", s.DashboardPage)
	mux.HandleFunc("/dashboard/books", s.BooksPage)
	mux.HandleFunc("/dashboard/rss", s.RssPage)
	mux.HandleFunc("/dashboard/user", s.UserPage)

	mux.HandleFunc("/api/auth/login", handlers.LoginHandler)
	mux.HandleFunc("/api/auth/register", handlers.RegisterHandler)
	mux.HandleFunc("/api/auth/logout", handlers.LogoutHandler)
	mux.HandleFunc("/api/rss", middleware.AuthMiddleware(handlers.GetRssHandler))
	mux.HandleFunc("/api/books", middleware.AuthMiddleware(handlers.BooksHandler))
	mux.HandleFunc("/api/books/{id}", middleware.AuthMiddleware(handlers.DeleteBooksHandler))
	mux.HandleFunc("/api/books/{id}/content", middleware.AuthMiddleware(handlers.GetBookContentHandler))

	// 静态资源（共享 CSS / JS），供各模板页面引用
	mux.Handle("/templates/", http.FileServer(http.FS(tmplFS)))

	mux.HandleFunc("/", handlers.IndexRedirect)

	return mux
}
func InitServer(s *handlers.AppServer, tmplFS embed.FS) {
	var addr string
	flag.StringVar(&addr, "addr", "0.0.0.0:8087", "地址")

	mux := MuxRoutes(s, tmplFS)

	server := ServerRoutes(addr, mux)

	log.Printf("The server listen on %s", server.Addr)
	err := server.ListenAndServe()
	if err != nil {
		log.Printf("The server listened failed,detail: %s", err)
		return
	}

}
