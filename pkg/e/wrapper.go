package e

import "fmt"

func Wrap(op, msg string, err error) error {
	return fmt.Errorf("%s: %s due to error: %w", op, msg, err)
}
