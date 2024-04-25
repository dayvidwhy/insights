package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

// Package level variable to hold our DB pointer
var db *sql.DB

// Increment the page view count for a URL
func IncrementPageView(url string) {
	_, err := db.Exec(`
		INSERT INTO page_views (url, count)
		VALUES ($1, 1) ON CONFLICT (url)
		DO UPDATE SET count = page_views.count + 1`, url)
	if err != nil {
		panic(err)
	}

	_, e := db.Exec(`
		INSERT INTO page_views_individual (url)
		VALUES ($1)
	`, url)

	if e != nil {
		panic(e)
	}
}

// Retrieve page views for a given URL, return 0 count if not found
func FetchPageViews(url string) int {
	fmt.Println("Fetch pageviews for " + url)
	row := db.QueryRow(`
		SELECT count
		FROM page_views
		WHERE url = $1`, url)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

// Instantiate the database connection
func SetupDb() {
	// Create the connection from env variables
	connStr := "postgres://" +
		os.Getenv("POSTGRES_USER") + ":" +
		os.Getenv("POSTGRES_PASSWORD") + "@" +
		os.Getenv("POSTGRES_HOST") + "/" +
		os.Getenv("POSTGRES_DB") + "?sslmode=disable"

	// Setup our connection
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected to database")

	// Setup table to store overall pageviews
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS page_views (
		id SERIAL PRIMARY KEY,
		url TEXT NOT NULL,
		count INT NOT NULL DEFAULT 0,
		UNIQUE(url)
	)`)
	if err != nil {
		panic(err)
	}

	// Setup table to store individual pageviews
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS page_views_individual (
		id SERIAL PRIMARY KEY,
		url TEXT NOT NULL,
		createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		panic(err)
	}
}