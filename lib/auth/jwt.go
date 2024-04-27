package auth

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	echojwt "github.com/labstack/echo-jwt/v4"
)

func validateClaims(token *jwt.Token) (jwt.MapClaims, error) {
	// Set claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		err := errors.New("issue processing token")
		log.Println(err)
		return nil, err
	}

	return claims, nil
}

// Create JWT token
func GenerateJWTToken(email string, accountId int) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims, err := validateClaims(token)
	if err != nil {
		return "", err
	}
	claims["email"] = email
	claims["accountId"] = accountId
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		log.Println(err)
		return "", err
	}

	return t, nil
}

// Extract the account ID from the JWT token
func GetAccountIdFromJwt(c echo.Context) (int, error) {
	token := c.Get("user").(*jwt.Token)

	claims, err := validateClaims(token)
	if err != nil {
		return 0, err
	}

	accountId := int(claims["accountId"].(float64))

	return accountId, nil
}

// Create JWT middleware
func GetJwtMiddleware() echo.MiddlewareFunc {
	return echojwt.JWT([]byte(os.Getenv("JWT_SECRET")))
}
