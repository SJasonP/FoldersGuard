package cli

import (
	"foldersguard/internal/app"
	"foldersguard/internal/model"
)

func countFolders(plan model.PlannedProject) int {
	return app.CountFolders(plan)
}
