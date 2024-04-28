package accounts

import (
	"database/sql"
	"errors"
	"log"

	"golang.org/x/crypto/bcrypt"
)

type AccountsStore struct {
	db *sql.DB
}

// Create user accounts tables
func SetupAccounts(db *sql.DB) *AccountsStore {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS accounts (
		id SERIAL PRIMARY KEY,
		email TEXT NOT NULL,
		password TEXT NOT NULL,
		UNIQUE(email)
	)`)
	if err != nil {
		log.Fatal(err)
	}

	// Setup table to store access tokens
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS access_tokens (
		id SERIAL PRIMARY KEY,
		accountId INT NOT NULL,
		token TEXT NOT NULL,
		expiry BIGINT NOT NULL,
		createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatal(err)
	}

	return &AccountsStore{db}
}

func (as *AccountsStore) storeUserAccount(
	email string,
	password string,
) error {
	// Check if the user already exists
	var queriedEmail string
	err := as.db.QueryRow(`
		SELECT email
		FROM accounts
		WHERE email = $1`, email).Scan(&queriedEmail)

	if err == nil {
		log.Println(err)
		return errors.New("user already exists")
	}

	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
		return err
	}

	// insert into db
	_, e := as.db.Exec(`
		INSERT INTO accounts (email, password)
		VALUES ($1, $2)`,
		email, hashedPassword)
	if e != nil {
		return err
	}

	return nil
}
