package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type passwordOptions struct {
	passwordStdin bool
	passwordEnv   string
}

func (c cli) readPassword(options passwordOptions) (string, error) {
	if options.passwordStdin {
		data, err := io.ReadAll(c.in)
		if err != nil {
			return "", fmt.Errorf("read password from stdin: %w", err)
		}
		password := strings.TrimRight(string(data), "\r\n")
		if password == "" {
			return "", fmt.Errorf("password must not be empty")
		}
		return password, nil
	}
	if options.passwordEnv != "" {
		password := os.Getenv(options.passwordEnv)
		if password == "" {
			return "", fmt.Errorf("password environment variable %s is empty or unset", options.passwordEnv)
		}
		return password, nil
	}
	return "", fmt.Errorf("password input is required")
}
