//go:build !windows

package lifecycle

import (
	"context"
	"errors"
)

func DoUpgrade(cancel context.CancelFunc, done chan int) error {
	// TODO: Implement this function
	return nil
}
	return errors.New("not implemented")
}
