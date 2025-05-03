package server

import "errors"

var (
	ErrNotExists = errors.New("not exists")
	ErrInternal  = errors.New("internal server error")
)
