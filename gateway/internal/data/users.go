package data

import "database/sql"

type User struct {
	FullName string
	Email    string
	Username string
	Id       string
	Password string
}
type UserModel struct {
	DB *sql.DB
}

// Register
// Login
// GetByEmail
// GetByUsername
// Update
// Delete
// Verify User Mail
// func (u *User) Register
