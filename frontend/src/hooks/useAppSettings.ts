import { useEffect, useState } from 'react';
import type { MessageInstance } from 'antd/es/message/interface';
import { ClearRecentPaths, ReadSettings, SaveSettings } from '../../wailsjs/go/main/App';
import type { SettingsModel } from '../types';
import type { ThemeMode } from '../theme';

type UseAppSettingsArgs = {
  messageApi: MessageInstance;
  t: (key: string) => string;
  setLanguage: (language: 'en-US' | 'zh-CN') => void;
  setThemeMode: (mode: ThemeMode) => void;
};

export function useAppSettings({ messageApi, t, setLanguage, setThemeMode }: UseAppSettingsArgs) {
  const [settings, setSettings] = useState<SettingsModel | null>(null);
  const [settingsLoading, setSettingsLoading] = useState(true);
  const [settingsSaving, setSettingsSaving] = useState(false);

  const applySettings = (nextSettings: SettingsModel) => {
    setSettings(nextSettings);
    setThemeMode((nextSettings.theme || 'system') as ThemeMode);
    if (nextSettings.language === 'zh-CN') {
      setLanguage('zh-CN');
      return;
    }
    if (nextSettings.language === 'en-US') {
      setLanguage('en-US');
      return;
    }
    const browserLanguage = typeof navigator !== 'undefined' ? navigator.language : 'en-US';
    setLanguage(browserLanguage.toLowerCase().startsWith('zh') ? 'zh-CN' : 'en-US');
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

  const handleClearRecentPaths = async () => {
    setSettingsSaving(true);
    try {
      const cleared = await ClearRecentPaths();
      applySettings(cleared);
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
    handleClearRecentPaths,
  };
}
