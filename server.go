package main

import (
	"net/http"
	"github.com/labstack/echo/v4"
)

type Response struct {
	Status string `json:"status"`

}

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Blank")
	})

	e.POST("/view", func(c echo.Context) error {
		response := &Response{
			Status: "Saved",
		}
		return c.JSON(http.StatusOK, response)
	})

	e.Logger.Fatal(e.Start(":1323"))
}