package showcase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/SachPlayZ/rivz-asn/backend/internal/attachments"
	"github.com/google/uuid"
)

const (
	tokenPrefix   = "fayde_pub_"
	presignExpiry = time.Hour
)

// Service handles business logic for showcase entries and their tokens.
type Service struct {
	repo     Repository
	s3Client attachments.Storage
	appURL   string
}

// NewService creates a new showcase Service. s3Client may be nil if S3 isn't
// configured — image endpoints check for that (see Handler.checkConfigured).
func NewService(repo Repository, s3Client attachments.Storage, appURL string) *Service {
	return &Service{repo: repo, s3Client: s3Client, appURL: appURL}
}

func (s *Service) Create(ctx context.Context, userID string, req CreateRequest) (*Entry, error) {
	e, err := s.repo.CreateEntry(ctx, userID, req)
	if err != nil {
		return nil, err
	}
	return s.attachImageURLs(e), nil
}

func (s *Service) List(ctx context.Context, userID string) ([]*Entry, error) {
	entries, err := s.repo.ListEntries(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		s.attachImageURLs(e)
	}
	return entries, nil
}

func (s *Service) Get(ctx context.Context, id, userID string) (*Entry, error) {
	e, err := s.repo.GetEntry(ctx, id, userID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return s.attachImageURLs(e), nil
}

func (s *Service) Update(ctx context.Context, id, userID string, req UpdateRequest) (*Entry, error) {
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
func (s *Service) attachImageURLs(e *Entry) *Entry {
	if e.logoS3Key != nil {
		u := fmt.Sprintf("%s/showcase/%s/logo", s.appURL, e.ID)
		e.LogoURL = &u
	}
	if e.bannerS3Key != nil {
		u := fmt.Sprintf("%s/showcase/%s/banner", s.appURL, e.ID)
		e.BannerURL = &u
	}
	return e
}

// PublicList returns all entries for the user matching slug (username or ID),
// with image URLs attached, for the unauthenticated /p/{slug} route.
func (s *Service) PublicList(ctx context.Context, slug string) (*PublicEntries, error) {
	entries, displayName, err := s.repo.ListEntriesByUsername(ctx, slug)
	if err != nil {
		return nil, ErrNotFound
	}
	for _, e := range entries {
		s.attachImageURLs(e)
	}
	return &PublicEntries{OwnerDisplayName: displayName, Entries: entries}, nil
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
func (s *Service) UploadLogo(ctx context.Context, id, userID, filename, contentType string, body io.Reader, size int64) (*Entry, error) {
	oldKey, _, err := s.repo.GetEntryKeys(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}

	key := fmt.Sprintf("showcase/%s/logo-%s-%s", id, uuid.New().String(), filename)
	if err := s.s3Client.Upload(ctx, key, contentType, body, size); err != nil {
		return nil, fmt.Errorf("showcase: upload logo: %w", err)
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
func (s *Service) UploadBanner(ctx context.Context, id, userID, filename, contentType string, body io.Reader, size int64) (*Entry, error) {
	_, oldKey, err := s.repo.GetEntryKeys(ctx, id)
	if err != nil {
		return nil, ErrNotFound
	}

	key := fmt.Sprintf("showcase/%s/banner-%s-%s", id, uuid.New().String(), filename)
	if err := s.s3Client.Upload(ctx, key, contentType, body, size); err != nil {
		return nil, fmt.Errorf("showcase: upload banner: %w", err)
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

// GenerateToken creates a new fayde_pub_-prefixed read-only showcase token.
// The raw token is only ever returned here, once — only its hash is stored.
func (s *Service) GenerateToken(ctx context.Context, userID, name string) (*CreateTokenResult, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("showcase: generate token: %w", err)
	}
	rawToken := tokenPrefix + hex.EncodeToString(b)
	h := sha256.Sum256([]byte(rawToken))
	hash := hex.EncodeToString(h[:])
	prefix := rawToken[:len(tokenPrefix)+8]

	t, err := s.repo.CreateToken(ctx, userID, name, hash, prefix)
	if err != nil {
		return nil, err
	}
	return &CreateTokenResult{Token: rawToken, ShowcaseToken: *t}, nil
}

// ValidateShowcaseToken implements auth.ShowcaseTokenValidator.
func (s *Service) ValidateShowcaseToken(ctx context.Context, rawToken string) (string, error) {
	h := sha256.Sum256([]byte(rawToken))
	hash := hex.EncodeToString(h[:])
	result, err := s.repo.FindTokenByHash(ctx, hash)
	if err != nil {
		return "", ErrNotFound
	}
	go s.repo.UpdateTokenLastUsed(context.Background(), result.ID) //nolint:errcheck
	return result.UserID, nil
}

func (s *Service) ListTokens(ctx context.Context, userID string) ([]*ShowcaseToken, error) {
	return s.repo.ListTokens(ctx, userID)
}

func (s *Service) DeleteToken(ctx context.Context, id, userID string) error {
	err := s.repo.DeleteToken(ctx, id, userID)
	if err != nil {
		if err == ErrNotFound || strings.Contains(err.Error(), "no rows") {
			return ErrNotFound
		}
		return err
	}
	return nil
}
