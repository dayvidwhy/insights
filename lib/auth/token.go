package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	accounts "insights/lib/accounts"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// Validate whether the token is valid
func (as *AuthStore) validateAccessToken(token string) error {
	var queriedToken string
	var expiry int64
	err := as.db.QueryRow(`
		SELECT token, expiry FROM access_tokens
		WHERE token = $1`,
		token).Scan(&queriedToken, &expiry)

	if err != nil {
		log.Println(err)
		return err
	}

	// check whether the token is still valid
	if expiry < time.Now().UTC().UnixMilli() {
		return errors.New("token expired")
	}

	return nil
}

// Validate access token for posting pageviews from clients
func (ah *AuthHandler) TokenAuth(c echo.Context) (string, error) {
	auth := c.Request().Header.Get("Authorization")

	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		err := errors.New("authentication required")
		log.Println(err)
		return "", err
	}
	token := strings.TrimPrefix(auth, "Bearer ")

	var err = ah.store.validateAccessToken(token)
	if err != nil {
		log.Println(err)
		return "", errors.New("invalid token")
	}

	return token, nil
}

func (as *AuthStore) GetAccountIdFromToken(token string) (int, error) {
	var accountId int
	err := as.db.QueryRow(`
		SELECT accountId FROM access_tokens
		WHERE token = $1`, token).Scan(&accountId)

	if err != nil {
		log.Println(err)
		return 0, errors.New("invalid token")
	}

	return accountId, nil
}

func (ah *AuthHandler) GetAccountId(token string) (int, error) {
	return ah.store.GetAccountIdFromToken(token)
}

// Remove the access token from the database
func (as *AuthStore) deleteAccessToken(accountId int, tokenId int) error {
	// validate that the user owns the token
	var id int
	err := as.db.QueryRow(`
		SELECT id FROM access_tokens
		WHERE id = $1 AND accountId = $2
	`, tokenId, accountId).Scan(&id)

	if err != nil {
		log.Println(err)
		return errors.New("issue revoking token")
	}

	// remove the token from the db
	_, err = as.db.Exec(`
		DELETE FROM access_tokens
		WHERE id = $1 AND accountId = $2`,
		tokenId, accountId)
	if err != nil {
		log.Println(err)
		return errors.New("issue revoking token")
	}

	return nil
}

// Handler for revoking a given access token
func (ah *AuthHandler) RevokeAccessToken(c echo.Context) error {
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

	err = ah.store.deleteAccessToken(accountId, tokenPayload.TokenId)
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

func (as *AuthStore) createAccessToken(accountId int) (string, int64, error) {
	// Generate a random token
	token, err := generateToken(64)
	if err != nil {
		return "", 0, err
	}

	// insert token into the db, set expiry to be 30 days from now
	var tokenId int64
	err = as.db.QueryRow(`
		INSERT INTO access_tokens (token, expiry, accountId)
		VALUES ($1, $2, $3)
		RETURNING id`,
		token,
		time.Now().AddDate(0, 0, 30).UTC().UnixMilli(),
		accountId,
	).Scan(&tokenId)

	if err != nil {
		log.Println(err)
		return "", 0, err
	}

	return token, tokenId, nil
}

// Generate a client side access token
func (ah *AuthHandler) GetAccessToken(c echo.Context) error {
	log.Println("Generating access token")
	accountId, err := GetAccountIdFromJwt(c)
	log.Println("Account ID: ", accountId)
	if err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized.")
	}

	token, tokenId, err := ah.store.createAccessToken(accountId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error creating token.")
	}

	return c.JSON(http.StatusOK, &accounts.AccountTokenResponse{
		Message: "Authorized",
		Token:   token,
		TokenId: tokenId,
	})
}
