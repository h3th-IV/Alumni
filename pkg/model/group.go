package model

import "time"

type GroupMessage struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}
