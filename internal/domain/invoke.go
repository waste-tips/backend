package domain

import (
	"backend/internal/infrastructure/container"
	"context"
	"fmt"
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
	case r.Method == http.MethodPost && strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data"):
		// Handle waste sorting request
		appContainer.Logger.Info(spanCtx, map[string]interface{}{
			"message": "Processing waste sorting request",
			"method":  r.Method,
			"path":    r.URL.Path,
		})

		response, err := appContainer.WasteSortingHandler.HandleRequest(spanCtx, r)
		if err != nil {
			appContainer.Logger.Error(spanCtx, map[string]interface{}{
				"message": "Failed to process waste sorting request",
				"error":   err.Error(),
			})
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Determine status code based on response
		statusCode := http.StatusOK
		if !response.Success {
			statusCode = http.StatusBadRequest
		}

		appContainer.WasteSortingHandler.WriteJSONResponse(w, response, statusCode)

		appContainer.Logger.Info(spanCtx, map[string]interface{}{
			"message":     "Waste sorting request processed successfully",
			"success":     response.Success,
			"status_code": statusCode,
		})

	default:
		appContainer.Logger.Warning(spanCtx, map[string]interface{}{
			"message":      "Unsupported request",
			"method":       r.Method,
			"content_type": r.Header.Get("Content-Type"),
		})
		http.Error(w, "Not found", http.StatusNotFound)
	}
}