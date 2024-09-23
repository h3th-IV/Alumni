package model

import "time"

type Chat struct {
	Id          int       `json:"message_id"`
	SenderID    int       `json:"sender_id"`
	RecipientID int       `json:"recipient_id"`
	Email       string    `json:"email"` //recipient
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type GroupChat struct {
	Message string `json:"message"`
	GroupID int    `json:"group_id"`
}
