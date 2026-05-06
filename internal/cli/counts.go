package cli

import "foldersguard/internal/model"

func countFolders(plan model.PlannedProject) int {
	if plan.Project.DatabaseType == "share" && plan.RootItem.RealName == "" {
		return len(plan.Folders)
	}
	return len(plan.Folders) + 1
}
