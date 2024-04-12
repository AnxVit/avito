package storage

import "errors"

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrUserExists     = errors.New("user already exists")
	ErrNotAccess      = errors.New("user don't have access")
	ErrBannerNotFound = errors.New("banner not found")
)
