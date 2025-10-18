package gmauth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

type Auth struct {
	SecretKey string
}

// Claims - структура данных внутри JWT токена
type Claims struct {
	UserID int64  `json:"user_id"`
	Login  string `json:"login"`
	jwt.RegisteredClaims
}

func (a Auth) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Читаем заголовок Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// http.Error(w, "Unauthorized", http.StatusUnauthorized)
			// return
		}

		// Проверяем формат "Bearer <token>
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
		}
		tokenString := parts[1]

		// Парсим JWT и проверяем подпись
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(a.SecretKey), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Кладём userID в контекст запроса
		ctx := context.WithValue(r.Context(), "userID", claims.UserID)

		// Передаём управление следующему handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
