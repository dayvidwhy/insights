package accounts

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AccountResponse struct {
	Message string `json:"message"`
}

type AccountTokenResponse struct {
	Message string `json:"message"`
	Token   string `json:"token"`
	TokenId int64  `json:"tokenId"`
}

type AccountsHandler struct {
	store *AccountsStore
}

func NewAccounts(store *AccountsStore) *AccountsHandler {
	return &AccountsHandler{store: store}
}

// Create a new user account
func (ah *AccountsHandler) CreateAccount(c echo.Context) error {
	u := new(User)
	if err := c.Bind(u); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid payload.")
	}

	if u.Email == "" || u.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Email or password is empty.")
	}

	err := ah.store.storeUserAccount(u.Email, u.Password)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error creating account.")
	}

	return c.JSON(http.StatusOK, &AccountResponse{
		Message: "Account has been created.",
	})
}
