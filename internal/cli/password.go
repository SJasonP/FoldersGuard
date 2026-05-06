package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"foldersguard/internal/db"

	"golang.org/x/term"
)

type passwordOptions struct {
	passwordStdin bool
	passwordEnv   string
}

func (c cli) readPassword(options passwordOptions) (string, error) {
	return c.readPasswordFor(options, passwordPrompt{
		label:        "Password",
		confirm:      false,
		confirmLabel: "Confirm password",
	})
}

func (c cli) readNewPassword(options passwordOptions) (string, error) {
	return c.readPasswordFor(options, passwordPrompt{
		label:        "Password",
		confirm:      true,
		confirmLabel: "Confirm password",
	})
}

type passwordPrompt struct {
	label        string
	confirm      bool
	confirmLabel string
}

func (c cli) readPasswordFor(options passwordOptions, prompt passwordPrompt) (string, error) {
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
	password, err := c.readInteractivePassword(prompt.label)
	if err != nil {
		return "", err
	}
	if password == "" {
		return "", fmt.Errorf("password must not be empty")
	}
	if prompt.confirm {
		confirmation, err := c.readInteractivePassword(prompt.confirmLabel)
		if err != nil {
			return "", err
		}
		if password != confirmation {
			return "", fmt.Errorf("password confirmation does not match")
		}
	}
	return password, nil
}

func (c cli) readDatabasePassword(projectRef string, options passwordOptions) (string, error) {
	return c.readPassword(options)
}

func hasPasswordInput(options passwordOptions) bool {
	return options.passwordStdin || options.passwordEnv != ""
}

func (c cli) readInteractivePassword(label string) (string, error) {
	file, ok := c.in.(*os.File)
	if !ok || !term.IsTerminal(int(file.Fd())) {
		return "", fmt.Errorf("password input is required")
	}
	if label == "" {
		label = "Password"
	}
	fmt.Fprintf(c.err, "%s: ", label)
	data, err := term.ReadPassword(int(file.Fd()))
	fmt.Fprintln(c.err)
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	return strings.TrimRight(string(data), "\r\n"), nil
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
	if selected > 1 {
		return "", false, fmt.Errorf("only one share password mode may be selected")
	}
	if options.noPassword {
		return db.UnprotectedSharePassword, false, nil
	}

	password, err := c.readPasswordFor(passwordOptions{
		passwordStdin: options.passwordStdin,
		passwordEnv:   options.passwordEnv,
	}, passwordPrompt{
		label:        "Share password",
		confirm:      true,
		confirmLabel: "Confirm share password",
	})
	if err != nil {
		return "", false, err
	}
	return password, true, nil
}
