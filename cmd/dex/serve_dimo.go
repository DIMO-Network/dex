package main

import (
	"fmt"
	"time"

	"github.com/dexidp/dex/storage"
)

// validateIDTokenExpiry checks client's token lifetime override, if it has one.
func validateIDTokenExpiry(client storage.Client) error {
	if client.IDTokenExpiry == "" {
		return nil
	}
	if _, err := time.ParseDuration(client.IDTokenExpiry); err != nil {
		return fmt.Errorf("invalid config: IDTokenExpiry field must be a valid duration for client %q", client.ID)
	}
	return nil
}
