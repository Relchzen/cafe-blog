package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/relchzen/cafe-blog/internal/database"
	"github.com/relchzen/cafe-blog/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func generateJWT(userId int, email string) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", fmt.Errorf("JWT_SECRET environment variable is not set")
	}

	claims := jwt.MapClaims{
		"user_id": userId,
		"email":   email,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func Register(w http.ResponseWriter, r *http.Request) {
	// Validate the method is a POST request
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode JSON to RegisterRequest struct
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate if email and password value exists
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Hash password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error processing password", http.StatusInternalServerError)
		return
	}

	// Insert user and scans the return id into UserID
	var UserID int
	query := `INSERT INTO users(email, password_hash) VALUES ($1, $2) RETURNING id`
	err = database.DB.QueryRow(query, req.Email, string(hashedPassword)).Scan(&UserID)
	if err != nil {
		http.Error(w, "Email already exists", http.StatusConflict)
	}

	// Set Response header to indicate sending json
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Create and send response
	response := map[string]interface{}{
		"message": "User reqistered successfully",
		"user_id": UserID,
	}

	json.NewEncoder(w).Encode(response)
}

func Login(w http.ResponseWriter, r *http.Request) {
	// Validate the method is a POST request
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode JSON to LoginRequest struct
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate if email and password value exists
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	var user models.User
	query := `SELECT id, email, password_hash FROM users WHERE email = $1`
	row := database.DB.QueryRow(query, req.Email)
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	token, err := generateJWT(user.ID, user.Email)

	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login Successful",
		"token":   token,
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
		},
	})
}
