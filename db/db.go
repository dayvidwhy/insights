package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

// Increment the page view count for a URL
func IncrementPageView(db *sql.DB, url string) {
	fmt.Println("Incrementing page view for " + url)
	_, err := db.Exec(`
		INSERT INTO page_views (url, count)
		VALUES ($1, 1) ON CONFLICT (url)
		DO UPDATE SET count = page_views.count + 1`, url)
	if err != nil {
		panic(err)
	}
}

func FetchPageViews(db *sql.DB, url string) int {
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
func SetupDb() *sql.DB {
	connStr := "postgres://" +
		os.Getenv("POSTGRES_USER") + ":" +
		os.Getenv("POSTGRES_PASSWORD") + "@" +
		os.Getenv("POSTGRES_HOST") + "/" +
		os.Getenv("POSTGRES_DB") + "?sslmode=disable"

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
