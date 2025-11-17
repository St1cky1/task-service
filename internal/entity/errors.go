package entity

import "errors"

var (
	ErrForbidden        = errors.New("forbidden: access denied")
	ErrNoFieldsToUpdate = errors.New("no fields to update")
	ErrTaskNotFound     = errors.New("task not found")
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidTaskData  = errors.New("invalid task data")
	ErrInvalidUserData  = errors.New("invalid user data")
)
