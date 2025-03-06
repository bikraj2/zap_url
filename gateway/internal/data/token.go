package data

import (
	"database/sql"
	"time"
)

type Token struct {
	Plaintext string    `json:"plaintext"`
	UserId    string    `json:"-"`
	Type      string    `json:"type"`
	Hash      []byte    `json:"-"`
	Expiry    time.Time `json:"expiry"`
}
type TokenModel struct {
	DB *sql.DB
}

// GenerateToken
// SearchForUser
// DeleteForUser
