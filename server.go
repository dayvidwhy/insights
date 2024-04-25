package main

import (
	"fmt"
	database "insights/db"
	"net/http"

	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

// Response type for counting views
type ViewCountResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Url     string `json:"url"`
}

type ViewCountFetch struct {
	Status string `json:"status"`
	Views  int    `json:"views"`
	Url    string `json:"url"`
}

// Payload to POST /views
type PageView struct {
	Url string `json:"url"`
}

func main() {
	e := echo.New()
	database.SetupDb()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Blank")
	})

	e.POST("/views", func(c echo.Context) error {
		// get the JSON body from the request
		u := new(PageView)
		if err := c.Bind(u); err != nil {
			return err
		}

		if u.Url == "" {
			return c.JSON(http.StatusBadRequest, &ViewCountResponse{
				Status:  "fail",
				Message: "URL is empty.",
				Url:     "",
			})
		}

		fmt.Println("Tracking URL: " + u.Url)
		database.IncrementPageView(u.Url)

		return c.JSON(http.StatusOK, &ViewCountResponse{
			Status:  "success",
			Message: "URL has been tracked.",
			Url:     u.Url,
		})
	})

	e.GET("/views", func(c echo.Context) error {
		url := c.QueryParam("url")

		return c.JSON(http.StatusOK, &ViewCountFetch{
			Status: "success",
			Views:  database.FetchPageViews(url),
			Url:    url,
		})
	})

	e.Logger.Fatal(e.Start(":1323"))
}
