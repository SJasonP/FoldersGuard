import { useEffect, useState } from 'react';
import type { MessageInstance } from 'antd/es/message/interface';
import type { HookAPI as ModalHookAPI } from 'antd/es/modal/useModal';
import { ReadSettings, SaveSettings } from '../../wailsjs/go/main/App';
import type { SettingsModel } from '../types';
import type { ThemeMode } from '../theme';
import { resolveLanguageSetting, type SupportedLanguage } from '../i18n';
import { showOperationError } from '../components/common/operationError';
import { bytesToPartSizeMB, partSizeMBToSettingsBytes } from '../partSize';

type UseAppSettingsArgs = {
  enabled: boolean;
  messageApi: MessageInstance;
  modalApi: ModalHookAPI;
  t: (key: string) => string;
  systemLanguage: SupportedLanguage;
  setLanguage: (language: SupportedLanguage) => void;
  setThemeMode: (mode: ThemeMode) => void;
};

export function useAppSettings({
  enabled,
  messageApi,
  modalApi,
  t,
  systemLanguage,
  setLanguage,
  setThemeMode,
}: UseAppSettingsArgs) {
  const [settings, setSettings] = useState<SettingsModel | null>(null);
  const [settingsLoading, setSettingsLoading] = useState(true);
  const [settingsSaving, setSettingsSaving] = useState(false);

  const applySettings = (nextSettings: SettingsModel) => {
    const uiSettings = {
      ...nextSettings,
      defaultMaxPartSize: bytesToPartSizeMB(nextSettings.defaultMaxPartSize),
    };
    setSettings(uiSettings);
    setThemeMode((nextSettings.theme || 'system') as ThemeMode);
    setLanguage(resolveLanguageSetting(nextSettings.language, systemLanguage));
  };

  useEffect(() => {
    let cancelled = false;
    if (!enabled) {
      setSettingsLoading(false);
      return () => {
        cancelled = true;
      };
    }
    const loadSettings = async () => {
      setSettingsLoading(true);
      try {
        const nextSettings = await ReadSettings();
        if (cancelled) {
          return;
        }
        applySettings(nextSettings);
      } catch (error) {
        if (!cancelled) {
          showOperationError(modalApi, t('errorLoadingSettings'), error, t);
        }
      } finally {
        if (!cancelled) {
          setSettingsLoading(false);
        }
      }
    };
    void loadSettings();
    return () => {
      cancelled = true;
    };
  }, [enabled, messageApi, t]);

  useEffect(() => {
    if (settings?.language === 'system') {
      setLanguage(systemLanguage);
    }
  }, [settings?.language, setLanguage, systemLanguage]);

  const handleSaveSettings = async (values: SettingsModel) => {
    setSettingsSaving(true);
    try {
      const saved = await SaveSettings({
        ...values,
        defaultMaxPartSize: partSizeMBToSettingsBytes(values.defaultMaxPartSize),
      });
      applySettings(saved);
      messageApi.success(t('settingsSaved'));
    } catch (error) {
      showOperationError(modalApi, t('errorSavingSettings'), error, t);
    } finally {
      setSettingsSaving(false);
    }
  };

  return {
    settings,
    settingsLoading,
    settingsSaving,
    handleSaveSettings,
  };
}
