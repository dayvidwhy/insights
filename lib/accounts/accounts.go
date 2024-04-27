package accounts

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	db "insights/db"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AccountResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type AccountTokenResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Token   string `json:"token"`
	TokenId int64  `json:"tokenId"`
}

// Create user accounts tables
func SetupAccounts() {
	_, err := db.Database.Exec(`CREATE TABLE IF NOT EXISTS accounts (
		id SERIAL PRIMARY KEY,
		email TEXT NOT NULL,
		password TEXT NOT NULL,
		UNIQUE(email)
	)`)
	if err != nil {
		log.Fatal(err)
	}

	// Setup table to store access tokens
	_, err = db.Database.Exec(`CREATE TABLE IF NOT EXISTS access_tokens (
		id SERIAL PRIMARY KEY,
		accountId INT NOT NULL,
		token TEXT NOT NULL,
		expiry BIGINT NOT NULL,
		createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatal(err)
	}
}

func GetAccountIdFromToken(token string) (int, error) {
	row := db.Database.QueryRow(`
		SELECT accountId FROM access_tokens
		WHERE token = $1`, token)

	var accountId int
	err := row.Scan(&accountId)
	if err != nil {
		return 0, errors.New("invalid token")
	}

	return accountId, nil
}

func LogInUser(email string, password string) (int, error) {
	// Check if the user exists
	row := db.Database.QueryRow(`
		SELECT email, password, id
		FROM accounts
		WHERE email = $1`, email)

	var queriedEmail string
	var queriedPassword string
	var queriedId int
	err := row.Scan(&queriedEmail, &queriedPassword, &queriedId)
	if err != nil {
		log.Println("Error logging in user id: " + email + "err: " + err.Error())
		return 0, err
	}

	// Check if the password matches
	if err := bcrypt.CompareHashAndPassword([]byte(queriedPassword), []byte(password)); err != nil {
		log.Println(err)
		return 0, err
	}

	return queriedId, nil
}

func CreateUserAccount(
	email string,
	password string,
) error {
	// Check if the user already exists
	row := db.Database.QueryRow(`
		SELECT email
		FROM accounts
		WHERE email = $1`, email)
	var queriedEmail string
	err := row.Scan(&queriedEmail)
	if err == nil {
		return errors.New("user already exists")
	}

	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// insert into db
	_, e := db.Database.Exec(`
		INSERT INTO accounts (email, password)
		VALUES ($1, $2)`,
		email, hashedPassword)
	if e != nil {
		return err
	}

	return nil
}

// Simple access token implementation
func generateToken(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

func CreateAccessToken(accountId int) (string, int64, error) {
	// Generate a random token
	token, err := generateToken(64)
	if err != nil {
		log.Println(err)
		log.Println("Error creating access token")
		return "", 0, err
	}

	// insert token into the db, set expiry to be 30 days from now
	row := db.Database.QueryRow(`
		INSERT INTO access_tokens (token, expiry, accountId)
		VALUES ($1, $2, $3)
		RETURNING id`,
		token,
		time.Now().AddDate(0, 0, 30).UTC().UnixMilli(),
		accountId,
	)

	var tokenId int64
	err = row.Scan(&tokenId)
	if err != nil {
		log.Println(err)
		return "", 0, err
	}

	return token, tokenId, nil
}

func RevokeAccessToken(accountId int, tokenId int) error {
	// validate that the user owns the token
	row := db.Database.QueryRow(`
		SELECT id FROM access_tokens
		WHERE id = $1 AND accountId = $2
	`, tokenId, accountId)

	var id int
	err := row.Scan(&id)
	if err != nil {
		return errors.New("issue revoking token")
	}

	// remove the token from the db
	_, err = db.Database.Exec(`
		DELETE FROM access_tokens
		WHERE id = $1 AND accountId = $2`,
		tokenId, accountId)
	if err != nil {
		return errors.New("issue revoking token")
	}

	return nil
}

// Validate whether the token is valid
func ValidateAccessToken(token string) error {
	row := db.Database.QueryRow(`
		SELECT token, expiry FROM access_tokens
		WHERE token = $1`,
		token)

	var queriedToken string
	var expiry int64
	err := row.Scan(&queriedToken, &expiry)
	if err != nil {
		return err
	}

	// check whether the token is still valid
	if expiry < time.Now().UTC().UnixMilli() {
		return errors.New("token expired")
	}

	return nil
}
