package localization

// ErrorMessages contains localized error messages
type ErrorMessages struct {
	InvalidPostalCode string `json:"invalid_postal_code"`
	InvalidImage      string `json:"invalid_image"`
	RecaptchaFailed   string `json:"recaptcha_failed"`
	ProcessingError   string `json:"processing_error"`
	MissingFields     string `json:"missing_fields"`
}

// Localizer handles localization of messages
type Localizer struct {
	supportedLanguages map[string]bool
	errorMessages      map[string]ErrorMessages
}

// NewLocalizer creates a new localizer instance
func NewLocalizer() *Localizer {
	supportedLanguages := map[string]bool{
		"de": true, "en": true, "tr": true, "ru": true, "pl": true,
		"ar": true, "ku": true, "it": true, "bs": true, "hr": true,
		"sr": true, "ro": true, "el": true, "es": true, "fr": true,
		"hi": true, "ur": true, "vi": true, "zh": true, "fa": true,
		"ps": true, "ta": true, "sq": true, "da": true, "uk": true,
	}

	errorMessages := map[string]ErrorMessages{
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
		"ar": {
			InvalidPostalCode: "رمز بريدي ألماني غير صالح",
			InvalidImage:      "ملف صورة غير صالح",
			RecaptchaFailed:   "فشل التحقق من reCAPTCHA",
			ProcessingError:   "خطأ في معالجة طلبك",
			MissingFields:     "حقول مطلوبة مفقودة",
		},
		"ku": {
			InvalidPostalCode: "Koda postê ya Almanî ya nederust",
			InvalidImage:      "Pelê wêneyê nederust",
			RecaptchaFailed:   "Piştrastkirina reCAPTCHA têk çû",
			ProcessingError:   "Di pêvajoya daxwaza te de çewtî",
			MissingFields:     "Zeviyên pêwîst kêm in",
		},
		"it": {
			InvalidPostalCode: "Codice postale tedesco non valido",
			InvalidImage:      "File immagine non valido",
			RecaptchaFailed:   "Verifica reCAPTCHA fallita",
			ProcessingError:   "Errore nell'elaborazione della richiesta",
			MissingFields:     "Campi obbligatori mancanti",
		},
		"bs": {
			InvalidPostalCode: "Neispravan njemački poštanski broj",
			InvalidImage:      "Neispravna datoteka slike",
			RecaptchaFailed:   "reCAPTCHA provjera neuspješna",
			ProcessingError:   "Greška pri obradi zahtjeva",
			MissingFields:     "Nedostaju obavezna polja",
		},
		"hr": {
			InvalidPostalCode: "Neispravan njemački poštanski broj",
			InvalidImage:      "Neispravna datoteka slike",
			RecaptchaFailed:   "reCAPTCHA provjera neuspješna",
			ProcessingError:   "Greška pri obradi zahtjeva",
			MissingFields:     "Nedostaju obavezna polja",
		},
		"sr": {
			InvalidPostalCode: "Неисправан немачки поштански број",
			InvalidImage:      "Неисправна датотека слике",
			RecaptchaFailed:   "reCAPTCHA провера неуспешна",
			ProcessingError:   "Грешка при обради захтева",
			MissingFields:     "Недостају обавезна поља",
		},
		"ro": {
			InvalidPostalCode: "Cod poștal german invalid",
			InvalidImage:      "Fișier imagine invalid",
			RecaptchaFailed:   "Verificarea reCAPTCHA a eșuat",
			ProcessingError:   "Eroare la procesarea cererii",
			MissingFields:     "Câmpuri obligatorii lipsă",
		},
		"el": {
			InvalidPostalCode: "Μη έγκυρος γερμανικός ταχυδρομικός κώδικας",
			InvalidImage:      "Μη έγκυρο αρχείο εικόνας",
			RecaptchaFailed:   "Η επαλήθευση reCAPTCHA απέτυχε",
			ProcessingError:   "Σφάλμα επεξεργασίας του αιτήματός σας",
			MissingFields:     "Λείπουν υποχρεωτικά πεδία",
		},
		"es": {
			InvalidPostalCode: "Código postal alemán inválido",
			InvalidImage:      "Archivo de imagen inválido",
			RecaptchaFailed:   "Verificación reCAPTCHA fallida",
			ProcessingError:   "Error procesando su solicitud",
			MissingFields:     "Faltan campos requeridos",
		},
		"fr": {
			InvalidPostalCode: "Code postal allemand invalide",
			InvalidImage:      "Fichier image invalide",
			RecaptchaFailed:   "Échec de la vérification reCAPTCHA",
			ProcessingError:   "Erreur lors du traitement de votre demande",
			MissingFields:     "Champs requis manquants",
		},
		"hi": {
			InvalidPostalCode: "अमान्य जर्मन पोस्टल कोड",
			InvalidImage:      "अमान्य छवि फ़ाइल",
			RecaptchaFailed:   "reCAPTCHA सत्यापन विफल",
			ProcessingError:   "आपके अनुरोध को संसाधित करने में त्रुटि",
			MissingFields:     "आवश्यक फ़ील्ड गुम हैं",
		},
		"ur": {
			InvalidPostalCode: "غلط جرمن پوسٹل کوڈ",
			InvalidImage:      "غلط تصویری فائل",
			RecaptchaFailed:   "reCAPTCHA تصدیق ناکام",
			ProcessingError:   "آپ کی درخواست پر عمل کرنے میں خرابی",
			MissingFields:     "ضروری فیلڈز غائب ہیں",
		},
		"vi": {
			InvalidPostalCode: "Mã bưu điện Đức không hợp lệ",
			InvalidImage:      "Tệp hình ảnh không hợp lệ",
			RecaptchaFailed:   "Xác minh reCAPTCHA thất bại",
			ProcessingError:   "Lỗi xử lý yêu cầu của bạn",
			MissingFields:     "Thiếu các trường bắt buộc",
		},
		"zh": {
			InvalidPostalCode: "无效的德国邮政编码",
			InvalidImage:      "无效的图像文件",
			RecaptchaFailed:   "reCAPTCHA验证失败",
			ProcessingError:   "处理您的请求时出错",
			MissingFields:     "缺少必填字段",
		},
		"fa": {
			InvalidPostalCode: "کد پستی آلمان نامعتبر",
			InvalidImage:      "فایل تصویر نامعتبر",
			RecaptchaFailed:   "تأیید reCAPTCHA ناموفق",
			ProcessingError:   "خطا در پردازش درخواست شما",
			MissingFields:     "فیلدهای ضروری موجود نیست",
		},
		"ps": {
			InvalidPostalCode: "د آلمان د پوستې غلط کوډ",
			InvalidImage:      "د انځور غلط دوتنه",
			RecaptchaFailed:   "د reCAPTCHA تصدیق ناکام",
			ProcessingError:   "ستاسو د غوښتنې پروسس کولو کې تېروتنه",
			MissingFields:     "اړین ساحې ورک دي",
		},
		"ta": {
			InvalidPostalCode: "தவறான ஜெர்மன் அஞ்சல் குறியீடு",
			InvalidImage:      "தவறான படக் கோப்பு",
			RecaptchaFailed:   "reCAPTCHA சரிபார்ப்பு தோல்வி",
			ProcessingError:   "உங்கள் கோரிக்கையை செயலாக்குவதில் பிழை",
			MissingFields:     "தேவையான புலங்கள் காணவில்லை",
		},
		"sq": {
			InvalidPostalCode: "Kod postar gjerman i pavlefshëm",
			InvalidImage:      "Skedar imazhi i pavlefshëm",
			RecaptchaFailed:   "Verifikimi reCAPTCHA dështoi",
			ProcessingError:   "Gabim në përpunimin e kërkesës suaj",
			MissingFields:     "Mungojnë fushat e detyrueshme",
		},
		"da": {
			InvalidPostalCode: "Ugyldig tysk postnummer",
			InvalidImage:      "Ugyldig billedfil",
			RecaptchaFailed:   "reCAPTCHA-verifikation mislykkedes",
			ProcessingError:   "Fejl ved behandling af din anmodning",
			MissingFields:     "Manglende påkrævede felter",
		},
		"uk": {
			InvalidPostalCode: "Недійсний німецький поштовий індекс",
			InvalidImage:      "Недійсний файл зображення",
			RecaptchaFailed:   "Перевірка reCAPTCHA не вдалася",
			ProcessingError:   "Помилка обробки вашого запиту",
			MissingFields:     "Відсутні обов'язкові поля",
		},
	}

	return &Localizer{
		supportedLanguages: supportedLanguages,
		errorMessages:      errorMessages,
	}
}

// IsLanguageSupported checks if the language is supported
func (l *Localizer) IsLanguageSupported(language string) bool {
	return l.supportedLanguages[language]
}

// GetErrorMessage returns localized error message
func (l *Localizer) GetErrorMessage(language, messageType string) string {
	if messages, exists := l.errorMessages[language]; exists {
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
	if messages, exists := l.errorMessages["en"]; exists {
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