package router

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

func parseJWT(secretKey string, token string) (string, error) {
	tokenClaims, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := tokenClaims.Claims.(jwt.MapClaims); ok && tokenClaims.Valid {
		//check exp
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				return "", fmt.Errorf("token is expired")
			}
		}
		return claims["user"].(string), nil
	}

	return "", fmt.Errorf("invalid token")
}
func generateJWT(secretKey string, user string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": user,
		"exp":  time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	// Подпись токена (замените "secret" на ваш ключ)
	jwtString, _ := token.SignedString([]byte(secretKey))
	return jwtString
}
