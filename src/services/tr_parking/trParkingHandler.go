package trparking

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

func GetDetailTrParkingHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	trParkingID, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid parking ID", http.StatusBadRequest)
		return
	}

	var trParking models.TrParking
	err = db.DB.QueryRow(`
			SELECT id, parking_spot_id, start_time, end_time, plate_number, customer_name, checkout_time
			FROM tr_parking
			WHERE id = $1 AND deleted_at IS NULL
		`, trParkingID).Scan(&trParking.ID, &trParking.ParkingSpotId, &trParking.StartTime, &trParking.EndTime, &trParking.PlateNumber, &trParking.CustomerName, &trParking.CheckoutTime)

	if err != nil {
		http.Error(w, "Tr parking not found", http.StatusNotFound)
		return
	}

	type TrParkingWithTransaction struct {
		models.TrParking
		ParkingSpot *models.ParkingSpot `json:"parking_spot,omitempty"`
	}

	response := struct {
		TrParking TrParkingWithTransaction `json:"tr_parking"`
	}{
		TrParking: TrParkingWithTransaction{
			TrParking: trParking,
		},
	}

	var parkingSpot models.ParkingSpot
	err = db.DB.QueryRow(`
			SELECT id, spot_number, is_occupied, location, vehicle_size
			FROM parking_spot
			WHERE id = $1
			LIMIT 1
		`, trParking.ParkingSpotId).Scan(&parkingSpot.ID, &parkingSpot.SpotNumber, &parkingSpot.IsOccupied, &parkingSpot.Location, &parkingSpot.VehicleSize)

	if err == nil {
		response.TrParking.ParkingSpot = &parkingSpot
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to encode tr parking details", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

func CheckoutTrParkingHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	trParkingID, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid parking ID", http.StatusBadRequest)
		return
	}

	currentTime := time.Now()

	var trParking models.TrParking
	err = db.DB.QueryRow(`
			SELECT id, parking_spot_id, start_time, end_time, plate_number, customer_name, checkout_time
			FROM tr_parking
			WHERE id = $1 AND deleted_at IS NULL AND checkout_time IS NULL
		`, trParkingID).Scan(&trParking.ID, &trParking.ParkingSpotId, &trParking.StartTime, &trParking.EndTime, &trParking.PlateNumber, &trParking.CustomerName, &trParking.CheckoutTime)

	if err != nil {
		http.Error(w, "Tr parking not found", http.StatusNotFound)
		return
	}

	res, err := db.DB.Exec(`
			UPDATE tr_parking 
			SET checkout_time = $1, updated_at = $2 
			WHERE id = $3 AND checkout_time IS NULL
	`, currentTime, currentTime, trParkingID)

	if err != nil {
		fmt.Println(err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil || rowsAffected == 0 {
		http.Error(w, "Tr parking not found or no changes made", http.StatusNotFound)
		return
	} else {
		res2, err2 := db.DB.Exec(`
				UPDATE parking_spot 
				SET is_occupied = false
				WHERE id = $1
		`, trParking.ParkingSpotId)

		if err2 != nil {
			fmt.Println(err2)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		rowsAffected2, err2 := res2.RowsAffected()
		if err2 != nil || rowsAffected2 == 0 {
			http.Error(w, "Parking spot not found or no changes made", http.StatusNotFound)
			return
		}
	}

	responseJSON, err := json.Marshal(rowsAffected)

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}
