package utils

import (
	"encoding/json"
	"fmt"
	"ithelp/models"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var secret = []byte(os.Getenv("JWT_SECRET"))

func GenerateJWT(userID int, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func ParseJWT(tokenStr string) (models.User, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return secret, nil
	})

	var user models.User
	if err != nil || !token.Valid {
		return user, fmt.Errorf("invalid token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		user.ID = int(claims["user_id"].(float64))
		user.Role = claims["role"].(string)

		// return full JSON string to avoid DB call in `me`
		bytes, _ := json.Marshal(user)
		json.Unmarshal(bytes, &user)
		return user, nil
	}

	return user, fmt.Errorf("invalid claims")
}