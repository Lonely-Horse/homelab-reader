package handlers

import (
	"database/sql"
	"encoding/json"
	"homelab-reader/bootstrap"
	"homelab-reader/pkg/auth"
	"homelab-reader/pkg/models"
	"net/http"
	"time"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("The method isn't Post"))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1024)

	var req models.RegisterReq
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("The decoder failed"))
		return
	}

	if req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("The password is empty"))
		return
	}

	//加密密码的同时，将用户名和哈希密码写入数据库
	pwd_hash := auth.HashPassword(req.Password)
	insert := "INSERT INTO users (username,password_hash) VALUES (?,?)"
	_, err = bootstrap.DB.Exec(insert, req.Username, pwd_hash)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("The data didn't insert"))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Successful registry!"))
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("The method isn't Post"))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1024)

	//解析r.body里面的请求体内容
	var req models.LoginReq
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("The decode failed"))
		return
	}

	var storedID int64
	var storedHash string
	query1 := "SELECT id,password_hash FROM users WHERE username = ?"
	err = bootstrap.DB.QueryRow(query1, req.Username).Scan(&storedID, &storedHash)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Invalid username or password"))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("The database didn't select now"))
		return
	}

	if !auth.CheckPassword(req.Password, storedHash) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("The password isn't right"))
		return
	}

	session_token, err := auth.GenerateToken()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("The token didn't generate"))
		return
	}

	//设定过期时间为7天
	expiresAt := time.Now().Add(time.Hour * 168)

	//在数据库中，保存随机值和过期时间
	query2 := "INSERT INTO sessions (user_id,token,expires_at) VALUES (?,?,?)"
	_, err = bootstrap.DB.Exec(query2, storedID, session_token, expiresAt)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("The database insert failed"))
		return
	}

	//定义cookie配置，在http层次保存相关的随机值和过期时间，开启httponly
	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    session_token,
		Expires:  expiresAt,
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, cookie)
}
