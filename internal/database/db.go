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

	cafePhotosTable := `
    CREATE TABLE IF NOT EXISTS cafe_photos (
        id SERIAL PRIMARY KEY,
        cafe_id INTEGER REFERENCES cafes(id) ON DELETE CASCADE,
        photo_url VARCHAR(500) NOT NULL,
        display_order INTEGER DEFAULT 0,
        created_at TIMESTAMP DEFAULT NOW()
    );`

	cafePhotosIndex := `
    CREATE INDEX IF NOT EXISTS idx_cafe_photos_cafe_id ON cafe_photos(cafe_id);`

	cafesIndex := `
    CREATE INDEX IF NOT EXISTS idx_cafes_user_id ON cafes(user_id);`

	// Execute table creation in order
	if _, err := DB.Exec(usersTable); err != nil {
		return fmt.Errorf("error creating users table: %w", err)
	}

	if _, err := DB.Exec(cafesTable); err != nil {
		return fmt.Errorf("error creating cafes table: %w", err)
	}

	if _, err := DB.Exec(cafePhotosTable); err != nil {
		return fmt.Errorf("error creating cafe_photos table: %w", err)
	}

	// Create indexes
	if _, err := DB.Exec(cafePhotosIndex); err != nil {
		return fmt.Errorf("error creating cafe_photos index: %w", err)
	}

	if _, err := DB.Exec(cafesIndex); err != nil {
		return fmt.Errorf("error creating cafes index: %w", err)
	}

	log.Println("Tables created successfully")
	return nil
}
