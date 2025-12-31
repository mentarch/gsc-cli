package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/pkg/browser"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	// Search Console API scope
	searchConsoleScope = "https://www.googleapis.com/auth/webmasters.readonly"
)

// LoadClientConfig loads OAuth2 config from client_secret.json
func LoadClientConfig(clientSecretPath string) (*oauth2.Config, error) {
	data, err := os.ReadFile(clientSecretPath)
	if err != nil {
		return nil, fmt.Errorf("could not read client secret file: %w", err)
	}

	config, err := google.ConfigFromJSON(data, searchConsoleScope)
	if err != nil {
		return nil, fmt.Errorf("could not parse client secret: %w", err)
	}

	return config, nil
}

// LoginFlow performs the OAuth2 login flow with browser-based consent
func LoginFlow(clientSecretPath string) (*oauth2.Token, error) {
	config, err := LoadClientConfig(clientSecretPath)
	if err != nil {
		return nil, err
	}

	// Find an available port for the callback server
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, fmt.Errorf("could not start callback server: %w", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	config.RedirectURL = fmt.Sprintf("http://localhost:%d/callback", port)

	// Channel to receive the auth code
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	// Start local server to handle callback
	server := &http.Server{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errMsg := r.URL.Query().Get("error")
			if errMsg == "" {
				errMsg = "no authorization code received"
			}
			errChan <- fmt.Errorf("authorization failed: %s", errMsg)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "<html><body><h1>Authorization Failed</h1><p>%s</p></body></html>", errMsg)
			return
		}

		codeChan <- code
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `<html><body>
			<h1>Authorization Successful!</h1>
			<p>You can close this window and return to the terminal.</p>
			<script>window.close();</script>
		</body></html>`)
	})

	go func() {
		if err := server.Serve(listener); err != http.ErrServerClosed {
			errChan <- fmt.Errorf("callback server error: %w", err)
		}
	}()

	// Generate authorization URL and open browser
	authURL := config.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.ApprovalForce)

	fmt.Println("Opening browser for authorization...")
	fmt.Println("If the browser doesn't open, visit this URL:")
	fmt.Println(authURL)
	fmt.Println()

	if err := browser.OpenURL(authURL); err != nil {
		fmt.Printf("Could not open browser automatically: %v\n", err)
	}

	// Wait for authorization code or error
	var authCode string
	select {
	case authCode = <-codeChan:
		// Success
	case err := <-errChan:
		server.Shutdown(context.Background())
		return nil, err
	case <-time.After(5 * time.Minute):
		server.Shutdown(context.Background())
		return nil, fmt.Errorf("authorization timed out after 5 minutes")
	}

	server.Shutdown(context.Background())

	// Exchange code for token
	ctx := context.Background()
	token, err := config.Exchange(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf("could not exchange authorization code: %w", err)
	}

	return token, nil
}

// RefreshToken refreshes an expired token
func RefreshToken(clientSecretPath string, token *oauth2.Token) (*oauth2.Token, error) {
	config, err := LoadClientConfig(clientSecretPath)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	tokenSource := config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("could not refresh token: %w", err)
	}

	return newToken, nil
}

// GetValidToken returns a valid token, refreshing if necessary
func GetValidToken(clientSecretPath string) (*oauth2.Token, error) {
	token, err := GetToken()
	if err != nil {
		return nil, err
	}
	if token == nil {
		return nil, fmt.Errorf("not logged in - run 'gsc auth login' first")
	}

	// Check if token is expired
	if token.Expiry.Before(time.Now()) {
		token, err = RefreshToken(clientSecretPath, token)
		if err != nil {
			return nil, err
		}
		// Save refreshed token
		if err := SetToken(token); err != nil {
			return nil, err
		}
	}

	return token, nil
}

// TokenInfo returns basic info about the stored token
type TokenInfo struct {
	HasToken    bool
	Expiry      time.Time
	IsExpired   bool
	TokenType   string
}

// GetTokenInfo returns information about the stored token
func GetTokenInfo() (*TokenInfo, error) {
	token, err := GetToken()
	if err != nil {
		return nil, err
	}

	if token == nil {
		return &TokenInfo{HasToken: false}, nil
	}

	return &TokenInfo{
		HasToken:  true,
		Expiry:    token.Expiry,
		IsExpired: token.Expiry.Before(time.Now()),
		TokenType: token.TokenType,
	}, nil
}

// parseClientSecret extracts basic info from client_secret.json
func parseClientSecret(path string) (projectID string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	var secret struct {
		Installed struct {
			ProjectID string `json:"project_id"`
		} `json:"installed"`
		Web struct {
			ProjectID string `json:"project_id"`
		} `json:"web"`
	}

	if err := json.Unmarshal(data, &secret); err != nil {
		return "", err
	}

	if secret.Installed.ProjectID != "" {
		return secret.Installed.ProjectID, nil
	}
	return secret.Web.ProjectID, nil
}
