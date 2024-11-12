package models

import (
	"time"

	_ "github.com/lib/pq"
)

type ParkingSpot struct {
	ID          int        `gorm:"primaryKey;autoIncrement;column:id"`
	SpotNumber  *string    `gorm:"type:varchar;column:spot_number"`
	IsOccupied  *bool      `gorm:"type:boolean;column:is_occupied"`
	Location    *string    `gorm:"type:varchar;column:location"`
	VehicleSize *string    `gorm:"type:varchar;column:vehicle_size"`
	CreatedAt   *time.Time `gorm:"column:created_at"`
	CreatedBy   *string    `gorm:"type:varchar;column:created_by"`
	UpdatedAt   *time.Time `gorm:"column:updated_at"`
	UpdatedBy   *string    `gorm:"type:varchar;column:updated_by"`
	DeletedAt   *time.Time `gorm:"column:deleted_at"`
}

type TrParking struct {
	ID            int        `gorm:"primaryKey;autoIncrement;column:id"`
	ParkingSpotId int        `gorm:"type:integer;column:parking_spot_id"`
	StartTime     *time.Time `gorm:"column:start_time"`
	EndTime       *time.Time `gorm:"column:end_time"`
	PlateNumber   *string    `gorm:"type:varchar;column:plate_number"`
	CustomerName  *string    `gorm:"type:varchar;column:customer_name"`
	CreatedAt     *time.Time `gorm:"column:created_at"`
	CreatedBy     *string    `gorm:"type:varchar;column:created_by"`
	UpdatedAt     *time.Time `gorm:"column:updated_at"`
	UpdatedBy     *string    `gorm:"type:varchar;column:updated_by"`
	DeletedAt     *time.Time `gorm:"column:deleted_at"`
	CheckoutTime  *time.Time `gorm:"column:checkout_time"`
	// ParkingSpot   ParkingSpot `gorm:"foreignKey:ParkingSpotId;references:ID"`
}
