package auth

import (
	accounts "insights/lib/accounts"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	store *UserStore
}

func NewUsers(store *UserStore) *UserHandler {
	return &UserHandler{store: store}
}

// Validate user credentials and return a JWT token
func (ah *UserHandler) LoginUser(c echo.Context) error {
	u := new(accounts.User)
	if err := c.Bind(u); err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid credentials.")
	}

	accountId, err := ah.store.checkUserLogin(u.Email, u.Password)
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
