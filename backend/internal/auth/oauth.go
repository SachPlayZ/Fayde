package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// OAuthHandler handles Google and GitHub OAuth flows.
type OAuthHandler struct {
	googleCfg   *oauth2.Config
	githubCfg   *oauth2.Config
	repo        Repository
	svc         *Service
	jwtSecret   string
	frontendURL string
}

// NewOAuthHandler creates an OAuthHandler.
// Pass empty clientID/secret to disable a provider (handler returns 501).
func NewOAuthHandler(
	googleClientID, googleClientSecret,
	githubClientID, githubClientSecret,
	appURL, frontendURL, jwtSecret string,
	repo Repository,
	svc *Service,
) *OAuthHandler {
	h := &OAuthHandler{
		repo:        repo,
		svc:         svc,
		jwtSecret:   jwtSecret,
		frontendURL: frontendURL,
	}

	if googleClientID != "" {
		h.googleCfg = &oauth2.Config{
			ClientID:     googleClientID,
			ClientSecret: googleClientSecret,
			RedirectURL:  appURL + "/auth/google/callback",
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		}
	}

	if githubClientID != "" {
		h.githubCfg = &oauth2.Config{
			ClientID:     githubClientID,
			ClientSecret: githubClientSecret,
			RedirectURL:  appURL + "/auth/github/callback",
			Scopes:       []string{"read:user", "user:email"},
			Endpoint:     github.Endpoint,
		}
	}

	return h
}

// GoogleLogin redirects the user to Google's consent screen.
func (h *OAuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	if h.googleCfg == nil {
		http.Error(w, "Google OAuth not configured", http.StatusNotImplemented)
		return
	}
	state := h.newState()
	http.Redirect(w, r, h.googleCfg.AuthCodeURL(state, oauth2.AccessTypeOnline), http.StatusFound)
}

// GoogleCallback handles Google's redirect back after consent.
func (h *OAuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	if h.googleCfg == nil {
		http.Error(w, "Google OAuth not configured", http.StatusNotImplemented)
		return
	}
	if !h.verifyState(r.URL.Query().Get("state")) {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}
	code := r.URL.Query().Get("code")
	tok, err := h.googleCfg.Exchange(r.Context(), code)
	if err != nil {
		http.Redirect(w, r, h.frontendURL+"/login?error=oauth", http.StatusFound)
		return
	}

	info, err := fetchGoogleUserInfo(tok.AccessToken)
	if err != nil || info.Email == "" {
		http.Redirect(w, r, h.frontendURL+"/login?error=oauth", http.StatusFound)
		return
	}

	h.finishOAuth(w, r, info.Email, "google", info.ID)
}

// GitHubLogin redirects the user to GitHub's consent screen.
func (h *OAuthHandler) GitHubLogin(w http.ResponseWriter, r *http.Request) {
	if h.githubCfg == nil {
		http.Error(w, "GitHub OAuth not configured", http.StatusNotImplemented)
		return
	}
	state := h.newState()
	http.Redirect(w, r, h.githubCfg.AuthCodeURL(state, oauth2.AccessTypeOnline), http.StatusFound)
}

// GitHubCallback handles GitHub's redirect back after consent.
func (h *OAuthHandler) GitHubCallback(w http.ResponseWriter, r *http.Request) {
	if h.githubCfg == nil {
		http.Error(w, "GitHub OAuth not configured", http.StatusNotImplemented)
		return
	}
	if !h.verifyState(r.URL.Query().Get("state")) {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}
	code := r.URL.Query().Get("code")
	tok, err := h.githubCfg.Exchange(r.Context(), code)
	if err != nil {
		http.Redirect(w, r, h.frontendURL+"/login?error=oauth", http.StatusFound)
		return
	}

	email, id, err := fetchGitHubUserInfo(r.Context(), tok.AccessToken)
	if err != nil || email == "" {
		http.Redirect(w, r, h.frontendURL+"/login?error=oauth", http.StatusFound)
		return
	}

	h.finishOAuth(w, r, email, "github", id)
}

// finishOAuth upserts the user, issues a JWT, and redirects to the frontend callback page.
func (h *OAuthHandler) finishOAuth(w http.ResponseWriter, r *http.Request, email, provider, providerID string) {
	user, err := h.repo.UpsertOAuthUser(r.Context(), strings.ToLower(email), provider, providerID)
	if err != nil {
		http.Redirect(w, r, h.frontendURL+"/login?error=oauth", http.StatusFound)
		return
	}

	jwtTok, err := h.svc.IssueTokenForOAuthUser(user)
	if err != nil {
		http.Redirect(w, r, h.frontendURL+"/login?error=oauth", http.StatusFound)
		return
	}

	target := fmt.Sprintf("%s/auth/oauth-callback?token=%s", h.frontendURL, jwtTok)
	http.Redirect(w, r, target, http.StatusFound)
}

// newState returns an HMAC-signed state token valid for 10 minutes.
func (h *OAuthHandler) newState() string {
	nonce := make([]byte, 8)
	_, _ = rand.Read(nonce)
	exp := strconv.FormatInt(time.Now().Add(10*time.Minute).Unix(), 10)
	payload := hex.EncodeToString(nonce) + ":" + exp
	sig := h.sign(payload)
	return payload + "." + sig
}

// verifyState checks the HMAC signature and expiry.
func (h *OAuthHandler) verifyState(state string) bool {
	parts := strings.SplitN(state, ".", 2)
	if len(parts) != 2 {
		return false
	}
	payload, sig := parts[0], parts[1]
	if h.sign(payload) != sig {
		return false
	}
	subParts := strings.SplitN(payload, ":", 2)
	if len(subParts) != 2 {
		return false
	}
	exp, err := strconv.ParseInt(subParts[1], 10, 64)
	if err != nil {
		return false
	}
	return time.Now().Unix() <= exp
}

func (h *OAuthHandler) sign(msg string) string {
	mac := hmac.New(sha256.New, []byte(h.jwtSecret))
	mac.Write([]byte(msg))
	return hex.EncodeToString(mac.Sum(nil))
}

// --- Provider-specific user info fetchers ---

type googleUserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func fetchGoogleUserInfo(accessToken string) (*googleUserInfo, error) {
	req, _ := http.NewRequest(http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("google: non-200 from userinfo")
	}
	var info googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}

type githubUser struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

type githubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func fetchGitHubUserInfo(ctx context.Context, accessToken string) (email, id string, err error) {
	do := func(url string, out any) error {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Accept", "application/vnd.github+json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return json.Unmarshal(body, out)
	}

	var user githubUser
	if err := do("https://api.github.com/user", &user); err != nil {
		return "", "", err
	}
	id = strconv.Itoa(user.ID)

	// Use account email if set publicly.
	if user.Email != "" {
		return user.Email, id, nil
	}

	// Fall back to fetching the primary verified email.
	var emails []githubEmail
	if err := do("https://api.github.com/user/emails", &emails); err != nil {
		return "", id, err
	}
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, id, nil
		}
	}
	return "", id, errors.New("github: no verified primary email found")
}
