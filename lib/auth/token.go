package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	accounts "insights/lib/accounts"
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type TokenHandler struct {
	store *TokenStore
}

func NewTokens(store *TokenStore) *TokenHandler {
	return &TokenHandler{store}
}

// Validate access token for posting pageviews from clients
func (th *TokenHandler) TokenAuth(c echo.Context) (string, error) {
	auth := c.Request().Header.Get("Authorization")

	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		err := errors.New("authentication required")
		log.Println(err)
		return "", err
	}
	token := strings.TrimPrefix(auth, "Bearer ")

	var err = th.store.validateAccessToken(token)
	if err != nil {
		log.Println(err)
		return "", errors.New("invalid token")
	}

	return token, nil
}

func (th *TokenHandler) GetAccountId(token string) (int, error) {
	return th.store.GetAccountIdFromToken(token)
}

// Handler for revoking a given access token
func (th *TokenHandler) RevokeAccessToken(c echo.Context) error {
	accountId, err := GetAccountIdFromJwt(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized.")
	}
	var tokenPayload struct {
		TokenId int `json:"tokenId"`
	}
	if err := c.Bind(&tokenPayload); err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid payload.")
	}

	err = th.store.deleteAccessToken(accountId, tokenPayload.TokenId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, &accounts.AccountResponse{
		Message: "Token has been revoked.",
	})
}

// Simple access token implementation
func generateToken(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		log.Println(err)
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

// Generate a client side access token
func (th *TokenHandler) GetAccessToken(c echo.Context) error {
	log.Println("Generating access token")
	accountId, err := GetAccountIdFromJwt(c)
	log.Println("Account ID: ", accountId)
	if err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized.")
	}

	token, tokenId, err := th.store.createAccessToken(accountId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error creating token.")
	}

	return c.JSON(http.StatusOK, &accounts.AccountTokenResponse{
		Message: "Authorized",
		Token:   token,
		TokenId: tokenId,
	})
}
