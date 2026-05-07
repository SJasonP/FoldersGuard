import { main } from '../wailsjs/go/models';

export type NavigationKey = 'home' | 'settings' | 'about';
export type AppInfoModel = main.AppInfo;
export type ApplyProjectChangesResultModel = main.ApplyProjectChangesResult;
export type CreateProjectRequestModel = main.CreateProjectRequest;
export type CreateProjectResultModel = main.CreateProjectResult;
export type CreateShareResultModel = main.CreateShareResult;
export type DecryptProjectResultModel = main.DecryptProjectResult;
export type DecryptShareResultModel = main.DecryptShareResult;
export type DeleteProjectRequestModel = main.DeleteProjectRequest;
export type DeleteProjectResultModel = main.DeleteProjectResult;
export type ExportProjectRequestModel = main.ExportProjectRequest;
export type ExportProjectResultModel = main.ExportProjectResult;
export type ImportProjectRequestModel = main.ImportProjectRequest;
export type ImportProjectResultModel = main.ImportProjectResult;
export type InspectProjectRequestModel = main.InspectProjectRequest;
export type InspectProjectResultModel = main.InspectProjectResult;
export type LocalProjectSummary = main.LocalProjectSummary;
export type ProjectBrowserItemModel = main.ProjectBrowserItem;
export type ProjectBrowserStateModel = main.ProjectBrowserState;
export type ShareableItemModel = main.ShareableItem;
export type SettingsModel = main.Settings;
export type ShareSummaryModel = main.ShareSummary;
export type VerifyProjectRequestModel = main.VerifyProjectRequest;
export type VerifyProjectResultModel = main.VerifyProjectResult;

export type LocalProjectRow = {
  key: string;
  projectId: string;
  fileName: string;
  modifiedTime: string;
  availabilityStatus: string;
};
