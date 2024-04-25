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

type ViewCountFetchByDate struct {
	Status string `json:"status"`
	Start  string `json:"start"`
	End    string `json:"end"`
	Url    string `json:"url"`
	Views  []struct {
		Time string `json:"time"`
	}
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

	/*
	* Create a new page view.
	 */
	e.POST("/views/create", func(c echo.Context) error {
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

	/*
	* Fetch page views for a given URL.
	 */
	e.GET("/views/count", func(c echo.Context) error {
		url := c.QueryParam("url")

		return c.JSON(http.StatusOK, &ViewCountFetch{
			Status: "success",
			Views:  database.FetchPageViews(url),
			Url:    url,
		})
	})

	/*
	 * Fetch page views by date.
	 * Example: /views/counts?url=[url]&start=2021-01-01&end=2021-01-31
	 */
	e.GET("/views/counts", func(c echo.Context) error {
		url := c.QueryParam("url")
		start := c.QueryParam("start")
		end := c.QueryParam("end")
		pageViews := database.FetchPageViewsByDate(url, start, end)
		return c.JSON(http.StatusOK, &ViewCountFetchByDate{
			Status: "success",
			Start:  start,
			End:    end,
			Url:    url,
			Views:  pageViews,
		})
	})

	e.Logger.Fatal(e.Start(":1323"))
}
