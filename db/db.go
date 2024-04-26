package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// Package level variable to hold our DB pointer
var Database *sql.DB

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
	Database, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Test the connection
	err = Database.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to database successfully.")
}
