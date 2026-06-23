package totp

import (
	"context"
	"fmt"

	"github.com/pquerna/otp/totp"
)

// Service handles TOTP business logic.
type Service struct {
	repo Repository
}

// NewService creates a new TOTP Service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// GenerateSecret creates a new TOTP secret for the user and returns setup info.
func (s *Service) GenerateSecret(ctx context.Context, userID, userEmail string) (*SetupResponse, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Fayde",
		AccountName: userEmail,
		SecretSize:  20,
	})
	if err != nil {
		return nil, fmt.Errorf("totp.GenerateSecret: %w", err)
	}

	if _, err := s.repo.Create(ctx, userID, key.Secret()); err != nil {
		return nil, fmt.Errorf("totp.GenerateSecret: %w", err)
	}

	return &SetupResponse{
		Secret: key.Secret(),
		QRURL:  key.URL(),
	}, nil
}

// Enable validates the TOTP code and enables TOTP for the user.
func (s *Service) Enable(ctx context.Context, userID, code string) error {
	rec, err := s.repo.Get(ctx, userID)
	if err != nil {
		return ErrNotFound
	}
	if !totp.Validate(code, rec.Secret) {
		return ErrInvalidCode
	}
	return s.repo.Enable(ctx, userID)
}

// Verify checks whether a TOTP code is valid for the user.
// If TOTP is not enabled, it returns true (always pass).
func (s *Service) Verify(ctx context.Context, userID, code string) (bool, error) {
	rec, err := s.repo.Get(ctx, userID)
	if err != nil {
		return false, ErrNotFound
	}
	if !rec.Enabled {
		return true, nil
	}
	return totp.Validate(code, rec.Secret), nil
}

// Disable validates the TOTP code and disables TOTP for the user.
func (s *Service) Disable(ctx context.Context, userID, code string) error {
	rec, err := s.repo.Get(ctx, userID)
	if err != nil {
		return ErrNotFound
	}
	if !totp.Validate(code, rec.Secret) {
		return ErrInvalidCode
	}
	return s.repo.Disable(ctx, userID)
}

// IsTOTPEnabled returns whether TOTP is enabled for the user.
func (s *Service) IsTOTPEnabled(ctx context.Context, userID string) (bool, error) {
	return s.repo.GetUserTOTPEnabled(ctx, userID)
}
