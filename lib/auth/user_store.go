package auth

import (
	"database/sql"
	"log"

	"golang.org/x/crypto/bcrypt"
)

type UserStore struct {
	db *sql.DB
}

func SetupUsers(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

func (as *UserStore) checkUserLogin(email string, password string) (int, error) {
	// Check if the user exists
	var queriedEmail string
	var queriedPassword string
	var queriedId int
	err := as.db.QueryRow(`
		SELECT email, password, id
		FROM accounts
		WHERE email = $1`, email).Scan(&queriedEmail, &queriedPassword, &queriedId)

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
