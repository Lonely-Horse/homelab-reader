package handlers

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"
)

type AppServer struct {
	Tmpl   *template.Template
	ViewFS fs.FS
}

func IndexRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/dashboard", http.StatusFound) // 302
}

func (s *AppServer) BooksPage(w http.ResponseWriter, r *http.Request) {
	err := s.Tmpl.ExecuteTemplate(w, "books.html", nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Failed execute the config")
		w.Write([]byte("Failed executetemplate html"))
		return
	}
}

func (s *AppServer) RssPage(w http.ResponseWriter, r *http.Request) {
	err := s.Tmpl.ExecuteTemplate(w, "rss.html", nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Failed execute the config")
		w.Write([]byte("Failed executetemplate html"))
		return
	}
}

func (s *AppServer) UserPage(w http.ResponseWriter, r *http.Request) {
	err := s.Tmpl.ExecuteTemplate(w, "user.html", nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Failed execute the config")
		w.Write([]byte("Failed executetemplate html"))
		return
	}
}

func (s *AppServer) DashboardPage(w http.ResponseWriter, r *http.Request) {
	err := s.Tmpl.ExecuteTemplate(w, "dashboard.html", nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Failed execute the config")
		w.Write([]byte("Failed executetemplate html"))
		return
	}
}
