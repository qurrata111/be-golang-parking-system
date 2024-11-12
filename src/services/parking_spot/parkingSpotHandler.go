package parkingspot

import (
	"be-golang-parking-system/models"
	"be-golang-parking-system/src/helper/db"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

func GetAllParkingSpotPaginateHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	search := r.URL.Query().Get("search")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	query := "SELECT id, spot_number, is_occupied, location, vehicle_size FROM parking_spot WHERE deleted_at IS NULL"
	args := []interface{}{}
	argID := 1

	// Adjust main query with search conditions
	if search != "" {
		query += fmt.Sprintf(" AND (spot_number ILIKE $%d OR location ILIKE $%d)", argID, argID+1)
		args = append(args, "%"+search+"%", "%"+search+"%")
		argID += 2
	}

	// Add pagination to main query
	query += fmt.Sprintf(" ORDER BY spot_number LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, limit, offset)

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to get parking spot", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Scan results into parkingSpots slice
	var parkingSpots []models.ParkingSpot
	for rows.Next() {
		var parkingSpot models.ParkingSpot
		if err := rows.Scan(&parkingSpot.ID, &parkingSpot.SpotNumber, &parkingSpot.IsOccupied, &parkingSpot.Location, &parkingSpot.VehicleSize); err != nil {
			http.Error(w, "Failed to scan parking spot", http.StatusInternalServerError)
			return
		}
		parkingSpots = append(parkingSpots, parkingSpot)
	}

	// Separate countQuery and arguments
	countQuery := "SELECT COUNT(*) FROM parking_spot WHERE deleted_at IS NULL"
	countArgs := []interface{}{}

	// Adjust count query with search conditions
	if search != "" {
		countQuery += " AND (spot_number ILIKE $1 OR location ILIKE $2)"
		countArgs = append(countArgs, "%"+search+"%", "%"+search+"%")
	}

	// Execute count query
	var totalParkingSpots int
	err = db.DB.QueryRow(countQuery, countArgs...).Scan(&totalParkingSpots)
	if err != nil {
		http.Error(w, "Failed to count parking spots", http.StatusInternalServerError)
		return
	}

	// Calculate total pages
	totalPages := (totalParkingSpots + limit - 1) / limit

	// Prepare the response
	response := map[string]interface{}{
		"parking_spots": parkingSpots,
		"pagination": map[string]interface{}{
			"current_page": page,
			"total_pages":  totalPages,
			"total_tasks":  totalParkingSpots,
		},
	}

	// Encode the response as JSON
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to encode parking spots", http.StatusInternalServerError)
		return
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

func GetDetailParkingSpotHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	parkingSpotID, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid parking spot ID", http.StatusBadRequest)
		return
	}

	var parkingSpot models.ParkingSpot
	err = db.DB.QueryRow(`
		SELECT id, spot_number, is_occupied, location, vehicle_size
		FROM parking_spot
		WHERE id = $1 AND deleted_at IS NULL
	`, parkingSpotID).Scan(&parkingSpot.ID, &parkingSpot.SpotNumber, &parkingSpot.IsOccupied, &parkingSpot.Location, &parkingSpot.VehicleSize)

	if err != nil {
		http.Error(w, "Parking spot not found", http.StatusNotFound)
		return
	}

	type ParkingSpotWithTransaction struct {
		models.ParkingSpot
		TrParking *models.TrParking `json:"tr_parking,omitempty"`
	}

	response := struct {
		ParkingSpot ParkingSpotWithTransaction `json:"parking_spot"`
	}{
		ParkingSpot: ParkingSpotWithTransaction{
			ParkingSpot: parkingSpot,
		},
	}

	if parkingSpot.IsOccupied != nil && *parkingSpot.IsOccupied {
		var trParking models.TrParking
		err = db.DB.QueryRow(`
			SELECT id, parking_spot_id, start_time, end_time, plate_number, customer_name, checkout_time
			FROM tr_parking
			WHERE parking_spot_id = $1 AND checkout_time IS NULL 
			ORDER BY start_time DESC
			LIMIT 1
		`, parkingSpotID).Scan(&trParking.ID, &trParking.ParkingSpotId, &trParking.StartTime, &trParking.EndTime, &trParking.PlateNumber, &trParking.CustomerName, &trParking.CheckoutTime)

		if err == nil {
			response.ParkingSpot.TrParking = &trParking
		}
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to encode parking spot details", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

func CheckoutParkingSpotHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	parkingSpotID, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid parking spot ID", http.StatusBadRequest)
		return
	}

	var parkingSpot models.ParkingSpot
	err = db.DB.QueryRow(`
			SELECT id, spot_number, is_occupied, location, vehicle_size
			FROM parking_spot
			WHERE id = $1 AND deleted_at IS NULL AND is_occupied = true
		`, parkingSpotID).Scan(&parkingSpot.ID, &parkingSpot.SpotNumber, &parkingSpot.IsOccupied, &parkingSpot.Location, &parkingSpot.VehicleSize)

	if err != nil {
		http.Error(w, "Occupied parking space not found", http.StatusNotFound)
		return
	}

	currentTime := time.Now()

	res, err := db.DB.Exec(`
			UPDATE parking_spot 
			SET is_occupied = false 
			WHERE id = $1
	`, parkingSpotID)

	if err != nil {
		fmt.Println(err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "Parking spot not found or no changes made", http.StatusNotFound)
		return
	} else {
		db.DB.Exec(`
				UPDATE tr_parking 
				SET checkout_time = $1, updated_at = $2 
				WHERE parking_spot_id = $3 AND checkout_time IS NULL
			`, currentTime, currentTime, parkingSpotID)
	}

	responseJSON, err := json.Marshal(rowsAffected)

	if err != nil {
		http.Error(w, "Error reading rows", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

func BookingParkingSpotHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	parkingSpotID, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid parking spot ID", http.StatusBadRequest)
		return
	}

	var req models.TrParking
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	errors := validateBookParkingSpotRequest(req)
	if len(errors) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": errors,
		})
		return
	}

	currentTime := time.Now()

	res, err := db.DB.Exec(`
			UPDATE parking_spot
			SET is_occupied = true, updated_at = $1
			WHERE id = $2 AND deleted_at IS NULL AND is_occupied = false
	`, currentTime, parkingSpotID)

	if err != nil {
		fmt.Println(err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()

	if err != nil || rowsAffected == 0 {
		http.Error(w, "Parking spot not found or no changes made", http.StatusNotFound)
		return
	} else {
		_, err := db.DB.Exec(`
				INSERT INTO tr_parking (parking_spot_id, start_time, plate_number, customer_name, created_at)
				VALUES ($1, $2, $3, $4, $5)
			`, parkingSpotID, currentTime, req.PlateNumber, req.CustomerName, currentTime)

		if err != nil {
			http.Error(w, "Failed to book customer a spot", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(req)
}
