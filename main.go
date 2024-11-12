package main

import (
	"be-golang-parking-system/src/helper/db"
	parkingspot "be-golang-parking-system/src/services/parking_spot"
	trparking "be-golang-parking-system/src/services/tr_parking"
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func init() {
	db.Init()
	log.Println("Connected to postgresql")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	router := httprouter.New()

	router.GET("/parking-spot/paginate", parkingspot.GetAllParkingSpotPaginateHandler)
	router.GET("/parking-spot/detail/:id", parkingspot.GetDetailParkingSpotHandler)
	router.POST("/parking-spot/checkout/:id", parkingspot.CheckoutParkingSpotHandler)
	router.POST("/parking-spot/book/:id", parkingspot.BookingParkingSpotHandler)
	router.GET("/tr-parking/detail/:id", trparking.GetDetailTrParkingHandler)
	router.POST("/tr-parking/checkout/:id", trparking.CheckoutTrParkingHandler)

	fmt.Println("Server is running on port 8080...")
	if err := http.ListenAndServe(":8080", corsMiddleware(router)); err != nil {
		log.Fatal("Failed to start server", err)
	}
}
