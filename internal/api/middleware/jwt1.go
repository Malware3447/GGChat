package middliware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

// Предполагаем, что CustomClaims и SecretToken определены
type CustomClaims struct {
	UserId int
	jwt.RegisteredClaims
}

// Измененная мидлварь
func JWTMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			cookieHeader, ok := r.Header["Cookie"]

			if ok && len(cookieHeader) > 0 {
			} else {
				fmt.Println("Сырой заголовок Cookie: НЕ НАЙДЕН в r.Header!")
			}
			fmt.Println("-------------------------------")

			cookie, err := r.Cookie("UserToken")

			if err != nil {
				fmt.Println("Ошибка r.Cookie('UserToken'): ", err)

				if ok && len(cookieHeader) > 0 {
					fmt.Println("Сырой заголовок 'Cookie' был, но r.Cookie() не нашла 'UserToken'.")
					for _, c := range r.Cookies() {
						fmt.Printf("    Найдено куки: Имя=%s, Значение (часть)=%s...\n", c.Name, c.Value[:10])
					}
				}

				http.Error(w, "Ошибка аутентификации", http.StatusUnauthorized)
				return
			}

			tokenString := cookie.Value

			claims := &CustomClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				fmt.Println("Ошибка валидации токена:", err)
				http.Error(w, "Invalid Token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), "user_id", claims.UserId)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
