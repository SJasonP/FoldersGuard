import { main } from '../wailsjs/go/models';

export type NavigationKey = 'home' | 'settings' | 'about';
export type AppInfoModel = main.AppInfo;
export type DeleteProjectRequestModel = main.DeleteProjectRequest;
export type DeleteProjectResultModel = main.DeleteProjectResult;
export type ExportProjectRequestModel = main.ExportProjectRequest;
export type ExportProjectResultModel = main.ExportProjectResult;
export type InspectProjectRequestModel = main.InspectProjectRequest;
export type InspectProjectResultModel = main.InspectProjectResult;
export type LocalProjectSummary = main.LocalProjectSummary;
export type SettingsModel = main.Settings;

export type LocalProjectRow = {
  key: string;
  projectId: string;
  fileName: string;
  modifiedTime: string;
  availabilityStatus: string;
};
