//go:build !windows && !darwin

package cmd

import (
	"context"
	"errors"

	"github.com/ollama/ollama/api"
)

func startApp(ctx context.Context, client *api.Client) error {
	// TODO: Implement this function
	return nil
}
	return errors.New("could not connect to ollama server, run 'ollama serve' to start it")
}
