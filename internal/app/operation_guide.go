package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type OperationGuideInput struct {
	ProjectID  string
	Operations []ProjectContentOperation
	CreatedAt  time.Time
	Format     string
}

func (s Service) WriteOperationGuide(input OperationGuideInput) (string, error) {
	if len(input.Operations) == 0 {
		return "", nil
	}
	format := input.Format
	if format == "" {
		format = GuideFormatTXT
	}
	switch format {
	case GuideFormatTXT, GuideFormatMD:
	default:
		return "", fmt.Errorf("unsupported operation guide format %q", format)
	}
	if err := os.MkdirAll(s.OperationGuidesDir(), 0o755); err != nil {
		return "", fmt.Errorf("create operation guide directory: %w", err)
	}
	createdAt := input.CreatedAt.UTC()
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	name := fmt.Sprintf("%s-%s.%s", input.ProjectID, createdAt.Format("20060102-150405.000000000"), format)
	path := filepath.Join(s.OperationGuidesDir(), name)
	data := []byte(renderOperationGuide(input.ProjectID, input.Operations, createdAt, format))
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", fmt.Errorf("write operation guide: %w", err)
	}
	return path, nil
}

func renderOperationGuide(projectID string, operations []ProjectContentOperation, createdAt time.Time, format string) string {
	var builder strings.Builder
	switch format {
	case GuideFormatMD:
		builder.WriteString("# FoldersGuard Operation Guide\n\n")
		builder.WriteString(fmt.Sprintf("- Project ID: `%s`\n", projectID))
		builder.WriteString(fmt.Sprintf("- Created At: `%s`\n\n", createdAt.Format(time.RFC3339)))
		builder.WriteString("## Operations\n\n")
		for index, operation := range operations {
			switch operation.Type {
			case "delete":
				builder.WriteString(fmt.Sprintf("%d. Delete `%s`.\n", index+1, operation.TargetPath))
			default:
				builder.WriteString(fmt.Sprintf("%d. %s `%s` to `%s`.\n", index+1, guideOperationTitle(operation.Type), operation.SourcePath, operation.TargetPath))
			}
		}
	default:
		builder.WriteString("FoldersGuard Operation Guide\n")
		builder.WriteString(fmt.Sprintf("Project ID: %s\n", projectID))
		builder.WriteString(fmt.Sprintf("Created At: %s\n\n", createdAt.Format(time.RFC3339)))
		builder.WriteString("Operations:\n")
		for index, operation := range operations {
			switch operation.Type {
			case "delete":
				builder.WriteString(fmt.Sprintf("%d. delete target=%s\n", index+1, operation.TargetPath))
			default:
				builder.WriteString(fmt.Sprintf("%d. %s source=%s target=%s\n", index+1, operation.Type, operation.SourcePath, operation.TargetPath))
			}
		}
	}
	builder.WriteByte('\n')
	return builder.String()
}

func guideOperationTitle(operationType string) string {
	if operationType == "" {
		return ""
	}
	return strings.ToUpper(operationType[:1]) + operationType[1:]
}
