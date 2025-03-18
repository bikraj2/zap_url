package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/bikraj2/url_shortener/gateway/internal"
	"github.com/lib/pq"
)

type User struct {
	FullName string `form:"full_name" binding:"required"`
	Email    string `form:"email" binding:"required,email"`
	Username string `form:"username" binding:"required"`
	Id       string `form:"id"`
	Password
}
type UserRepository struct {
	DB *sql.DB
}
type Password struct {
	PlainTextPassword string `form:"plain_text_password"`
	HashedPassword    string `form:"-"`
}

func New(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

// Register
func (u *UserRepository) RegisterUser(user *User) error {

	firstQuery := `
  SELECT id 
  FROM USERS
  WHERE email = $1 
`

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var userID int
	err := u.DB.QueryRowContext(ctx, firstQuery, user.Email).Scan(&userID)
	if err == nil {
		return customerror.NewDuplicateError("user with that email already exists", nil)
	} else if err != sql.ErrNoRows {
		return customerror.NewInternalServerError("eror while quering the database", nil)
	}

	query := `
  INSERT INTO USERS (full_name, email, username)
  VALUES ($1, $2, $3)
`

	_, err = u.DB.ExecContext(ctx, query, user.FullName, user.Email, user.Username)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			if pgErr.Code == "23505" { // Unique constraint violation
				return customerror.NewDuplicateError("user with that email already exists", nil)
			}
		}
		return customerror.NewInternalServerError("eror while quering the database", nil)
	}

	return nil
}

// Login
// GetByEmail
// GetByUsername
// Update
// Delete
// Verify User Mail
// func (u *User) Register
