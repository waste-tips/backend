package models

import "mime/multipart"

// WasteSortingRequest represents the incoming request structure
type WasteSortingRequest struct {
	PostalCode    string                `json:"postal_code"`
	RecaptchaCode string                `json:"recaptcha_code"`
	Language      string                `json:"language"`
	ImageFile     multipart.File        `json:"-"`
	ImageHeader   *multipart.FileHeader `json:"-"`
}

// WasteSortingResponse represents the API response structure
type WasteSortingResponse struct {
	Success bool   `json:"success"`
	HTML    string `json:"html,omitempty"`
	Error   string `json:"error,omitempty"`
}