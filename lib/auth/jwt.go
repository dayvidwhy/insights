package auth

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	echojwt "github.com/labstack/echo-jwt/v4"
)

// Create JWT token
func GenerateJWTToken(email string, accountId int) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["email"] = email
	claims["accountId"] = accountId
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}

	return t, nil
}

// Extract the account ID from the JWT token
func GetAccountIdFromJwt(c echo.Context) (int, error) {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	accountId := int(claims["accountId"].(float64))
	return accountId, nil
}

// Create JWT middleware
func GetJwtMiddleware() echo.MiddlewareFunc {
	return echojwt.JWT([]byte(os.Getenv("JWT_SECRET")))
}
