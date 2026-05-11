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
	Language   string
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
	data := []byte(renderOperationGuide(input.ProjectID, input.Operations, createdAt, format, input.Language))
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", fmt.Errorf("write operation guide: %w", err)
	}
	return path, nil
}

type operationGuideText struct {
	Title        string
	ProjectID    string
	CreatedAt    string
	Operations   string
	DeleteVerb   string
	MoveVerb     string
	UploadVerb   string
	UnknownVerb  string
	SourceLabel  string
	TargetLabel  string
	ManualNotice string
}

func renderOperationGuide(projectID string, operations []ProjectContentOperation, createdAt time.Time, format string, language string) string {
	var builder strings.Builder
	text := operationGuideTextForLanguage(language)
	switch format {
	case GuideFormatMD:
		builder.WriteString("# " + text.Title + "\n\n")
		builder.WriteString(text.ManualNotice + "\n\n")
		builder.WriteString(fmt.Sprintf("- %s: `%s`\n", text.ProjectID, projectID))
		builder.WriteString(fmt.Sprintf("- %s: `%s`\n\n", text.CreatedAt, createdAt.Format(time.RFC3339)))
		builder.WriteString("## " + text.Operations + "\n\n")
		for index, operation := range operations {
			switch operation.Type {
			case "delete":
				builder.WriteString(fmt.Sprintf("%d. %s `%s`.\n", index+1, text.DeleteVerb, operation.TargetPath))
			default:
				builder.WriteString(fmt.Sprintf("%d. %s `%s` -> `%s`.\n", index+1, guideOperationTitle(operation.Type, text), operation.SourcePath, operation.TargetPath))
			}
		}
	default:
		builder.WriteString(text.Title + "\n")
		builder.WriteString(text.ManualNotice + "\n\n")
		builder.WriteString(fmt.Sprintf("%s: %s\n", text.ProjectID, projectID))
		builder.WriteString(fmt.Sprintf("%s: %s\n\n", text.CreatedAt, createdAt.Format(time.RFC3339)))
		builder.WriteString(text.Operations + ":\n")
		for index, operation := range operations {
			switch operation.Type {
			case "delete":
				builder.WriteString(fmt.Sprintf("%d. %s %s=%s\n", index+1, text.DeleteVerb, text.TargetLabel, operation.TargetPath))
			default:
				builder.WriteString(fmt.Sprintf("%d. %s %s=%s %s=%s\n", index+1, guideOperationTitle(operation.Type, text), text.SourceLabel, operation.SourcePath, text.TargetLabel, operation.TargetPath))
			}
		}
	}
	builder.WriteByte('\n')
	return builder.String()
}

func operationGuideTextForLanguage(language string) operationGuideText {
	if language == LanguageZHCN {
		return operationGuideText{
			Title:        "FoldersGuard 操作指南",
			ProjectID:    "项目 ID",
			CreatedAt:    "创建时间",
			Operations:   "操作步骤",
			DeleteVerb:   "删除",
			MoveVerb:     "移动",
			UploadVerb:   "上传",
			UnknownVerb:  "执行",
			SourceLabel:  "源路径",
			TargetLabel:  "目标路径",
			ManualNotice: "请按以下步骤手动处理加密内容。完成这些步骤前，FG 数据中的修改可能还没有对应到实际加密内容。",
		}
	}
	return operationGuideText{
		Title:        "FoldersGuard Operation Guide",
		ProjectID:    "Project ID",
		CreatedAt:    "Created At",
		Operations:   "Operations",
		DeleteVerb:   "Delete",
		MoveVerb:     "Move",
		UploadVerb:   "Upload",
		UnknownVerb:  "Apply",
		SourceLabel:  "source",
		TargetLabel:  "target",
		ManualNotice: "Follow these steps to manually update encrypted content. Until they are complete, FG data changes may not match the physical encrypted content.",
	}
}

func operationGuideLanguage(requested string, settingsLanguage string) string {
	switch requested {
	case LanguageENUS, LanguageZHCN:
		return requested
	}
	switch settingsLanguage {
	case LanguageZHCN:
		return LanguageZHCN
	default:
		return LanguageENUS
	}
}

func guideOperationTitle(operationType string, text operationGuideText) string {
	switch operationType {
	case "move":
		return text.MoveVerb
	case "upload":
		return text.UploadVerb
	case "":
		return text.UnknownVerb
	default:
		return operationType
	}
}
