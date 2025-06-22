package services

import (
	"backend/internal/domain/models"
	"backend/internal/infrastructure/localization"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"regexp"
	"strings"

	"google.golang.org/genai"
)

// WasteSortingService handles waste sorting business logic
type WasteSortingService struct {
	aiClient      *genai.Client
	localization  *localization.Localizer
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
	fmt.Sscanf(postalCode, "%d", &code)
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

	model := s.aiClient.GenerativeModel("gemini-1.5-flash")
	
	// Create prompt based on language
	prompt := s.createPrompt(language, postalCode)

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
func (s *WasteSortingService) createPrompt(language, postalCode string) string {
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

		"ar": fmt.Sprintf(`حلل صورة النفايات/القمامة هذه وقدم تعليمات فرز النفايات لألمانيا، الرمز البريدي %s.
حدد نوع النفايات هذا واشرح في أي حاوية يجب وضعها (Restmüll، Gelbe Tonne/Gelber Sack، Papiertonne، Biotonne، حاوية الزجاج، إلخ).
اشمل اللوائح المحلية المحددة للرمز البريدي %s إذا كانت ذات صلة.
قدم إجابتك فقط كـ HTML صالح بدون أي نص إضافي أو markdown أو شروحات.
استخدم هيكل HTML مناسب مع العناوين والفقرات والقوائم حسب الاقتضاء.`, postalCode, postalCode),
	}

	if prompt, exists := prompts[language]; exists {
		return prompt
	}
	return prompts["en"] // fallback to English
}