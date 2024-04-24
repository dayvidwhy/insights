package main

import (
	"fmt"
	database "insights/db"
	"net/http"

	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

type Response struct {
	Status string `json:"status"`
	Url    string `json:"url"`
}

type PageView struct {
	Url string `json:"url"`
}

func main() {
	e := echo.New()
	localDb := database.SetupDb()

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
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"status": "fail",
				"error":  "URL is empty.",
			})
		}

		fmt.Println("Tracking URL: " + u.Url)
		database.IncrementPageView(localDb, u.Url)

		return c.JSON(http.StatusOK, map[string]string{
			"status": "success",
			"url":    u.Url,
		})
	})

	e.GET("/views", func(c echo.Context) error {
		url := c.QueryParam("url")

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "success",
			"views":  database.FetchPageViews(localDb, url),
		})
	})

	e.Logger.Fatal(e.Start(":1323"))
}
