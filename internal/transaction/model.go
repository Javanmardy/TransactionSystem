package transaction

import "time"

type Transaction struct {
	ID        int       `json:"id,omitempty"`
	UserID    int       `json:"user_id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
