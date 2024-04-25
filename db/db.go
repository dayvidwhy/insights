package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// Package level variable to hold our DB pointer
var db *sql.DB

// Increment the page view count for a URL
func IncrementPageView(url string) error {
	_, err := db.Exec(`
		INSERT INTO page_views (url, count)
		VALUES ($1, 1) ON CONFLICT (url)
		DO UPDATE SET count = page_views.count + 1`, url)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO page_views_individual (url)
		VALUES ($1)
	`, url)

	return err
}

// Retrieve page views for a given URL, return 0 count if not found
func FetchPageViews(url string) int {
	fmt.Println("Fetch pageviews for: " + url)
	row := db.QueryRow(`
		SELECT count
		FROM page_views
		WHERE url = $1`, url)

	var count int
	err := row.Scan(&count)

	// If we don't find a record, return 0
	if err != nil {
		return 0
	}

	return count
}

type PageView struct {
	Time string `json:"time"`
}

type PageViews []struct {
	Time string `json:"time"`
}

func FetchPageViewsByDate(
	url string,
	start string,
	end string,
) (
	PageViews,
	error,
) {
	var pageViews PageViews
	rows, err := db.Query(`
		SELECT createdAt
		FROM page_views_individual
		WHERE url = $1 AND createdAt BETWEEN $2 AND $3
	`, url, start, end)

	if err != nil {
		return pageViews, err
	}

	// defer closing until we're done with the rows
	defer rows.Close()

	for rows.Next() {
		var pageView PageView
		err := rows.Scan(&pageView.Time)
		if err != nil {
			fmt.Println("Error fetching page views by date: ", err)
			return nil, err
		}
		pageViews = append(pageViews, pageView)
	}
	return pageViews, nil
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
		log.Fatal(err)
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to database successfully.")

	// Setup table to store overall pageviews
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS page_views (
		id SERIAL PRIMARY KEY,
		url TEXT NOT NULL,
		count INT NOT NULL DEFAULT 0,
		UNIQUE(url)
	)`)
	if err != nil {
		log.Fatal(err)
	}

	// Setup table to store individual pageviews
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS page_views_individual (
		id SERIAL PRIMARY KEY,
		url TEXT NOT NULL,
		createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatal(err)
	}
}
