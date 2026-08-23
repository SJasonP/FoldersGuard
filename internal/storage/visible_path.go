package storage

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// validateStorageVisiblePath enforces the on-disk format: encrypted object
// paths are relative slash-separated UUIDs. Besides detecting malformed
// databases, this prevents imported metadata from escaping a content root.
func validateStorageVisiblePath(value string) error {
	if value == "" {
		return fmt.Errorf("is required")
	}
	if strings.HasPrefix(value, "/") || strings.HasSuffix(value, "/") || strings.Contains(value, "\\") {
		return fmt.Errorf("must be a relative slash-separated UUID path")
	}
	for _, component := range strings.Split(value, "/") {
		if component == "" {
			return fmt.Errorf("contains an empty component")
		}
		parsed, err := uuid.Parse(component)
		if err != nil || parsed.String() != component {
			return fmt.Errorf("component %q is not a canonical UUID", component)
		}
	}
	return nil
}
