package models

type ApiError struct {
	Error string
}

func NewApiError(msg string) ApiError {
	return ApiError{msg}
}
