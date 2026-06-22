package apitokens

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) Generate(ctx context.Context, userID, name string) (*CreateResult, error) {
	b := make([]byte, 32)
	rand.Read(b) //nolint:errcheck
	rawToken := "rivz_" + hex.EncodeToString(b)
	h := sha256.Sum256([]byte(rawToken))
	hash := hex.EncodeToString(h[:])
	prefix := rawToken[:12]
	token, err := s.repo.Create(ctx, userID, name, hash, prefix)
	if err != nil {
		return nil, err
	}
	return &CreateResult{Token: rawToken, APIToken: *token}, nil
}

func (s *Service) ValidateToken(ctx context.Context, rawToken string) (*LookupResult, error) {
	h := sha256.Sum256([]byte(rawToken))
	hash := hex.EncodeToString(h[:])
	result, err := s.repo.FindByHash(ctx, hash)
	if err != nil {
		return nil, ErrNotFound
	}
	go s.repo.UpdateLastUsed(context.Background(), result.ID) //nolint:errcheck
	return result, nil
}

func (s *Service) List(ctx context.Context, userID string) ([]*APIToken, error) {
	return s.repo.List(ctx, userID)
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
	err := s.repo.Delete(ctx, id, userID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return ErrNotFound
		}
		return err
	}
	return nil
}
