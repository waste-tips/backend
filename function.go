package backend

import (
	"github.com/DeryabinSergey/waste-tips-backend/internal/domain"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	functions.HTTP("Invoke", domain.Invoke)
}
