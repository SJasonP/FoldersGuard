package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"foldersguard/internal/db"
	"foldersguard/internal/format"
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

func (c cli) readDatabasePassword(projectRef string, options passwordOptions) (string, error) {
	if format.IsSetExtension(projectRef) && !options.passwordStdin && options.passwordEnv == "" {
		return db.UnprotectedSharePassword, nil
	}
	return c.readPassword(options)
}

type sharePasswordOptions struct {
	passwordStdin bool
	passwordEnv   string
	noPassword    bool
}

func (c cli) readSharePassword(options sharePasswordOptions) (string, bool, error) {
	selected := 0
	if options.passwordStdin {
		selected++
	}
	if options.passwordEnv != "" {
		selected++
	}
	if options.noPassword {
		selected++
	}
	if selected != 1 {
		return "", false, fmt.Errorf("exactly one share password mode is required")
	}
	if options.noPassword {
		return db.UnprotectedSharePassword, false, nil
	}

	password, err := c.readPassword(passwordOptions{
		passwordStdin: options.passwordStdin,
		passwordEnv:   options.passwordEnv,
	})
	if err != nil {
		return "", false, err
	}
	return password, true, nil
}
