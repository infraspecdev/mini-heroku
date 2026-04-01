// cli/keychain/keychain.go
package keychain

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	service = "mini-heroku"
	account = "api-key"
)

// Set stores apiKey in the OS keychain.
func Set(apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("api key cannot be empty")
	}
	return keyring.Set(service, account, apiKey)
}

// Get retrieves the API key from the OS keychain.
// Returns a descriptive error when no key is stored yet.
func Get() (string, error) {
	key, err := keyring.Get(service, account)
	if err == keyring.ErrNotFound {
		return "", fmt.Errorf("no API key found — run: mini config set-api-key <key>")
	}
	if err != nil {
		return "", fmt.Errorf("reading keychain: %w", err)
	}
	return key, nil
}

// Delete removes the stored API key from the OS keychain.
func Delete() error {
	return keyring.Delete(service, account)
}