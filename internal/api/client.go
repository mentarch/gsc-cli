package api

import (
	"context"
	"fmt"

	"gsc-cli/internal/auth"
	"gsc-cli/internal/config"

	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/searchconsole/v1"
)

// Client wraps the Google Search Console API client
type Client struct {
	service *searchconsole.Service
	siteURL string
}

// NewClientForSites creates a client for listing sites (no site URL required)
func NewClientForSites() (*Client, error) {
	return NewClient("")
}

// NewClient creates a new Search Console API client
func NewClient(siteURL string) (*Client, error) {
	clientSecretPath := config.GetClientSecretPath()
	if clientSecretPath == "" {
		return nil, fmt.Errorf("not configured - run 'gsc auth login' first")
	}

	token, err := auth.GetValidToken(clientSecretPath)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	// Create HTTP client with OAuth2 token
	oauthConfig, err := auth.LoadClientConfig(clientSecretPath)
	if err != nil {
		return nil, err
	}

	httpClient := oauthConfig.Client(ctx, token)

	// Create Search Console service
	service, err := searchconsole.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("could not create Search Console service: %w", err)
	}

	return &Client{
		service: service,
		siteURL: siteURL,
	}, nil
}

// NewClientFromToken creates a client using a provided token (for testing)
func NewClientFromToken(siteURL string, token *oauth2.Token, oauthConfig *oauth2.Config) (*Client, error) {
	ctx := context.Background()
	httpClient := oauthConfig.Client(ctx, token)

	service, err := searchconsole.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("could not create Search Console service: %w", err)
	}

	return &Client{
		service: service,
		siteURL: siteURL,
	}, nil
}

// ListSites returns all sites the user has access to
func (c *Client) ListSites() ([]*searchconsole.WmxSite, error) {
	resp, err := c.service.Sites.List().Do()
	if err != nil {
		return nil, fmt.Errorf("could not list sites: %w", err)
	}
	return resp.SiteEntry, nil
}

// GetSiteURL returns the configured site URL
func (c *Client) GetSiteURL() string {
	return c.siteURL
}

// Service returns the underlying Search Console service
func (c *Client) Service() *searchconsole.Service {
	return c.service
}
