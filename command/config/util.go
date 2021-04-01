package config

import (
	"github.com/jiangjiali/vault/command/token"
)

// DefaultTokenHelper returns the token helper that is configured for Vault.
func DefaultTokenHelper() (token.THelper, error) {
	config, err := LoadConfig("")
	if err != nil {
		return nil, err
	}

	path := config.TokenHelper
	if path == "" {
		return &token.InternalTokenHelper{}, nil
	}

	path, err = token.ExternalTokenHelperPath(path)
	if err != nil {
		return nil, err
	}
	return &token.ExternalTokenHelper{BinaryPath: path}, nil
}
