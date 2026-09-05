package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"homelab-reader/pkg/database"
	"homelab-reader/pkg/middleware"
	"homelab-reader/pkg/models"
	"homelab-reader/pkg/reader"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func GetBooksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("The method isn't Get"))
		return
	}

	var id, size int64
	var title, filePath, format string
	var createdAt time.Time
	var books []models.Book
	userID := r.Context().Value(middleware.UserIDKey).(int64)
	query := "SELECT id,title,filepath,format,size,created_at FROM books WHERE user_id = ? ORDER BY id ASC"
	rows, err := database.DB.Query(query, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("The database didn't select"))
		return
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&id, &title, &filePath, &format, &size, &createdAt)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("The rows didn't scan"))
			return
		}

		books = append(books, models.Book{
			ID:        id,
			Title:     title,
			FilePath:  filePath,
			Format:    format,
			Size:      size,
			CreatedAt: createdAt,
		})
	}

	err = rows.Err()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("The rows Scan have problem"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&books)

}

func UploadBooksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("The Method isn't Post"))
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(int64)
	r.Body = http.MaxBytesReader(w, r.Body, 20<<20)

	err := r.ParseMultipartForm(20 * 1024 * 1024)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("The multipart is too big"))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("The file didn't find"))
		return
	}
	defer file.Close()

	name := filepath.Base(header.Filename)
	ext := strings.ToLower(filepath.Ext(name))
	if ext != ".txt" && ext != ".epub" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Only .txt or .epub allowed"))
		return
	}

	token := make([]byte, 16)
	_, err = io.ReadFull(rand.Reader, token)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("The read have some problem"))
		return
	}

	filename := "book-" + hex.EncodeToString(token) + ext

	err = os.MkdirAll("books", 0o755)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("The dir didn't create"))
		return
	}

	dst, err := os.Create(filepath.Join("books", filename))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("The file didn't create"))
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("The file didn't copy"))
		return
	}

	info, err := os.Stat(filepath.Join("books", filename))
	if err != nil {
		log.Printf("stat failed: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("The file stat have problem"))
		return
	}

	size := info.Size()
	filePath := filepath.Join("books", filename)
	query := "INSERT INTO books (title,user_id,filepath,format,size,created_at) VALUES (?,?,?,?,?,CURRENT_TIMESTAMP)"
	title := filepath.Base(name)

	res, err := database.DB.Exec(query, title, userID, filePath, ext, size)
	if err != nil {
		log.Printf("insert failed: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("The file stat have problem"))
		return
	}
	newID, _ := res.LastInsertId()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"id": newID})
}

func DeleteBooksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("The Method isn't delete"))
		return
	}

	idstr := r.PathValue("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("The string switch to int failed"))
		return
	}

	userID := r.Context().Value(middleware.UserIDKey)

	var filepath string
	query1 := "SELECT filepath FROM books WHERE id = ? AND user_id = ?"
	err = database.DB.QueryRow(query1, id, userID).Scan(&filepath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("The database didn't select"))
		return
	}

	query2 := "DELETE FROM books WHERE id = ? AND user_id = ?"
	result, err := database.DB.Exec(query2, id, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err = os.Remove(filepath)
	if err != nil && !os.IsNotExist(err) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func BooksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetBooksHandler(w, r)
	case http.MethodPost:
		UploadBooksHandler(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("The method not allowed"))
		return
	}
}

func GetBookContentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("The Method isn't Get"))
		return
	}

	//获取到地址中的id，并转换为int
	parts := strings.Split(r.URL.Path, "/")

	idstr := parts[len(parts)-2]
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("The string switch to int failed"))
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(int64)

	var filePath string
	query := "SELECT filepath FROM books WHERE id = ? AND user_id = ?"
	err = database.DB.QueryRow(query, id, userID).Scan(&filePath)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("The rows scan failed"))
		return
	}

	offset, err := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("The string switch to int failed"))
		return
	}
	length, err := strconv.ParseInt(r.URL.Query().Get("length"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("The string switch to int failed"))
		return
	}
	if length > 1<<20 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("The length too big"))
		return
	}

	validBuf, err := reader.ReadTXTChunk(filePath, offset, length)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("The readtxtchunk failed"))
		return
	}

	fi, err := os.Stat(filePath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("The stat failed"))
		return
	}

	end := offset+int64(len(validBuf)) >= fi.Size()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.ContentResp{Content: string(validBuf), IsEnd: end})

}
