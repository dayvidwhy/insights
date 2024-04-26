package main

import (
	"fmt"
	database "insights/db"
	accounts "insights/lib/accounts"
	views "insights/lib/views"
	"log"
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

func auth(c echo.Context) error {
	log.Println("Authenticating request")
	auth := c.Request().Header.Get("Authorization")
	log.Println("auth was: " + auth)
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		log.Println("No auth token")
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

func main() {
	e := echo.New()
	database.SetupDb()
	views.SetupViews()
	accounts.SetupAccounts()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Blank")
	})

	/*
	* Create a new page view.
	 */
	e.POST("/views/create", func(c echo.Context) error {
		err := auth(c)
		if err != nil {
			return err
		}

		// get the JSON body from the request
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

	/*
	* Fetch page views for a given URL.
	 */
	e.GET("/views/count", func(c echo.Context) error {
		err := auth(c)
		if err != nil {
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
		err := auth(c)
		if err != nil {
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

	e.POST("/accounts/create", func(c echo.Context) error {
		// get the JSON body from the request
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

	e.POST("/accounts/token", func(c echo.Context) error {
		u := new(accounts.User)
		if err := c.Bind(u); err != nil {
			return err
		}

		var err error
		err = accounts.LogInUser(u.Email, u.Password)

		if err != nil {
			log.Println(err)
			return c.JSON(http.StatusUnauthorized, &accounts.AccountResponse{
				Status:  "fail",
				Message: "Unauthorized",
				Email:   u.Email,
			})
		}

		token, err := accounts.CreateAccessToken(u.Email)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &accounts.AccountResponse{
				Status:  "fail",
				Message: "Error creating token.",
				Email:   u.Email,
			})
		}

		return c.JSON(http.StatusOK, &accounts.AccountTokenResponse{
			Status:  "success",
			Message: "Authorized",
			Email:   u.Email,
			Token:   token,
		})
	})

	e.Logger.Fatal(e.Start(":1323"))
}
