package wastesorting

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"strings"

	"cloud.google.com/go/recaptchaenterprise/v2/apiv1"
	"cloud.google.com/go/recaptchaenterprise/v2/apiv1/recaptchaenterprisepb"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Request represents the incoming request structure
type Request struct {
	PostalCode    string `json:"postal_code"`
	RecaptchaCode string `json:"recaptcha_code"`
	Language      string `json:"language"`
}

// Response represents the API response structure
type Response struct {
	Success bool   `json:"success"`
	HTML    string `json:"html,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ErrorMessages contains localized error messages
type ErrorMessages struct {
	InvalidPostalCode string `json:"invalid_postal_code"`
	InvalidImage      string `json:"invalid_image"`
	RecaptchaFailed   string `json:"recaptcha_failed"`
	ProcessingError   string `json:"processing_error"`
	MissingFields     string `json:"missing_fields"`
}

var (
	supportedLanguages = map[string]bool{
		"de": true, "en": true, "tr": true, "ru": true, "pl": true,
		"ar": true, "ku": true, "it": true, "bs": true, "hr": true,
		"sr": true, "ro": true, "el": true, "es": true, "fr": true,
		"hi": true, "ur": true, "vi": true, "zh": true, "fa": true,
		"ps": true, "ta": true, "sq": true, "da": true, "uk": true,
	}

	errorMessages = map[string]ErrorMessages{
		"en": {
			InvalidPostalCode: "Invalid German postal code",
			InvalidImage:      "Invalid image file",
			RecaptchaFailed:   "reCAPTCHA verification failed",
			ProcessingError:   "Error processing your request",
			MissingFields:     "Missing required fields",
		},
		"de": {
			InvalidPostalCode: "Ungültige deutsche Postleitzahl",
			InvalidImage:      "Ungültige Bilddatei",
			RecaptchaFailed:   "reCAPTCHA-Verifizierung fehlgeschlagen",
			ProcessingError:   "Fehler bei der Verarbeitung Ihrer Anfrage",
			MissingFields:     "Pflichtfelder fehlen",
		},
		"ru": {
			InvalidPostalCode: "Неверный немецкий почтовый индекс",
			InvalidImage:      "Неверный файл изображения",
			RecaptchaFailed:   "Проверка reCAPTCHA не удалась",
			ProcessingError:   "Ошибка обработки вашего запроса",
			MissingFields:     "Отсутствуют обязательные поля",
		},
		"tr": {
			InvalidPostalCode: "Geçersiz Alman posta kodu",
			InvalidImage:      "Geçersiz resim dosyası",
			RecaptchaFailed:   "reCAPTCHA doğrulaması başarısız",
			ProcessingError:   "İsteğinizi işleme hatası",
			MissingFields:     "Gerekli alanlar eksik",
		},
		"pl": {
			InvalidPostalCode: "Nieprawidłowy niemiecki kod pocztowy",
			InvalidImage:      "Nieprawidłowy plik obrazu",
			RecaptchaFailed:   "Weryfikacja reCAPTCHA nie powiodła się",
			ProcessingError:   "Błąd przetwarzania Twojego żądania",
			MissingFields:     "Brakuje wymaganych pól",
		},
	}
)

func init() {
	functions.HTTP("ProcessWasteImage", ProcessWasteImage)
}

// ProcessWasteImage is the main Cloud Function handler
func ProcessWasteImage(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only allow POST requests
	if r.Method != "POST" {
		writeErrorResponse(w, "Method not allowed", "en", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		writeErrorResponse(w, "Failed to parse form", "en", http.StatusBadRequest)
		return
	}

	// Extract form fields
	postalCode := r.FormValue("postal_code")
	recaptchaCode := r.FormValue("recaptcha_code")
	language := r.FormValue("language")

	// Default to English if language not supported
	if !supportedLanguages[language] {
		language = "en"
	}

	// Validate required fields
	if postalCode == "" || recaptchaCode == "" {
		writeErrorResponse(w, getErrorMessage(language, "missing_fields"), language, http.StatusBadRequest)
		return
	}

	// Validate German postal code
	if !isValidGermanPostalCode(postalCode) {
		writeErrorResponse(w, getErrorMessage(language, "invalid_postal_code"), language, http.StatusBadRequest)
		return
	}

	// Get uploaded file
	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		writeErrorResponse(w, getErrorMessage(language, "invalid_image"), language, http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate image file
	if !isValidImageFile(fileHeader) {
		writeErrorResponse(w, getErrorMessage(language, "invalid_image"), language, http.StatusBadRequest)
		return
	}

	// Verify reCAPTCHA
	if !verifyRecaptcha(r.Context(), recaptchaCode) {
		writeErrorResponse(w, getErrorMessage(language, "recaptcha_failed"), language, http.StatusBadRequest)
		return
	}

	// Process image with Gemini AI
	htmlResult, err := processImageWithGemini(r.Context(), file, postalCode, language)
	if err != nil {
		log.Printf("Error processing image with Gemini: %v", err)
		writeErrorResponse(w, getErrorMessage(language, "processing_error"), language, http.StatusInternalServerError)
		return
	}

	// Return success response
	response := Response{
		Success: true,
		HTML:    htmlResult,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// isValidGermanPostalCode validates German postal codes (5 digits, 01001-99998)
func isValidGermanPostalCode(postalCode string) bool {
	// German postal codes are 5 digits, range 01001-99998
	matched, _ := regexp.MatchString(`^[0-9]{5}$`, postalCode)
	if !matched {
		return false
	}
	
	// Convert to int for range check
	code := 0
	fmt.Sscanf(postalCode, "%d", &code)
	return code >= 1001 && code <= 99998
}

// isValidImageFile checks if the uploaded file is a valid image
func isValidImageFile(fileHeader *multipart.FileHeader) bool {
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

// verifyRecaptcha verifies the reCAPTCHA Enterprise token
func verifyRecaptcha(ctx context.Context, token string) bool {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	siteKey := os.Getenv("RECAPTCHA_SITE_KEY")
	
	if projectID == "" || siteKey == "" {
		log.Printf("Missing reCAPTCHA configuration")
		return false
	}

	client, err := recaptchaenterprise.NewClient(ctx)
	if err != nil {
		log.Printf("Error creating reCAPTCHA client: %v", err)
		return false
	}
	defer client.Close()

	request := &recaptchaenterprisepb.CreateAssessmentRequest{
		Parent: fmt.Sprintf("projects/%s", projectID),
		Assessment: &recaptchaenterprisepb.Assessment{
			Event: &recaptchaenterprisepb.Event{
				Token:   token,
				SiteKey: siteKey,
			},
		},
	}

	response, err := client.CreateAssessment(ctx, request)
	if err != nil {
		log.Printf("Error creating reCAPTCHA assessment: %v", err)
		return false
	}

	// Check if token is valid and score is acceptable
	return response.TokenProperties.Valid && response.RiskAnalysis.Score >= 0.5
}

// processImageWithGemini processes the image using Gemini AI
func processImageWithGemini(ctx context.Context, file multipart.File, postalCode, language string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("missing Gemini API key")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", fmt.Errorf("failed to create Gemini client: %v", err)
	}
	defer client.Close()

	// Read image data
	imageData, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read image: %v", err)
	}

	model := client.GenerativeModel("gemini-1.5-flash")
	
	// Create prompt based on language
	prompt := createPrompt(language, postalCode)

	// Create image part
	imagePart := genai.ImageData("image/jpeg", imageData)
	
	// Generate content
	resp, err := model.GenerateContent(ctx, genai.Text(prompt), imagePart)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %v", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no response from Gemini")
	}

	// Extract text from response
	var result strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			result.WriteString(string(text))
		}
	}

	return result.String(), nil
}

// createPrompt creates a localized prompt for Gemini AI
func createPrompt(language, postalCode string) string {
	prompts := map[string]string{
		"en": fmt.Sprintf(`Analyze this waste/garbage image and provide waste sorting instructions for Germany, postal code %s. 
Identify what type of waste this is and explain which bin it should go into (Restmüll, Gelbe Tonne/Gelber Sack, Papiertonne, Biotonne, Glass container, etc.).
Include specific local regulations for postal code %s if relevant.
Provide your response ONLY as valid HTML without any additional text, markdown, or explanations. 
Use proper HTML structure with headings, paragraphs, and lists where appropriate.`, postalCode, postalCode),

		"de": fmt.Sprintf(`Analysiere dieses Müll-/Abfallbild und gib Anweisungen zur Mülltrennung für Deutschland, Postleitzahl %s.
Identifiziere, um welche Art von Abfall es sich handelt und erkläre, in welche Tonne er gehört (Restmüll, Gelbe Tonne/Gelber Sack, Papiertonne, Biotonne, Glascontainer, etc.).
Berücksichtige spezifische lokale Vorschriften für die Postleitzahl %s, falls relevant.
Gib deine Antwort NUR als gültiges HTML ohne zusätzlichen Text, Markdown oder Erklärungen.
Verwende eine ordnungsgemäße HTML-Struktur mit Überschriften, Absätzen und Listen, wo angemessen.`, postalCode, postalCode),

		"ru": fmt.Sprintf(`Проанализируй это изображение мусора/отходов и предоставь инструкции по сортировке отходов для Германии, почтовый индекс %s.
Определи, какой это тип отходов и объясни, в какой контейнер его следует поместить (Restmüll, Gelbe Tonne/Gelber Sack, Papiertonne, Biotonne, стеклянный контейнер и т.д.).
Включи специфические местные правила для почтового индекса %s, если это актуально.
Предоставь свой ответ ТОЛЬКО в виде валидного HTML без дополнительного текста, markdown или объяснений.
Используй правильную HTML структуру с заголовками, параграфами и списками где необходимо.`, postalCode, postalCode),

		"tr": fmt.Sprintf(`Bu atık/çöp görüntüsünü analiz et ve Almanya, posta kodu %s için atık ayırma talimatları ver.
Bu atığın ne tür olduğunu belirle ve hangi çöp kutusuna gitmesi gerektiğini açıkla (Restmüll, Gelbe Tonne/Gelber Sack, Papiertonne, Biotonne, Cam konteyneri, vb.).
Posta kodu %s için özel yerel düzenlemeler varsa dahil et.
Yanıtını SADECE ek metin, markdown veya açıklama olmadan geçerli HTML olarak ver.
Uygun olan yerlerde başlıklar, paragraflar ve listeler ile düzgün HTML yapısı kullan.`, postalCode, postalCode),

		"pl": fmt.Sprintf(`Przeanalizuj ten obraz odpadów/śmieci i podaj instrukcje sortowania odpadów dla Niemiec, kod pocztowy %s.
Zidentyfikuj, jaki to rodzaj odpadu i wyjaśnij, do którego pojemnika powinien trafić (Restmüll, Gelbe Tonne/Gelber Sack, Papiertonne, Biotonne, pojemnik na szkło, itp.).
Uwzględnij specyficzne lokalne przepisy dla kodu pocztowego %s, jeśli są istotne.
Podaj swoją odpowiedź TYLKO jako prawidłowy HTML bez dodatkowego tekstu, markdown lub wyjaśnień.
Użyj odpowiedniej struktury HTML z nagłówkami, akapitami i listami tam, gdzie to właściwe.`, postalCode, postalCode),
	}

	if prompt, exists := prompts[language]; exists {
		return prompt
	}
	return prompts["en"] // fallback to English
}

// getErrorMessage returns localized error message
func getErrorMessage(language, messageType string) string {
	if messages, exists := errorMessages[language]; exists {
		switch messageType {
		case "invalid_postal_code":
			return messages.InvalidPostalCode
		case "invalid_image":
			return messages.InvalidImage
		case "recaptcha_failed":
			return messages.RecaptchaFailed
		case "processing_error":
			return messages.ProcessingError
		case "missing_fields":
			return messages.MissingFields
		}
	}
	
	// Fallback to English
	if messages, exists := errorMessages["en"]; exists {
		switch messageType {
		case "invalid_postal_code":
			return messages.InvalidPostalCode
		case "invalid_image":
			return messages.InvalidImage
		case "recaptcha_failed":
			return messages.RecaptchaFailed
		case "processing_error":
			return messages.ProcessingError
		case "missing_fields":
			return messages.MissingFields
		}
	}
	
	return "An error occurred"
}

// writeErrorResponse writes an error response
func writeErrorResponse(w http.ResponseWriter, message, language string, statusCode int) {
	response := Response{
		Success: false,
		Error:   message,
	}
	
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}