package main

import (
	database "insights/db"
	accounts "insights/lib/accounts"
	auth "insights/lib/auth"
	views "insights/lib/views"
	"log"
	"net/http"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

func main() {
	e := echo.New()
	database.SetupDb()
	views.SetupViews()
	accounts.SetupAccounts()

	// Create a new page view.
	e.POST("/views/create", func(c echo.Context) error {
		token, err := auth.TokenAuth(c)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
		}

		u := new(views.PageViewSubmit)
		if err := c.Bind(u); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid payload.")
		}

		if u.Url == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "URL is empty.")
		}

		// Get the account ID from the token
		accountId, err := accounts.GetAccountIdFromToken(token)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
		}

		log.Println("Tracking URL: " + u.Url)
		views.IncrementPageView(accountId, u.Url)

		return c.JSON(http.StatusOK, &views.ViewCountResponse{
			Status:  "success",
			Message: "URL has been tracked.",
			Url:     u.Url,
		})
	})

	// Fetch page views for a given URL.
	e.GET("/views/count", func(c echo.Context) error {
		accountId, err := auth.UserAuth(c)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
		}

		url := c.QueryParam("url")
		pageViews := views.FetchPageViews(accountId, url)

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
		accountId, err := auth.UserAuth(c)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
		}

		url := c.QueryParam("url")
		start := c.QueryParam("start")
		end := c.QueryParam("end")
		_, err = time.Parse("2006-01-02 15:04:05.000", start)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid start date:"+start+". Please use UTC in the format: yyyy-mm-dd hh:mm:ss.fff")
		}
		_, err = time.Parse("2006-01-02 15:04:05.000", end)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid start date:"+end+". Please use UTC in the format: yyyy-mm-dd hh:mm:ss.fff")
		}
		pageViews, err := views.FetchPageViewsByDate(accountId, url, start, end)

		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching page views by date.")
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
			return echo.NewHTTPError(http.StatusBadRequest, "Email or password is empty.")
		}

		err := accounts.CreateUserAccount(u.Email, u.Password)

		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Error creating account.")
		}

		return c.JSON(http.StatusOK, &accounts.AccountResponse{
			Status:  "success",
			Message: "Account has been created.",
			Email:   u.Email,
		})
	})

	e.GET("/accounts/token", func(c echo.Context) error {
		accountId, err := auth.UserAuth(c)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
		}

		token, err := accounts.CreateAccessToken(accountId)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Error creating token.")
		}

		return c.JSON(http.StatusOK, &accounts.AccountTokenResponse{
			Status:  "success",
			Message: "Authorized",
			Token:   token,
		})
	})

	e.Logger.Fatal(e.Start(":1323"))
}
