package auth

import (
	db "insights/db"
	accounts "insights/lib/accounts"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func checkUserLogin(email string, password string) (int, error) {
	// Check if the user exists
	var queriedEmail string
	var queriedPassword string
	var queriedId int
	err := db.Database.QueryRow(`
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

// Validate user credentials and return a JWT token
func LoginUser(c echo.Context) error {
	u := new(accounts.User)
	if err := c.Bind(u); err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid credentials.")
	}

	accountId, err := checkUserLogin(u.Email, u.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials.")
	}

	// Create token
	token, err := GenerateJWTToken(u.Email, accountId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error creating token.")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"token": token,
	})
}
