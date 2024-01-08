package e

import "fmt"

func WrapError(op, msg string, err error) error {
	if err == nil {
		return fmt.Errorf("%s: %s", op, msg)
	}
	return fmt.Errorf("%s: %s due to error: %w", op, msg, err)
}

func WrapString(op, msg string, err error) string {
	if err == nil {
		return fmt.Sprintf("%s: %s", op, msg)
	}
	return fmt.Sprintf("%s: %s due to error: %v", op, msg, err)
}
