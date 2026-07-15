package projects

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/SachPlayZ/rivz-asn/backend/internal/attachments"
	"github.com/google/uuid"
)

const presignExpiry = time.Hour

// Service handles business logic for published project entries.
type Service struct {
	repo     Repository
	s3Client attachments.Storage
	appURL   string
}

// NewService creates a new projects Service. s3Client may be nil if S3 isn't
// configured — image endpoints check for that (see Handler.checkConfigured).
func NewService(repo Repository, s3Client attachments.Storage, appURL string) *Service {
	return &Service{repo: repo, s3Client: s3Client, appURL: appURL}
}

func (s *Service) Create(ctx context.Context, userID string, req CreateRequest) (*Project, error) {
	e, err := s.repo.CreateEntry(ctx, userID, req)
	if err != nil {
		return nil, err
	}
	return s.attachImageURLs(e), nil
}

func (s *Service) List(ctx context.Context, userID string) ([]*Project, error) {
	entries, err := s.repo.ListEntries(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		s.attachImageURLs(e)
	}
	return entries, nil
}

func (s *Service) Get(ctx context.Context, id, userID string) (*Project, error) {
	e, err := s.repo.GetEntry(ctx, id, userID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return s.attachImageURLs(e), nil
}

func (s *Service) Update(ctx context.Context, id, userID string, req UpdateRequest) (*Project, error) {
	e, err := s.repo.UpdateEntry(ctx, id, userID, req)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return s.attachImageURLs(e), nil
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
	if err := s.repo.DeleteEntry(ctx, id, userID); err != nil {
		if err == ErrNotFound || strings.Contains(err.Error(), "no rows") {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// attachImageURLs sets LogoURL/BannerURL to the permanent backend redirect
// path (never a raw presigned S3 URL) so links embedded elsewhere don't expire.
func (s *Service) attachImageURLs(e *Project) *Project {
	if e.logoS3Key != nil {
		u := fmt.Sprintf("%s/projects/%s/logo", s.appURL, e.ID)
		e.LogoURL = &u
	}
	if e.bannerS3Key != nil {
		u := fmt.Sprintf("%s/projects/%s/banner", s.appURL, e.ID)
		e.BannerURL = &u
	}
	return e
}

// PublicList returns all entries for the user matching slug (username or ID),
// with image URLs attached, for the unauthenticated /p/{slug} route.
func (s *Service) PublicList(ctx context.Context, slug string) (*PublicProjects, error) {
	entries, displayName, err := s.repo.ListEntriesByUsername(ctx, slug)
	if err != nil {
		return nil, ErrNotFound
	}
	for _, e := range entries {
		s.attachImageURLs(e)
	}
	return &PublicProjects{OwnerDisplayName: displayName, Projects: entries}, nil
}

// PresignLogo resolves entry id's current logo key and returns a freshly
// presigned S3 URL, for the public 302-redirect route. Returns "" if unset.
func (s *Service) PresignLogo(ctx context.Context, id string) (string, error) {
	logoKey, _, err := s.repo.GetEntryKeys(ctx, id)
	if err != nil {
		return "", ErrNotFound
	}
	if logoKey == nil {
		return "", ErrNotFound
	}
	return s.s3Client.PresignURL(ctx, *logoKey, presignExpiry)
}

// PresignBanner is the banner equivalent of PresignLogo.
func (s *Service) PresignBanner(ctx context.Context, id string) (string, error) {
	_, bannerKey, err := s.repo.GetEntryKeys(ctx, id)
	if err != nil {
		return "", ErrNotFound
	}
	if bannerKey == nil {
		return "", ErrNotFound
	}
	return s.s3Client.PresignURL(ctx, *bannerKey, presignExpiry)
}

// UploadLogo stores a new logo image in S3 and records its key, deleting the
// previous logo (if any) after the swap succeeds.
func (s *Service) UploadLogo(ctx context.Context, id, userID, filename, contentType string, body io.Reader, size int64) (*Project, error) {
	oldKey, _, err := s.repo.GetEntryKeys(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}

	key := fmt.Sprintf("showcase/%s/logo-%s-%s", id, uuid.New().String(), filename)
	if err := s.s3Client.Upload(ctx, key, contentType, body, size); err != nil {
		return nil, fmt.Errorf("projects: upload logo: %w", err)
	}
	if err := s.repo.SetLogoKey(ctx, id, userID, key); err != nil {
		_ = s.s3Client.Delete(ctx, key)
		return nil, err
	}
	if oldKey != nil {
		_ = s.s3Client.Delete(ctx, *oldKey)
	}
	return s.Get(ctx, id, userID)
}

// UploadBanner is the banner equivalent of UploadLogo.
func (s *Service) UploadBanner(ctx context.Context, id, userID, filename, contentType string, body io.Reader, size int64) (*Project, error) {
	_, oldKey, err := s.repo.GetEntryKeys(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}

	key := fmt.Sprintf("showcase/%s/banner-%s-%s", id, uuid.New().String(), filename)
	if err := s.s3Client.Upload(ctx, key, contentType, body, size); err != nil {
		return nil, fmt.Errorf("projects: upload banner: %w", err)
	}
	if err := s.repo.SetBannerKey(ctx, id, userID, key); err != nil {
		_ = s.s3Client.Delete(ctx, key)
		return nil, err
	}
	if oldKey != nil {
		_ = s.s3Client.Delete(ctx, *oldKey)
	}
	return s.Get(ctx, id, userID)
}

func (s *Service) DeleteLogo(ctx context.Context, id, userID string) error {
	oldKey, err := s.repo.ClearLogoKey(ctx, id, userID)
	if err != nil {
		return err
	}
	if oldKey != "" {
		_ = s.s3Client.Delete(ctx, oldKey)
	}
	return nil
}

func (s *Service) DeleteBanner(ctx context.Context, id, userID string) error {
	oldKey, err := s.repo.ClearBannerKey(ctx, id, userID)
	if err != nil {
		return err
	}
	if oldKey != "" {
		_ = s.s3Client.Delete(ctx, oldKey)
	}
	return nil
}
