# Waste Sorting API for Germany

A Google Cloud Function API backend for a waste sorting application in Germany. This API processes images of waste items and provides localized sorting instructions using AI.

## Features

- **Multi-language support**: 25 languages including German, English, Turkish, Russian, Polish, Arabic, and more
- **Image processing**: Uses Google Gemini AI to analyze waste images
- **Postal code validation**: Validates German postal codes (01001-99998)
- **reCAPTCHA Enterprise integration**: Spam protection
- **CORS support**: Ready for SPA integration
- **Localized error messages**: Error responses in user's language

## API Endpoint

**POST** `/` (Cloud Function endpoint)

### Request Format

Multipart form data with the following fields:

- `postal_code` (string, required): German postal code (5 digits)
- `recaptcha_code` (string, required): reCAPTCHA Enterprise token
- `language` (string, required): Language code (de, en, tr, ru, pl, etc.)
- `image` (file, required): Image file (JPEG, PNG, GIF, WebP)

### Response Format

```json
{
  "success": true,
  "html": "<div><h2>Waste Sorting Instructions</h2>...</div>"
}
```

Or in case of error:

```json
{
  "success": false,
  "error": "Error message in requested language"
}
```

## Environment Variables

Set these environment variables in Google Cloud Functions:

- `GOOGLE_CLOUD_PROJECT`: Your Google Cloud project ID
- `RECAPTCHA_SITE_KEY`: Your reCAPTCHA Enterprise site key
- `GEMINI_API_KEY`: Your Google Gemini API key

## Supported Languages

The API supports 25 languages:
- German (de) - Primary language
- English (en) - Fallback language
- Turkish (tr), Russian (ru), Polish (pl)
- Arabic (ar), Kurdish (ku), Italian (it)
- Bosnian (bs), Croatian (hr), Serbian (sr)
- Romanian (ro), Greek (el), Spanish (es)
- French (fr), Hindi (hi), Urdu (ur)
- Vietnamese (vi), Chinese (zh), Persian (fa)
- Pashto (ps), Tamil (ta), Albanian (sq)
- Danish (da), Ukrainian (uk)

## Deployment

1. Deploy to Google Cloud Functions with Go 1.21 runtime
2. Set the entry point to `ProcessWasteImage`
3. Configure environment variables
4. Enable the following APIs:
   - Cloud Functions API
   - reCAPTCHA Enterprise API
   - Vertex AI API (for Gemini)

```bash
gcloud functions deploy backend-service \
  --entry-point=Invoke \
  --runtime go123 \
  --gen2 \
  --region=europe-west3 \
  --trigger-http \
  --set-secrets="RECAPTCHA_SITE_KEY=projects/519359753202/secrets/RECAPTCHA_SITE_KEY:latest" \
  --allow-unauthenticated
```

## Local Development

```bash
# Install Functions Framework
go mod tidy

# Run locally
functions-framework --target=ProcessWasteImage --port=8080
```

## Security Features

- reCAPTCHA Enterprise verification
- File type validation (images only)
- Postal code format validation
- Request size limits (10MB max)
- CORS protection

## Error Handling

All errors return localized messages in the requested language with appropriate HTTP status codes:

- 400: Bad Request (validation errors)
- 405: Method Not Allowed
- 500: Internal Server Error

## Integration with Frontend

Example JavaScript code for calling the API:

```javascript
const formData = new FormData();
formData.append('postal_code', '10115');
formData.append('recaptcha_code', recaptchaToken);
formData.append('language', 'de');
formData.append('image', imageFile);

const response = await fetch('https://your-function-url', {
  method: 'POST',
  body: formData
});

const result = await response.json();
if (result.success) {
  document.getElementById('result').innerHTML = result.html;
} else {
  console.error(result.error);
}
```