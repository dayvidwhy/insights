package accounts

import (
	"errors"
	db "insights/db"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
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

func StoreUserAccount(
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

// Create a new user account
func CreateAccount(c echo.Context) error {
	u := new(User)
	if err := c.Bind(u); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid payload.")
	}

	if u.Email == "" || u.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Email or password is empty.")
	}

	err := StoreUserAccount(u.Email, u.Password)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error creating account.")
	}

	return c.JSON(http.StatusOK, &AccountResponse{
		Status:  "success",
		Message: "Account has been created.",
	})
}
