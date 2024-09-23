package model

import "time"

type ConnectionRequest struct {
	Id             int       `json:"id"`
	RecipientEmail string    `json"email"`
	FromUserId     int       `json:"from_user_id"`
	ToUserId       int       `json:"to_user_id"`
	Status         string    `json:"status"` //enum kinda-- pending, accepted, rejected
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// connection already accaepted --like linkedIn networks
type Connection struct {
	Id               int       `json:"id"`
	UserId           int       `json:"user_id"`
	ConnectionUserId int       `json:"connection_user_id"`
	ConnectedAt      time.Time `json:"connected_at"`
}

type ConnectionRequestEmail struct {
	Email string `json:"email"`
}
