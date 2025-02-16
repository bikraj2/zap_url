package repository

import "errors"

var (
	ErrNoKeyleft = errors.New("store doesnot have a key at the moment")
)
