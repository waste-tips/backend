package container

import (
	"backend/internal/domain/handlers"
	"backend/internal/domain/services"
	"backend/internal/infrastructure/config"
	"backend/internal/infrastructure/localization"
	"backend/internal/infrastructure/recaptcha"
	"backend/libs/logger"
	"backend/libs/tracer"
	"context"
	"fmt"
	"google.golang.org/genai"
)

// Container holds all application dependencies
type Container struct {
	Config             *config.Config
	Logger             *logger.Log
	Tracer             *tracer.Tracer
	Ai                 *genai.Client
	Localizer          *localization.Localizer
	RecaptchaService   *recaptcha.Service
	WasteSortingService *services.WasteSortingService
	WasteSortingHandler *handlers.WasteSortingHandler
}

// NewContainer creates and initializes the dependency injection container
func NewContainer(ctx context.Context) (*Container, error) {
	cfg := config.LoadConfig()

	// Initialize logger
	l, err := logger.Init(ctx, cfg.ProjectID, cfg.ApplicationName, cfg.GCPEnabled, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize tracer
	tr, err := tracer.Init(ctx, cfg.ProjectID, cfg.ApplicationName, cfg.GCPEnabled)
	if err != nil {
		l.Critical(ctx, map[string]interface{}{
			"message": "failed to initialize tracer",
			"error":   err.Error(),
		})
		return nil, fmt.Errorf("failed to initialize tracer: %w", err)
	}

	// Initialize Gemini client
	geminiClient, err := genai.NewClient(ctx, &genai.ClientConfig{
		HTTPOptions: genai.HTTPOptions{APIVersion: "v1"},
		Backend:     genai.BackendVertexAI,
		Project:     cfg.ProjectID,
		Location:    "europe-west4",
	})
	if err != nil {
		l.Critical(ctx, map[string]interface{}{
			"message": "failed to create Gemini client",
			"error":   err.Error(),
		})
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Initialize localizer
	localizer := localization.NewLocalizer()

	// Initialize reCAPTCHA service
	recaptchaService := recaptcha.NewService()

	// Initialize waste sorting service
	wasteSortingService := services.NewWasteSortingService(geminiClient, localizer, recaptchaService)

	// Initialize waste sorting handler
	wasteSortingHandler := handlers.NewWasteSortingHandler(wasteSortingService, localizer)

	return &Container{
		Config:              cfg,
		Logger:              l,
		Tracer:              tr,
		Ai:                  geminiClient,
		Localizer:           localizer,
		RecaptchaService:    recaptchaService,
		WasteSortingService: wasteSortingService,
		WasteSortingHandler: wasteSortingHandler,
	}, nil
}