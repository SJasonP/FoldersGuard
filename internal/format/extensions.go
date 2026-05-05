package format

import (
	"path/filepath"
	"strings"
)

type DatabaseShape struct {
	TopLevelObjects int
	SingleTopIsDir  bool
	FromShare       bool
}

func ExtensionForShape(shape DatabaseShape) string {
	if shape.TopLevelObjects == 1 && shape.SingleTopIsDir && !shape.FromShare {
		return ProjectExtension
	}
	return SetExtension
}

func IsProjectExtension(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ProjectExtension)
}

func IsSetExtension(path string) bool {
	return strings.EqualFold(filepath.Ext(path), SetExtension)
}
