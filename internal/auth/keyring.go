package auth

import (
	"encoding/json"
	"fmt"

	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
)

const serviceName = "gsc-cli"
const tokenKey = "oauth_token"

// GetToken retrieves the OAuth token from the OS keychain
func GetToken() (*oauth2.Token, error) {
	data, err := keyring.Get(serviceName, tokenKey)
	if err != nil {
		if err == keyring.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("could not retrieve token from keyring: %w", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal([]byte(data), &token); err != nil {
		return nil, fmt.Errorf("could not decode token: %w", err)
	}

	return &token, nil
}

// SetToken stores the OAuth token in the OS keychain
func SetToken(token *oauth2.Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("could not encode token: %w", err)
	}

	if err := keyring.Set(serviceName, tokenKey, string(data)); err != nil {
		return fmt.Errorf("could not store token in keyring: %w", err)
	}

	return nil
}

// DeleteToken removes the OAuth token from the keychain
func DeleteToken() error {
	if err := keyring.Delete(serviceName, tokenKey); err != nil {
		if err == keyring.ErrNotFound {
			return nil
		}
		return fmt.Errorf("could not delete token from keyring: %w", err)
	}
	return nil
}

// HasToken checks if a token exists in the keychain
func HasToken() bool {
	token, _ := GetToken()
	return token != nil
}
