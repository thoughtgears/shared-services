package models

import "time"

type User struct {
	ID         string    `json:"id" firestore:"id"`
	FirstName  string    `json:"first_name" firestore:"first_name"`
	LastName   string    `json:"last_name" firestore:"last_name"`
	Email      string    `json:"email" firestore:"email"`
	Phone      string    `json:"phone" firestore:"phone"`
	Address    Address   `json:"address" firestore:"address"`
	FirebaseID string    `json:"firebase_id" firestore:"firebase_id"`
	CreatedAt  time.Time `json:"created_at" firestore:"created_at,serverTimestamp"`
	UpdatedAt  time.Time `json:"updated_at" firestore:"updated_at,serverTimestamp"`
}

type Address struct {
	BuildingNumber string `json:"building_number" firestore:"building_number"`
	Street         string `json:"street" firestore:"street"`
	City           string `json:"city" firestore:"city"`
	PostCode       string `json:"zip_code" firestore:"zip_code"`
	Country        string `json:"country" firestore:"country"`
}
