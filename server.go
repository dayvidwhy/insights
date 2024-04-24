package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
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
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Blank")
	})

	e.POST("/views", func(c echo.Context) error {
		// get the JSON body from the request
		u := new(PageView)
		if err := c.Bind(u); err != nil {
			return err
		}

		fmt.Println("Tracking URL: " + u.Url)

		// build response
		response := map[string]string{
			"status": "success",
			"url":    u.Url,
		}

		if u.Url == "" {
			response["status"] = "Empty"
		}

		return c.JSON(http.StatusOK, response)
	})

	e.Logger.Fatal(e.Start(":1323"))
}
