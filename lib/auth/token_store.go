package auth

import (
	"database/sql"
	"errors"
	"log"
	"time"
)

type TokenStore struct {
	db *sql.DB
}

func SetupTokens(db *sql.DB) *TokenStore {
	// Setup table to store access tokens
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS access_tokens (
		id SERIAL PRIMARY KEY,
		accountId INT NOT NULL,
		token TEXT NOT NULL,
		expiry BIGINT NOT NULL,
		createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatal(err)
	}

	return &TokenStore{db}
}

// Validate whether the token is valid
func (as *TokenStore) validateAccessToken(token string) error {
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

func (as *TokenStore) GetAccountIdFromToken(token string) (int, error) {
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

// Remove the access token from the database
func (as *TokenStore) deleteAccessToken(accountId int, tokenId int) error {
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

func (th *TokenStore) createAccessToken(accountId int) (string, int64, error) {
	// Generate a random token
	token, err := generateToken(64)
	if err != nil {
		return "", 0, err
	}

	// insert token into the db, set expiry to be 30 days from now
	var tokenId int64
	err = th.db.QueryRow(`
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
