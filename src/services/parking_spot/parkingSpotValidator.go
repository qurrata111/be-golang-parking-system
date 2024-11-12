package parkingspot

import (
	"be-golang-parking-system/models"
)

func validateBookParkingSpotRequest(req models.TrParking) map[string]string {
	errors := make(map[string]string)

	if req.PlateNumber == nil || len(*req.PlateNumber) == 0 {
		errors["plate_number"] = "Plate Number is required"
	}
	if req.CustomerName == nil || len(*req.CustomerName) == 0 {
		errors["customer_name"] = "Customer Name is required"
	}

	return errors
}
