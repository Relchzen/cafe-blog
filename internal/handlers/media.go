package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/relchzen/cafe-blog/internal/database"
	"github.com/relchzen/cafe-blog/internal/middleware"
	"github.com/relchzen/cafe-blog/internal/utils"
)

func UploadCafePhotos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get params id
	idString := r.PathValue("id")
	if idString == "" {
		http.Error(w, "Missing parameter value", http.StatusBadRequest)
		return
	}

	// Convert string cafeID to int
	cafeID, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "Invalid cafe ID", http.StatusBadRequest)
		return
	}

	// get current user ID
	userID := r.Context().Value(middleware.UserIDKey).(int)

	// Get the owner of the cafe record
	var existingUserID int
	checkQuery := `SELECT user_id FROM cafes WHERE id = $1`
	err = database.DB.QueryRow(checkQuery, cafeID).Scan(&existingUserID)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Cafe not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Check if user is authorized
	if existingUserID != userID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Set maximum files capacity
	err = r.ParseMultipartForm(50 << 20) // 50MB
	if err != nil {
		http.Error(w, "File too large (Max. 50MB total)", http.StatusBadRequest)
		return
	}

	// Get the photo files
	files := r.MultipartForm.File["photos"]
	if len(files) == 0 {
		http.Error(w, "No files provided", http.StatusBadRequest)
		return
	}

	// Set maximum number of files to 10
	if len(files) > 10 {
		http.Error(w, "Maximum 10 photos allowed", http.StatusBadRequest)
		return
	}

	var maxOrder int
	orderQuery := `SELECT COALESCE(MAX(display_order), -1) FROM cafe_photos WHERE cafe_id = $1`
	err = database.DB.QueryRow(orderQuery, cafeID).Scan(&maxOrder)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	var uploadedURLs []string
	for i, fileHeader := range files {
		photoURL, err := utils.UploadImage(fileHeader)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		displayOrder := maxOrder + i + 1
		insertQuery := `INSERT INTO cafe_photos (cafe_id, photo_url, display_order) VALUES ($1, $2, $3)`
		_, err = database.DB.Exec(insertQuery, cafeID, photoURL, displayOrder)
		if err != nil {
			err := utils.DeleteImage(photoURL)
			if err != nil {
				http.Error(w, "Failed to undo upload", http.StatusInternalServerError)
				return
			}
			http.Error(w, "Failed to save to database", http.StatusInternalServerError)
			return
		}

		uploadedURLs = append(uploadedURLs, photoURL)
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Photo(s) uploaded successfully",
		"count":   len(uploadedURLs),
		"photos":  uploadedURLs,
	})
	if err != nil {
		return
	}
}

func DeleteCafePhotos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cafeID, err := strconv.Atoi(r.PathValue("cafeId"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	photoID, err := strconv.Atoi(r.PathValue("photoId"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(int)

	var existingUserID int
	var photoURL string
	checkQuery := `SELECT c.user_id, cp.photo_url 
			FROM cafes c 
		    JOIN cafe_photos cp ON cp.cafe_id = c.id
		    WHERE c.id = $1 AND cp.id = $2`
	err = database.DB.QueryRow(checkQuery, cafeID, photoID).Scan(&existingUserID, &photoURL)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Photo not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if existingUserID != userID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	deleteQuery := `DELETE FROM cafe_photos WHERE cafe_id = $1`
	_, err = database.DB.Exec(deleteQuery, photoID)
	if err != nil {
		http.Error(w, "failed to delete photo", http.StatusInternalServerError)
		return
	}

	err = utils.DeleteImage(photoURL)
	if err != nil {
		log.Printf("Warning: Failed to delete from S3: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Photo deleted successfully",
	})
}
