package repository

import "errors"

var (
	ErrNotFound = errors.New("the resource you are searching for is not available at the momemnt")
)
