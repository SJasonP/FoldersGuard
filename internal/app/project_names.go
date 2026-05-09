package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (s Service) ProjectNamesPath() string {
	return filepath.Join(s.DataDir, "project-names.json")
}

func (s Service) readProjectNames() (map[string]string, error) {
	if err := s.EnsureDataDir(); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(s.ProjectNamesPath())
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("read project names: %w", err)
	}

	names := map[string]string{}
	if err := json.Unmarshal(data, &names); err != nil {
		return nil, fmt.Errorf("decode project names: %w", err)
	}
	return names, nil
}

func (s Service) writeProjectNames(names map[string]string) error {
	if err := s.EnsureDataDir(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(names, "", "  ")
	if err != nil {
		return fmt.Errorf("encode project names: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(s.ProjectNamesPath(), data, 0o600); err != nil {
		return fmt.Errorf("write project names: %w", err)
	}
	return nil
}

func normalizeLocalProjectName(projectID, projectName string) string {
	trimmed := strings.TrimSpace(projectName)
	if trimmed == "" {
		return projectID
	}
	return trimmed
}

func (s Service) localProjectName(projectID string, names map[string]string) string {
	return normalizeLocalProjectName(projectID, names[projectID])
}

func (s Service) SaveLocalProjectName(input SaveLocalProjectNameInput) (SaveLocalProjectNameResult, error) {
	if input.ProjectID == "" {
		return SaveLocalProjectNameResult{}, fmt.Errorf("project id is required")
	}
	projectPath, err := s.ActiveProjectDatabasePath(input.ProjectID)
	if err != nil {
		return SaveLocalProjectNameResult{}, err
	}
	if err := ValidateDatabasePath(projectPath); err != nil {
		return SaveLocalProjectNameResult{}, err
	}

	names, err := s.readProjectNames()
	if err != nil {
		return SaveLocalProjectNameResult{}, err
	}
	name := normalizeLocalProjectName(input.ProjectID, input.ProjectName)
	names[input.ProjectID] = name
	if err := s.writeProjectNames(names); err != nil {
		return SaveLocalProjectNameResult{}, err
	}
	return SaveLocalProjectNameResult{
		ProjectID:   input.ProjectID,
		ProjectName: name,
	}, nil
}

func (s Service) deleteLocalProjectName(projectID string) error {
	names, err := s.readProjectNames()
	if err != nil {
		return err
	}
	if _, ok := names[projectID]; !ok {
		return nil
	}
	delete(names, projectID)
	return s.writeProjectNames(names)
}
