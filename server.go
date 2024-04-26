package main

import (
	"encoding/base64"
	"fmt"
	database "insights/db"
	accounts "insights/lib/accounts"
	views "insights/lib/views"
	"net/http"
	"strings"

	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

// Payload to POST /views
type PageView struct {
	Url string `json:"url"`
}

// Validate access token for posting pageviews from clients
func tokenAuth(c echo.Context) error {
	auth := c.Request().Header.Get("Authorization")

	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		return echo.NewHTTPError(http.StatusUnauthorized, "Authentication required")
	}
	token := strings.TrimPrefix(auth, "Bearer ")

	var err = accounts.ValidateAccessToken(token)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, &views.ViewCountResponse{
			Status:  "fail",
			Message: "Unauthorized",
			Url:     "",
		})
	}

	return nil
}

// Pull auth credentials off header
func extractAuth(c echo.Context) (string, string, error) {
	auth := c.Request().Header.Get("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Basic ") {
		return "", "", echo.NewHTTPError(http.StatusUnauthorized, "Authentication required")
	}
	userpass := strings.TrimPrefix(auth, "Basic ")
	decoded, err := base64.StdEncoding.DecodeString(userpass)
	if err != nil {
		return "", "", echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
	}
	creds := strings.Split(string(decoded), ":")
	if len(creds) != 2 {
		return "", "", echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
	}
	return creds[0], creds[1], nil
}

// Validate user credentials for fetching pageviews
func userAuth(c echo.Context) error {
	email, password, err := extractAuth(c)

	if err != nil {
		return c.JSON(http.StatusUnauthorized, &accounts.AccountResponse{
			Status:  "fail",
			Message: "Auth failed.",
			Email:   email,
		})
	}

	if err := accounts.LogInUser(email, password); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
	}

	return nil
}

func main() {
	e := echo.New()
	database.SetupDb()
	views.SetupViews()
	accounts.SetupAccounts()

	// Create a new page view.
	e.POST("/views/create", func(c echo.Context) error {
		if err := tokenAuth(c); err != nil {
			return err
		}

		u := new(PageView)
		if err := c.Bind(u); err != nil {
			return err
		}

		if u.Url == "" {
			return c.JSON(http.StatusBadRequest, &views.ViewCountResponse{
				Status:  "fail",
				Message: "URL is empty.",
				Url:     "",
			})
		}

		fmt.Println("Tracking URL: " + u.Url)
		views.IncrementPageView(u.Url)

		return c.JSON(http.StatusOK, &views.ViewCountResponse{
			Status:  "success",
			Message: "URL has been tracked.",
			Url:     u.Url,
		})
	})

	// Fetch page views for a given URL.
	e.GET("/views/count", func(c echo.Context) error {
		if err := userAuth(c); err != nil {
			return err
		}

		url := c.QueryParam("url")
		pageViews := views.FetchPageViews(url)

		return c.JSON(http.StatusOK, &views.ViewsCountFetch{
			Status: "success",
			Views:  pageViews,
			Url:    url,
		})
	})

	/*
	 * Fetch page views by date.
	 * Example: /views/counts?url=[url]&start=2021-01-01&end=2021-01-31
	 */
	e.GET("/views/counts", func(c echo.Context) error {
		if err := userAuth(c); err != nil {
			return err
		}

		url := c.QueryParam("url")
		start := c.QueryParam("start")
		end := c.QueryParam("end")
		pageViews, err := views.FetchPageViewsByDate(url, start, end)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, &views.ViewsCountFetchByDate{
				Status: "fail",
				Start:  start,
				End:    end,
				Url:    url,
				Views:  nil,
			})
		}

		return c.JSON(http.StatusOK, &views.ViewsCountFetchByDate{
			Status: "success",
			Start:  start,
			End:    end,
			Url:    url,
			Views:  pageViews,
		})
	})

	// Create a new user
	e.POST("/accounts/create", func(c echo.Context) error {
		u := new(accounts.User)
		if err := c.Bind(u); err != nil {
			return err
		}

		if u.Email == "" || u.Password == "" {
			return c.JSON(http.StatusBadRequest, &accounts.AccountResponse{
				Status:  "fail",
				Message: "Email or password is empty.",
				Email:   u.Email,
			})
		}

		err := accounts.CreateUserAccount(u.Email, u.Password)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, &accounts.AccountResponse{
				Status:  "fail",
				Message: "Error creating account.",
				Email:   u.Email,
			})
		}

		return c.JSON(http.StatusOK, &accounts.AccountResponse{
			Status:  "success",
			Message: "Account has been created.",
			Email:   u.Email,
		})
	})

	e.GET("/accounts/token", func(c echo.Context) error {
		if err := userAuth(c); err != nil {
			return err
		}

		email, _, _ := extractAuth(c)
		token, err := accounts.CreateAccessToken(email)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &accounts.AccountResponse{
				Status:  "fail",
				Message: "Error creating token.",
				Email:   email,
			})
		}

		return c.JSON(http.StatusOK, &accounts.AccountTokenResponse{
			Status:  "success",
			Message: "Authorized",
			Email:   email,
			Token:   token,
		})
	})

	e.Logger.Fatal(e.Start(":1323"))
}
