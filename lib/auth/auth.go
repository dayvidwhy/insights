package auth

import (
	"encoding/base64"
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

// Pull auth credentials off header
func ExtractAuth(c echo.Context) (string, string, error) {
	auth := c.Request().Header.Get("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Basic ") {
		return "", "", errors.New("authentication required")
	}
	userpass := strings.TrimPrefix(auth, "Basic ")
	decoded, err := base64.StdEncoding.DecodeString(userpass)
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}
	creds := strings.Split(string(decoded), ":")
	if len(creds) != 2 {
		return "", "", errors.New("invalid credentials")
	}
	return creds[0], creds[1], nil
}

// Validate user credentials for fetching pageviews
func UserAuth(c echo.Context) (int, error) {
	email, password, err := ExtractAuth(c)

	if err != nil {
		return 0, err
	}

	accountId, err := accounts.LogInUser(email, password)
	if err != nil {
		return 0, errors.New("login failed")
	}

	return accountId, nil
}