// Package email sends transactional emails via the Resend API.
package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Client sends email via Resend.
type Client struct {
	apiKey string
	from   string
}

// New creates a Resend email client.
func New(apiKey, from string) *Client {
	return &Client{apiKey: apiKey, from: from}
}

// SendVerification sends an email verification link to the given address.
func (c *Client) SendVerification(to, verifyURL string) error {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family:system-ui,sans-serif;background:#f9fafb;padding:40px 0;margin:0">
  <div style="max-width:480px;margin:0 auto;background:#fff;border-radius:12px;padding:40px;box-shadow:0 1px 3px rgba(0,0,0,.08)">
    <h1 style="font-size:22px;font-weight:700;margin:0 0 8px">Verify your email</h1>
    <p style="color:#6b7280;margin:0 0 28px;line-height:1.5">
      Click the button below to verify your email address and activate your account.
      The link expires in 24 hours.
    </p>
    <a href="%s"
       style="display:inline-block;background:#000;color:#fff;text-decoration:none;
              padding:12px 24px;border-radius:8px;font-weight:600;font-size:14px">
      Verify Email
    </a>
    <p style="margin:28px 0 0;font-size:12px;color:#9ca3af">
      If you didn't create an account you can safely ignore this email.
    </p>
  </div>
</body>
</html>`, verifyURL)

	return c.send(to, "Verify your email address", html)
}

func (c *Client) send(to, subject, html string) error {
	payload := map[string]any{
		"from":    c.from,
		"to":      []string{to},
		"subject": subject,
		"html":    html,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("email: marshal: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("email: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("email: send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("email: resend returned %d", resp.StatusCode)
	}
	return nil
}
