package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	configDir  = ".config/gsc-cli"
	configFile = "config"
	configType = "yaml"
)

// Init initializes the configuration system
func Init() error {
	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(path, 0700); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}

	viper.SetConfigName(configFile)
	viper.SetConfigType(configType)
	viper.AddConfigPath(path)

	// Defaults
	viper.SetDefault("site_url", "")
	viper.SetDefault("client_secret_path", "")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			configFilePath := filepath.Join(path, configFile+"."+configType)
			if err := viper.SafeWriteConfigAs(configFilePath); err != nil {
				return fmt.Errorf("could not create config file: %w", err)
			}
		} else {
			return fmt.Errorf("could not read config file: %w", err)
		}
	}

	return nil
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get home directory: %w", err)
	}
	return filepath.Join(home, configDir), nil
}

// GetSiteURL returns the configured Search Console site URL
func GetSiteURL() string {
	return viper.GetString("site_url")
}

// SetSiteURL sets the Search Console site URL
func SetSiteURL(url string) error {
	viper.Set("site_url", url)
	return viper.WriteConfig()
}

// GetClientSecretPath returns the path to the OAuth client secret file
func GetClientSecretPath() string {
	return viper.GetString("client_secret_path")
}

// SetClientSecretPath sets the path to the OAuth client secret file
func SetClientSecretPath(path string) error {
	viper.Set("client_secret_path", path)
	return viper.WriteConfig()
}

// IsConfigured returns true if the CLI has been configured
func IsConfigured() bool {
	return GetSiteURL() != "" && GetClientSecretPath() != ""
}
