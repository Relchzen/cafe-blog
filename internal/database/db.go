package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() error {
	databaseUrl := os.Getenv("DATABASE_URL")

	if databaseUrl == "" {
		return fmt.Errorf("DATABASE URL environment variable is not set")
	}

	var err error
	DB, err = sql.Open("postgres", databaseUrl)

	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}

	log.Println("Database connected successfully")
	return nil
}

func CreateTables() error {
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT NOW()
	);`

	cafesTable := `
	CREATE TABLE IF NOT EXISTS cafes (
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		name VARCHAR(255) NOT NULL,
		location VARCHAR(255),
		rating INTEGER CHECK (rating >= 1 AND rating <= 5),
		notes TEXT,
		photo_url VARCHAR(500),
		visit_date DATE,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
	);`

	if _, err := DB.Exec(usersTable); err != nil {
		return fmt.Errorf("error creating users table: %w", err)
	}

	if _, err := DB.Exec(cafesTable); err != nil {
		return fmt.Errorf("error creating cafes table: %w", err)
	}

	log.Println("Tables created successfully")
	return nil
}
