package domain

import (
	"backend/internal/infrastructure/container"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Invoke is the main entry point for Google Cloud Functions
func Invoke(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	ctx := r.Context()
	// Check if the request context is valid
	if ctx.Err() != nil {
		http.Error(w, "Request context is invalid", http.StatusInternalServerError)
		return
	}

	// Initialize container if not already done
	appContainer, err := container.NewContainer(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to initialize application: %v", err), http.StatusInternalServerError)
		return
	}
	if appContainer == nil {
		http.Error(w, fmt.Sprintf("Failed to initialize application: %v", err), http.StatusInternalServerError)
		return
	}
	defer func(ctx context.Context) {
		_ = appContainer.Tracer.Close(ctx)
		_ = appContainer.Logger.Close(ctx)
	}(ctx)

	spanCtx, span := appContainer.Tracer.Start(ctx, "Application Invoke")
	defer span.End()
	r = r.WithContext(spanCtx)

	switch {
	case r.Method == http.MethodPost && strings.Contains(r.Header.Get("Content-Type"), "application/json"):
		// Check if it's a Telegram webhook
		body, err := io.ReadAll(r.Body)
		if err != nil {
			appContainer.Logger.Error(spanCtx, map[string]interface{}{
				"message": "Failed to read request body",
				"error":   err.Error(),
			})
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		// Here should be reCaptcha validator Gemini request and response handling

	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}
