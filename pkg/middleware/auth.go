package middleware

import (
	"context"
	"homelab-reader/pkg/database"
	"net/http"
	"time"
)

type contextKey string

const UserIDKey contextKey = "userID"

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Don't find the right session_token"))
			return
		}

		var userID int64
		var expiresAt time.Time
		//定义一个查询语句，并直接开始查询到所需要的userID
		query1 := "SELECT user_id,expires_at FROM sessions WHERE token = ?"
		err = database.DB.QueryRow(query1, cookie.Value).Scan(&userID, &expiresAt)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Don't find the token"))
			return
		}

		if time.Now().After(expiresAt) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("The authentication failed"))
			return
		}

		//规定一个有效期低于3天时，在自动更新expires_at这个过期时间
		dif := time.Until(expiresAt)
		if dif <= 72*time.Hour {
			expiresAt = time.Now().Add(time.Hour * 168)
			query2 := "UPDATE sessions SET expires_at = ? WHERE token = ?"
			_, err = database.DB.Exec(query2, expiresAt, cookie.Value)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("The expires_at failed update"))
				return
			}
		}

		//构建出context,将我们从数据库中的userID值保存到context中，然后将context打包进r中，给下一个handler
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))

	}
}
