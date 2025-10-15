package models

import "time"

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type Appointment struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Title       string    `json:"title"`
	ScheduledAt time.Time `json:"scheduled_at"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type Payment struct {
	ID                    int64     `json:"id"`
	UserID                int64     `json:"user_id"`
	AppointmentID         int64     `json:"appointment_id"`
	StripePaymentIntentID string    `json:"stripe_payment_intent_id"`
	Amount                int64     `json:"amount"`
	Currency              string    `json:"currency"`
	Status                string    `json:"status"`
	CreatedAt             time.Time `json:"created_at"`
}

type Event struct {
	ID        int64           `json:"id"`
	UserID    int64           `json:"user_id"`
	EventType string          `json:"event_type"`
	Metadata  map[string]any  `json:"metadata"`
	CreatedAt time.Time       `json:"created_at"`
}
