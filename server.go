package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

func incrementPageView(db *sql.DB, url string) {
	fmt.Println("Incrementing page view for " + url)
	_, err := db.Exec(`INSERT INTO page_views (url, count)
		VALUES ($1, 1) ON CONFLICT (url)
		DO UPDATE SET count = page_views.count + 1`,
		url)
	if err != nil {
		panic(err)
	}
}

func setupDb() *sql.DB {
	connStr := "postgres://" +
		os.Getenv("POSTGRES_USER") + ":" +
		os.Getenv("POSTGRES_PASSWORD") + "@" +
		os.Getenv("POSTGRES_HOST") + "/" +
		os.Getenv("POSTGRES_DB") + "?sslmode=disable"
	fmt.Println(connStr)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected to database")
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS page_views (
		id SERIAL PRIMARY KEY,
		url TEXT NOT NULL,
		count INT NOT NULL DEFAULT 0,
		UNIQUE(url)
	)`)
	if err != nil {
		panic(err)
	}
	return db
}

type Response struct {
	Status string `json:"status"`
	Url    string `json:"url"`
}

type PageView struct {
	Url string `json:"url"`
}

func main() {
	e := echo.New()
	db := setupDb()

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
		incrementPageView(db, u.Url)

		return c.JSON(http.StatusOK, map[string]string{
			"status": "success",
			"url":    u.Url,
		})
	})

	e.GET("/views", func(c echo.Context) error {
		url := c.QueryParam("url")
		row := db.QueryRow("SELECT count FROM page_views WHERE url = $1", url)
		if row.Err() != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"status": "fail",
				"error":  "URL is not tracked.",
			})
		}

		var count int
		err := row.Scan(&count)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"status": "fail",
				"error":  "Internal error.",
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "success",
			"views":  count,
		})
	})

	e.Logger.Fatal(e.Start(":1323"))
}
