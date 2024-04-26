package auth

import (
	"errors"
	accounts "insights/lib/accounts"
	"strings"

	"github.com/labstack/echo/v4"
)

// Validate access token for posting pageviews from clients
func TokenAuth(c echo.Context) (string, error) {
	auth := c.Request().Header.Get("Authorization")

	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		return "", errors.New("authentication required")
	}
	token := strings.TrimPrefix(auth, "Bearer ")

	var err = accounts.ValidateAccessToken(token)
	if err != nil {
		return "", errors.New("invalid token")
	}

	return token, nil
}
