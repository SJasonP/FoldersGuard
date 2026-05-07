import { useEffect, useState } from 'react';
import type { MessageInstance } from 'antd/es/message/interface';
import { ReadSettings, SaveSettings } from '../../wailsjs/go/main/App';
import type { SettingsModel } from '../types';
import type { ThemeMode } from '../theme';
import { resolveLanguageSetting, type SupportedLanguage } from '../i18n';

type UseAppSettingsArgs = {
  messageApi: MessageInstance;
  t: (key: string) => string;
  systemLanguage: SupportedLanguage;
  setLanguage: (language: SupportedLanguage) => void;
  setThemeMode: (mode: ThemeMode) => void;
};

export function useAppSettings({ messageApi, t, systemLanguage, setLanguage, setThemeMode }: UseAppSettingsArgs) {
  const [settings, setSettings] = useState<SettingsModel | null>(null);
  const [settingsLoading, setSettingsLoading] = useState(true);
  const [settingsSaving, setSettingsSaving] = useState(false);

  const applySettings = (nextSettings: SettingsModel) => {
    setSettings(nextSettings);
    setThemeMode((nextSettings.theme || 'system') as ThemeMode);
    setLanguage(resolveLanguageSetting(nextSettings.language, systemLanguage));
  };

  useEffect(() => {
    let cancelled = false;
    const loadSettings = async () => {
      setSettingsLoading(true);
      try {
        const nextSettings = await ReadSettings();
        if (cancelled) {
          return;
        }
        applySettings(nextSettings);
      } catch {
        if (!cancelled) {
          messageApi.error(t('errorLoadingSettings'));
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
  }, [messageApi, t]);

  useEffect(() => {
    if (settings?.language === 'system') {
      setLanguage(systemLanguage);
    }
  }, [settings?.language, setLanguage, systemLanguage]);

  const handleSaveSettings = async (values: SettingsModel) => {
    setSettingsSaving(true);
    try {
      const saved = await SaveSettings(values);
      applySettings(saved);
      messageApi.success(t('settingsSaved'));
    } catch {
      messageApi.error(t('errorSavingSettings'));
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
