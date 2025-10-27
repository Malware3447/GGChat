package api

import (
	"GGChat/internal/config"
	"GGChat/internal/interfaces"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Jwt struct {
	cfg *config.Config
}

type CustomClaims struct {
	UserId int
	jwt.RegisteredClaims
}

func NewJwt(cfg *config.Config) *Jwt {
	return &Jwt{
		cfg: cfg,
	}
}

func (j *Jwt) NewToken(UserId int) (string, error) {
	expirationTime := time.Now().Add(15 * time.Minute)

	usid := strconv.Itoa(UserId)
	claims := CustomClaims{
		UserId: UserId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "my-golang-app",
			Subject:   usid,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(j.cfg.Jwt.SecretToken))
	if err != nil {
		return "", err
	}

	fmt.Println("-------Token: ", tokenString)

	return tokenString, nil
}

var _ interfaces.JwtInterface = (*Jwt)(nil)
