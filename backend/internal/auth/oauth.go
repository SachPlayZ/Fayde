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
	platform := r.URL.Query().Get("platform")
	state := h.newState(platform)
	http.Redirect(w, r, h.googleCfg.AuthCodeURL(state, oauth2.AccessTypeOnline), http.StatusFound)
}

// GoogleCallback handles Google's redirect back after consent.
func (h *OAuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	if h.googleCfg == nil {
		http.Error(w, "Google OAuth not configured", http.StatusNotImplemented)
		return
	}
	platform, ok := h.verifyState(r.URL.Query().Get("state"))
	if !ok {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}
	code := r.URL.Query().Get("code")
	tok, err := h.googleCfg.Exchange(r.Context(), code)
	if err != nil {
		h.redirectWithError(w, r, platform)
		return
	}

	info, err := fetchGoogleUserInfo(tok.AccessToken)
	if err != nil || info.Email == "" {
		h.redirectWithError(w, r, platform)
		return
	}

	h.finishOAuth(w, r, info.Email, "google", info.ID, info.Name, info.Picture, platform)
}

// GitHubLogin redirects the user to GitHub's consent screen.
func (h *OAuthHandler) GitHubLogin(w http.ResponseWriter, r *http.Request) {
	if h.githubCfg == nil {
		http.Error(w, "GitHub OAuth not configured", http.StatusNotImplemented)
		return
	}
	platform := r.URL.Query().Get("platform")
	state := h.newState(platform)
	http.Redirect(w, r, h.githubCfg.AuthCodeURL(state, oauth2.AccessTypeOnline), http.StatusFound)
}

// GitHubCallback handles GitHub's redirect back after consent.
func (h *OAuthHandler) GitHubCallback(w http.ResponseWriter, r *http.Request) {
	if h.githubCfg == nil {
		http.Error(w, "GitHub OAuth not configured", http.StatusNotImplemented)
		return
	}
	platform, ok := h.verifyState(r.URL.Query().Get("state"))
	if !ok {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}
	code := r.URL.Query().Get("code")
	tok, err := h.githubCfg.Exchange(r.Context(), code)
	if err != nil {
		h.redirectWithError(w, r, platform)
		return
	}

	email, id, name, avatarURL, err := fetchGitHubUserInfo(r.Context(), tok.AccessToken)
	if err != nil || email == "" {
		h.redirectWithError(w, r, platform)
		return
	}

	h.finishOAuth(w, r, email, "github", id, name, avatarURL, platform)
}

// finishOAuth upserts the user, issues a JWT, and redirects to the frontend callback page.
func (h *OAuthHandler) finishOAuth(w http.ResponseWriter, r *http.Request, email, provider, providerID, displayName, avatarURL, platform string) {
	user, err := h.repo.UpsertOAuthUser(r.Context(), strings.ToLower(email), provider, providerID, displayName, avatarURL)
	if err != nil {
		h.redirectWithError(w, r, platform)
		return
	}

	jwtTok, err := h.svc.IssueTokenForOAuthUser(user)
	if err != nil {
		h.redirectWithError(w, r, platform)
		return
	}

	isDesktop := platform == "desktop"

	if isDesktop {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Authenticating with Fayde</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            background-color: #0a0a0a;
            color: #ffffff;
            display: flex;
            align-items: center;
            justify-content: center;
            height: 100vh;
            margin: 0;
            text-align: center;
        }
        .container {
            max-width: 400px;
            padding: 40px;
            background-color: #121212;
            border: 1px solid #1f1f1f;
            border-radius: 16px;
            box-shadow: 0 4px 30px rgba(0, 0, 0, 0.7);
        }
        h1 {
            font-size: 24px;
            margin-bottom: 12px;
            color: #ffffff;
        }
        p {
            font-size: 14px;
            color: #a0a0a0;
            margin-bottom: 28px;
            line-height: 1.5;
        }
        .btn {
            display: inline-block;
            background-color: #ffffff;
            color: #000000;
            font-weight: 600;
            padding: 12px 24px;
            border-radius: 8px;
            text-decoration: none;
            transition: opacity 0.2s;
        }
        .btn:hover {
            opacity: 0.9;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Successfully Authenticated!</h1>
        <p>Your browser should have opened the Fayde desktop app. You can safely close this browser tab now.</p>
        <a href="fayde://auth/oauth-callback?token=%s" class="btn">Open Fayde</a>
    </div>
    <script>
        // Trigger the deep link callback for the desktop app via an iframe
        // to prevent the browser window from getting stuck in an infinite loading/pending state.
        const iframe = document.createElement('iframe');
        iframe.style.display = 'none';
        iframe.src = "fayde://auth/oauth-callback?token=%s";
        document.body.appendChild(iframe);

        setTimeout(function() {
            window.close();
        }, 5000);
    </script>
</body>
</html>
		`, jwtTok, jwtTok)
		return
	}

	target := fmt.Sprintf("%s/auth/oauth-callback?token=%s", h.frontendURL, jwtTok)
	http.Redirect(w, r, target, http.StatusFound)
}

func (h *OAuthHandler) redirectWithError(w http.ResponseWriter, r *http.Request, platform string) {
	isDesktop := platform == "desktop"

	if isDesktop {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Authentication Failed</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            background-color: #0a0a0a;
            color: #ffffff;
            display: flex;
            align-items: center;
            justify-content: center;
            height: 100vh;
            margin: 0;
            text-align: center;
        }
        .container {
            max-width: 400px;
            padding: 40px;
            background-color: #121212;
            border: 1px solid #1f1f1f;
            border-radius: 16px;
            box-shadow: 0 4px 30px rgba(0, 0, 0, 0.7);
        }
        h1 {
            font-size: 24px;
            margin-bottom: 12px;
            color: #ef4444;
        }
        p {
            font-size: 14px;
            color: #a0a0a0;
            margin-bottom: 28px;
            line-height: 1.5;
        }
        .btn {
            display: inline-block;
            background-color: #ef4444;
            color: #ffffff;
            font-weight: 600;
            padding: 12px 24px;
            border-radius: 8px;
            text-decoration: none;
            transition: opacity 0.2s;
        }
        .btn:hover {
            opacity: 0.9;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Authentication Failed</h1>
        <p>Something went wrong during login. Please return to the app and try again.</p>
        <a href="fayde://auth/oauth-callback?error=oauth" class="btn">Return to Fayde</a>
    </div>
    <script>
        const iframe = document.createElement('iframe');
        iframe.style.display = 'none';
        iframe.src = "fayde://auth/oauth-callback?error=oauth";
        document.body.appendChild(iframe);
    </script>
</body>
</html>
		`)
		return
	}

	target := h.frontendURL + "/login?error=oauth"
	http.Redirect(w, r, target, http.StatusFound)
}

// newState returns an HMAC-signed state token valid for 10 minutes.
func (h *OAuthHandler) newState(platform string) string {
	nonce := make([]byte, 8)
	_, _ = rand.Read(nonce)
	exp := strconv.FormatInt(time.Now().Add(10*time.Minute).Unix(), 10)
	payload := hex.EncodeToString(nonce) + ":" + exp + ":" + platform
	sig := h.sign(payload)
	return payload + "." + sig
}

// verifyState checks the HMAC signature and expiry.
func (h *OAuthHandler) verifyState(state string) (string, bool) {
	parts := strings.SplitN(state, ".", 2)
	if len(parts) != 2 {
		return "", false
	}
	payload, sig := parts[0], parts[1]
	if h.sign(payload) != sig {
		return "", false
	}
	subParts := strings.Split(payload, ":")
	if len(subParts) < 2 {
		return "", false
	}
	exp, err := strconv.ParseInt(subParts[1], 10, 64)
	if err != nil {
		return "", false
	}
	if time.Now().Unix() > exp {
		return "", false
	}
	platform := ""
	if len(subParts) >= 3 {
		platform = subParts[2]
	}
	return platform, true
}

func (h *OAuthHandler) sign(msg string) string {
	mac := hmac.New(sha256.New, []byte(h.jwtSecret))
	mac.Write([]byte(msg))
	return hex.EncodeToString(mac.Sum(nil))
}

// --- Provider-specific user info fetchers ---

type googleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
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
	ID        int    `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

type githubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func fetchGitHubUserInfo(ctx context.Context, accessToken string) (email, id, name, avatarURL string, err error) {
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
		return "", "", "", "", err
	}
	id = strconv.Itoa(user.ID)
	name = user.Name
	avatarURL = user.AvatarURL

	// Use account email if set publicly.
	if user.Email != "" {
		return user.Email, id, name, avatarURL, nil
	}

	// Fall back to fetching the primary verified email.
	var emails []githubEmail
	if err := do("https://api.github.com/user/emails", &emails); err != nil {
		return "", id, name, avatarURL, err
	}
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, id, name, avatarURL, nil
		}
	}
	return "", id, name, avatarURL, errors.New("github: no verified primary email found")
}
