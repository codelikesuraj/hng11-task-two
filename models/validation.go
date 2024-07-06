package models

type InputError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
