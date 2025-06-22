package services

import (
	"context"
	"fmt"
	"github.com/DeryabinSergey/waste-tips-backend/internal/domain/models"
	"github.com/DeryabinSergey/waste-tips-backend/internal/infrastructure/localization"
	"io"
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"

	"google.golang.org/genai"
)

// WasteSortingService handles waste sorting business logic
type WasteSortingService struct {
	aiClient         *genai.Client
	localization     *localization.Localizer
	recaptchaService RecaptchaService
}

// RecaptchaService interface for reCAPTCHA verification
type RecaptchaService interface {
	VerifyToken(ctx context.Context, token string) (bool, error)
}

// NewWasteSortingService creates a new waste sorting service
func NewWasteSortingService(aiClient *genai.Client, localizer *localization.Localizer, recaptchaService RecaptchaService) *WasteSortingService {
	return &WasteSortingService{
		aiClient:         aiClient,
		localization:     localizer,
		recaptchaService: recaptchaService,
	}
}

// ProcessWasteImage processes the waste sorting request
func (s *WasteSortingService) ProcessWasteImage(ctx context.Context, req *models.WasteSortingRequest) (*models.WasteSortingResponse, error) {
	// Validate postal code
	if !s.isValidGermanPostalCode(req.PostalCode) {
		return &models.WasteSortingResponse{
			Success: false,
			Error:   s.localization.GetErrorMessage(req.Language, "invalid_postal_code"),
		}, nil
	}

	// Validate image file
	if !s.isValidImageFile(req.ImageHeader) {
		return &models.WasteSortingResponse{
			Success: false,
			Error:   s.localization.GetErrorMessage(req.Language, "invalid_image"),
		}, nil
	}

	// Verify reCAPTCHA
	isValid, err := s.recaptchaService.VerifyToken(ctx, req.RecaptchaCode)
	if err != nil || !isValid {
		return &models.WasteSortingResponse{
			Success: false,
			Error:   s.localization.GetErrorMessage(req.Language, "recaptcha_failed"),
		}, nil
	}

	// Process image with Gemini AI
	htmlResult, err := s.processImageWithGemini(ctx, req.ImageFile, req.PostalCode, req.Language)
	if err != nil {
		return &models.WasteSortingResponse{
			Success: false,
			Error:   s.localization.GetErrorMessage(req.Language, "processing_error"),
		}, nil
	}

	return &models.WasteSortingResponse{
		Success: true,
		HTML:    htmlResult,
	}, nil
}

// isValidGermanPostalCode validates German postal codes (5 digits, 01001-99998)
func (s *WasteSortingService) isValidGermanPostalCode(postalCode string) bool {
	// German postal codes are 5 digits, range 01001-99998
	matched, _ := regexp.MatchString(`^[0-9]{5}$`, postalCode)
	if !matched {
		return false
	}

	// Convert to int for range check
	code := 0
	if _, err := fmt.Sscanf(postalCode, "%d", &code); err != nil {
		return false
	}
	return code >= 1001 && code <= 99998
}

// isValidImageFile checks if the uploaded file is a valid image
func (s *WasteSortingService) isValidImageFile(fileHeader *multipart.FileHeader) bool {
	if fileHeader == nil {
		return false
	}

	contentType := fileHeader.Header.Get("Content-Type")
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
	}

	for _, validType := range validTypes {
		if contentType == validType {
			return true
		}
	}
	return false
}

// processImageWithGemini processes the image using Gemini AI
func (s *WasteSortingService) processImageWithGemini(ctx context.Context, file multipart.File, postalCode, language string) (string, error) {
	// Read image data
	imageData, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read image: %v", err)
	}
	mimeType := http.DetectContentType(imageData)

	// Create prompt based on language
	prompt := fmt.Sprintf(`Analyze this waste/garbage image and provide waste sorting instructions on Language %s for Germany, postal code %s.
	Identify what type of waste this is and explain which bin it should go into (Restmüll, Gelbe Tonne/Gelber Sack, Papiertonne, Biotonne, Glass container, etc.).
	Provide detailed instructions on how to sort this waste item, including any specific preparation steps (e.g., rinsing, removing labels).
	Include specific local regulations for postal code %s if relevant.
	Provide your response ONLY as valid HTML without any additional text, markdown, or explanations.
	Use proper HTML structure with headings, paragraphs, and lists where appropriate.`, language, postalCode, postalCode)
	contents := []*genai.Content{{
		Parts: []*genai.Part{
			{Text: prompt},
			{InlineData: &genai.Blob{
				Data:     imageData,
				MIMEType: mimeType,
			}},
		},
		Role: genai.RoleUser,
	}}

	resp, err := s.aiClient.Models.GenerateContent(ctx, "gemini-2.0-flash", contents, &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{{Text: "You are an expert in waste management and recycling regulations in Germany. You analyze waste items and provide detailed sorting instructions in the specified language. Your responses must be in valid HTML format only, without any additional text or markdown."}},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %v", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no response from Gemini")
	}

	for _, candidate := range resp.Candidates {
		if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
			continue
		}

		var textResponse string
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				textResponse += part.Text
			}
		}

		if textResponse == "" {
			continue
		}

		re := regexp.MustCompile(`(?s)<body[^>]*>(.*?)</body>`)
		match := re.FindStringSubmatch(textResponse)
		if len(match) > 1 {
			textResponse = match[1]
		} else {
			textResponse = ""
		}

		textResponse = strings.TrimSpace(textResponse)
		if textResponse == "" {
			continue
		}

		return textResponse, nil
	}

	return "", fmt.Errorf("no valid text response from Gemini")
}
