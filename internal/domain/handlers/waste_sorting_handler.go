package handlers

import (
	"backend/internal/domain/models"
	"backend/internal/domain/services"
	"backend/internal/infrastructure/localization"
	"context"
	"encoding/json"
	"net/http"
)

// WasteSortingHandler handles HTTP requests for waste sorting
type WasteSortingHandler struct {
	service    *services.WasteSortingService
	localizer  *localization.Localizer
}

// NewWasteSortingHandler creates a new waste sorting handler
func NewWasteSortingHandler(service *services.WasteSortingService, localizer *localization.Localizer) *WasteSortingHandler {
	return &WasteSortingHandler{
		service:   service,
		localizer: localizer,
	}
}

// HandleRequest processes the waste sorting HTTP request
func (h *WasteSortingHandler) HandleRequest(ctx context.Context, r *http.Request) (*models.WasteSortingResponse, error) {
	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		return &models.WasteSortingResponse{
			Success: false,
			Error:   "Failed to parse form",
		}, nil
	}

	// Extract form fields
	postalCode := r.FormValue("postal_code")
	recaptchaCode := r.FormValue("recaptcha_code")
	language := r.FormValue("language")

	// Default to English if language not supported
	if !h.localizer.IsLanguageSupported(language) {
		language = "en"
	}

	// Validate required fields
	if postalCode == "" || recaptchaCode == "" {
		return &models.WasteSortingResponse{
			Success: false,
			Error:   h.localizer.GetErrorMessage(language, "missing_fields"),
		}, nil
	}

	// Get uploaded file
	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		return &models.WasteSortingResponse{
			Success: false,
			Error:   h.localizer.GetErrorMessage(language, "invalid_image"),
		}, nil
	}
	defer file.Close()

	// Create request model
	request := &models.WasteSortingRequest{
		PostalCode:    postalCode,
		RecaptchaCode: recaptchaCode,
		Language:      language,
		ImageFile:     file,
		ImageHeader:   fileHeader,
	}

	// Process the request
	return h.service.ProcessWasteImage(ctx, request)
}

// WriteJSONResponse writes a JSON response to the HTTP response writer
func (h *WasteSortingHandler) WriteJSONResponse(w http.ResponseWriter, response *models.WasteSortingResponse, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}