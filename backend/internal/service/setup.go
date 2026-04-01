package service

import (
	"context"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

// SetupService orchestrates first-run setup and setup-wizard state.
type SetupService struct {
	authSvc         *AuthService
	appSettingsRepo *repository.AppSettingsRepo
}

// NewSetupService creates a new setup service.
func NewSetupService(authSvc *AuthService, appSettingsRepo *repository.AppSettingsRepo) *SetupService {
	return &SetupService{
		authSvc:         authSvc,
		appSettingsRepo: appSettingsRepo,
	}
}

// Status returns whether the system has already been configured.
func (s *SetupService) Status(ctx context.Context) (bool, error) {
	return s.authSvc.IsConfigured(ctx)
}

// CreateFirstAdmin creates the first admin user for initial setup.
func (s *SetupService) CreateFirstAdmin(
	ctx context.Context,
	username, password, displayName string,
) (*model.User, error) {
	return s.authSvc.CreateFirstAdmin(ctx, username, password, displayName)
}

// WizardCompleted returns whether the setup wizard has been completed.
func (s *SetupService) WizardCompleted(ctx context.Context) (bool, error) {
	val, err := s.appSettingsRepo.Get(ctx, model.SettingSetupWizardCompleted)
	if err != nil {
		return false, err
	}
	return val == "true", nil
}

// CompleteWizard marks the setup wizard as completed.
func (s *SetupService) CompleteWizard(ctx context.Context) error {
	return s.appSettingsRepo.Set(ctx, model.SettingSetupWizardCompleted, "true")
}
