package router

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

func parseJWT(secretKey string, token string) (string, error) {
	tokenClaims, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := tokenClaims.Claims.(jwt.MapClaims); ok && tokenClaims.Valid {
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				return "", fmt.Errorf("token is expired")
			}
		}
		user, _ := claims["user"].(string)
		if user == "" {
			return "", fmt.Errorf("invalid token")
		}
		return user, nil
	}

	return "", fmt.Errorf("invalid token")
}

func generateJWT(secretKey string, user string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": user,
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	})
	jwtString, _ := token.SignedString([]byte(secretKey))
	return jwtString
}

func generateFileToken(secretKey, user, hash string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user":    user,
		"hash":    hash,
		"purpose": "file",
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	})
	return token.SignedString([]byte(secretKey))
}

func parseFileToken(secretKey, tokenString, expectedHash string) (string, error) {
	tokenClaims, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := tokenClaims.Claims.(jwt.MapClaims)
	if !ok || !tokenClaims.Valid {
		return "", fmt.Errorf("invalid token")
	}
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return "", fmt.Errorf("token is expired")
		}
	}
	purpose, _ := claims["purpose"].(string)
	if purpose != "file" {
		return "", fmt.Errorf("invalid token")
	}
	hash, _ := claims["hash"].(string)
	if hash != expectedHash {
		return "", fmt.Errorf("invalid token")
	}
	user, _ := claims["user"].(string)
	if user == "" {
		return "", fmt.Errorf("invalid token")
	}
	return user, nil
}

func parseMapClaims(secretKey, tokenString string) (jwt.MapClaims, error) {
	tokenClaims, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := tokenClaims.Claims.(jwt.MapClaims)
	if !ok || !tokenClaims.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func signMapClaims(secretKey string, claims jwt.MapClaims) (string, error) {
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(time.Hour).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}
