package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// Rekey changes the password of an existing SQLCipher database in place by
// re-encrypting it with a new password. The database is opened with its current
// password; its contents and structure are unchanged. Callers that need crash
// safety should rekey a copy and atomically replace the live database.
func Rekey(ctx context.Context, config Config, newPassword string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if strings.TrimSpace(config.Path) == "" {
		return errors.New("database path is required")
	}
	if newPassword == "" {
		return errors.New("new database password is required")
	}

	database, err := openSQLCipher(ctx, config.Path, config.Password)
	if err != nil {
		return err
	}
	defer database.Close()

	statement := fmt.Sprintf(`PRAGMA rekey = "%s"`, escapeSQLCipherPragmaString(newPassword))
	if _, err := database.ExecContext(ctx, statement); err != nil {
		return fmt.Errorf("rekey SQLCipher database: %w", err)
	}
	return nil
}
