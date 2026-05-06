import { main } from '../wailsjs/go/models';

export type NavigationKey = 'home' | 'settings' | 'about';
export type AppInfoModel = main.AppInfo;
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
