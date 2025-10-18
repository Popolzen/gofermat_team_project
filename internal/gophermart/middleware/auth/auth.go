package gmauth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// ============================================
// КОНТЕКСТ
// ============================================

type contextKey string

const userIDKey contextKey = "userID"

// ============================================
// СТРУКТУРЫ
// ============================================

type Auth struct {
	SecretKey string
}

// Claims - структура данных внутри JWT токена
type Claims struct {
	UserID int64  `json:"user_id"`
	Login  string `json:"login"`
	jwt.RegisteredClaims
}

// NewAuth создает новый экземпляр Auth
func NewAuth(secretKey string) *Auth {
	return &Auth{SecretKey: secretKey}
}

// AuthMiddleware проверяет JWT токен в каждом запросе
func (a *Auth) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Шаг 1: Читаем заголовок Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		//  Проверяем формат "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Парсим JWT и проверяем подпись
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
			// Проверяем алгоритм подписи (защита от атак)
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
		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)

		// Передаём управление следующему handler'у
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GenerateToken создаёт JWT токен для пользователя
func GenerateToken(userID int64, login string, secretKey string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Login:  login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 часа
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// GetUserID достаёт userID из контекста
func GetUserID(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDKey).(int64)
	return userID, ok
}

// MustGetUserID достаёт userID или паникует
func MustGetUserID(ctx context.Context) int64 {
	userID, ok := GetUserID(ctx)
	if !ok {
		panic("userID not found in context")
	}
	return userID
}
