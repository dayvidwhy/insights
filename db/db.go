package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// Instantiate the database connection
func SetupDb() *sql.DB {
	// Create the connection from env variables
	connStr := "postgres://" +
		os.Getenv("POSTGRES_USER") + ":" +
		os.Getenv("POSTGRES_PASSWORD") + "@" +
		os.Getenv("POSTGRES_HOST") + "/" +
		os.Getenv("POSTGRES_DB") + "?sslmode=disable"

	// Setup our connection
	var err error
	database, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Test the connection
	err = database.Ping()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to database successfully.")
	return database
}
