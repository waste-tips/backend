package recaptcha

import (
	"cloud.google.com/go/recaptchaenterprise/v2/apiv1"
	"cloud.google.com/go/recaptchaenterprise/v2/apiv1/recaptchaenterprisepb"
	"context"
	"fmt"
)

// Service handles reCAPTCHA Enterprise verification
type Service struct {
	projectID string
	siteKey   string
}

// NewService creates a new reCAPTCHA service
func NewService(projectId, siteKey string) *Service {
	return &Service{
		projectID: projectId,
		siteKey:   siteKey,
	}
}

// VerifyToken verifies the reCAPTCHA Enterprise token
func (s *Service) VerifyToken(ctx context.Context, token string) (bool, error) {
	if s.projectID == "" || s.siteKey == "" {
		return false, fmt.Errorf("missing reCAPTCHA configuration")
	}

	client, err := recaptchaenterprise.NewClient(ctx)
	if err != nil {
		return false, fmt.Errorf("error creating reCAPTCHA client: %v", err)
	}
	defer client.Close()

	request := &recaptchaenterprisepb.CreateAssessmentRequest{
		Parent: fmt.Sprintf("projects/%s", s.projectID),
		Assessment: &recaptchaenterprisepb.Assessment{
			Event: &recaptchaenterprisepb.Event{
				Token:   token,
				SiteKey: s.siteKey,
			},
		},
	}

	response, err := client.CreateAssessment(ctx, request)
	if err != nil {
		return false, fmt.Errorf("error creating reCAPTCHA assessment: %v", err)
	}

	// Check if token is valid and score is acceptable
	return response.TokenProperties.Valid && response.RiskAnalysis.Score >= 0.5, nil
}
