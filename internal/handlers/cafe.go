package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/relchzen/cafe-blog/internal/database"
	"github.com/relchzen/cafe-blog/internal/middleware"
	"github.com/relchzen/cafe-blog/internal/models"
	"github.com/relchzen/cafe-blog/internal/utils"
)

func ValidateCafeRequest(req models.CreateCafeRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}

	if req.Location == "" {
		return fmt.Errorf("location is required")
	}

	if req.Rating != nil {
		if *req.Rating < 1 || *req.Rating > 5 {
			return fmt.Errorf("rating must be between 1 and 5")
		}
	}

	if req.VisitDate != nil && *req.VisitDate != "" {
		_, err := utils.ValidateAndParseDate(*req.VisitDate)
		if err != nil {
			return err
		}
	}

	return nil
}

func ValidateUpdateCafeRequest(req models.UpdateCafeRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}

	if req.Location == "" {
		return fmt.Errorf("location is required")
	}

	if req.Rating != nil {
		if *req.Rating < 1 || *req.Rating > 5 {
			return fmt.Errorf("rating must be between 1 and 5")
		}
	}

	if req.VisitDate != nil && *req.VisitDate != "" {
		_, err := utils.ValidateAndParseDate(*req.VisitDate)
		if err != nil {
			return err
		}
	}

	return nil
}

/*
CreateCafe method to create a cafe record.
*/
func CreateCafe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userId, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Invalid user context", http.StatusInternalServerError)
		return
	}

	var req models.CreateCafeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := ValidateCafeRequest(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var cafeID int
	var createdAt time.Time
	query := `INSERT INTO cafes(user_id, name, location, rating, notes, visit_date) 
	VALUES ($1, $2, $3, $4, $5, $6) 
	RETURNING id, created_at`
	err := database.DB.QueryRow(query, userId, req.Name, req.Location, req.Rating, req.Notes, req.VisitDate).Scan(&cafeID, &createdAt)
	if err != nil {
		http.Error(w, "Failed to create new cafe record", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := map[string]interface{}{
		"message":    "Cafe visit created successfully",
		"cafe_id":    cafeID,
		"created_at": createdAt,
	}

	json.NewEncoder(w).Encode(response)
}

func ListCafes(w http.ResponseWriter, r *http.Request) {
	// Check method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(int)

	query := `SELECT id, user_id, name, location, rating, notes, photo_url, visit_date, created_at, updated_at
		FROM cafes
		WHERE user_id = $1
		ORDER BY visit_date DESC
	`

	rows, err := database.DB.Query(query, userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
	}

	defer rows.Close()

	var cafes []models.Cafe
	for rows.Next() {
		var cafe models.Cafe
		err := rows.Scan(&cafe.ID,
			&cafe.UserID,
			&cafe.Name,
			&cafe.Location,
			&cafe.Rating,
			&cafe.Notes,
			&cafe.PhotoURL,
			&cafe.VisitDate,
			&cafe.CreatedAt,
			&cafe.UpdatedAt,
		)
		if err != nil {
			fmt.Print(err)
			http.Error(w, "Error reading data", http.StatusInternalServerError)
		}

		cafes = append(cafes, cafe)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cafes)
}

func UpdateCafe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idString := r.PathValue("id")
	if idString == "" {
		http.Error(w, "Missing parameter value", http.StatusBadRequest)
		return
	}

	cafeID, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "Invalid cafe ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(int)

	var req models.UpdateCafeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := ValidateUpdateCafeRequest(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var existingUserID int
	checkQuery := `SELECT user_id FROM cafes WHERE id = $1`
	err = database.DB.QueryRow(checkQuery, idString).Scan(&existingUserID)
	if err == sql.ErrNoRows {
		http.Error(w, "Cafe not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if existingUserID != userID {
		http.Error(w, "Unauthorized - this cafe belongs to another user", http.StatusForbidden)
		return
	}

	updateQuery := `UPDATE cafes
					SET name=$1, location=$2, rating=$3, notes=$4, photo_url= $5, visit_date=$6, updated_at=NOW()
					WHERE id=$7
					RETURNING id, updated_at`

	var updatedID int
	var updatedAt time.Time
	err = database.DB.QueryRow(updateQuery,
		req.Name,
		req.Location,
		req.Rating,
		req.Notes,
		req.PhotoURL,
		req.VisitDate,
		cafeID,
	).Scan(&updatedID, &updatedAt)

	if err != nil {
		http.Error(w, "Failed to update cafe", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Cafe updated successfully",
		"cafe_id":    updatedID,
		"updated_at": updatedAt,
	})
}
