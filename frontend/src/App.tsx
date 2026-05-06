import { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  App as AntApp,
  Button,
  ConfigProvider,
  Descriptions,
  Drawer,
  Form,
  Input,
  Layout,
  Menu,
  Modal,
  Space,
  Typography,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import {
  FolderAddOutlined,
  HomeOutlined,
  ImportOutlined,
  InfoCircleOutlined,
  ReloadOutlined,
  SettingOutlined,
  ShareAltOutlined,
} from '@ant-design/icons';
import enUS from 'antd/locale/en_US';
import zhCN from 'antd/locale/zh_CN';
import { AppInfo, ClearRecentPaths, InspectProject, ListLocalProjects, ReadSettings, SaveSettings } from '../wailsjs/go/main/App';
import type { SupportedLanguage } from './i18n';
import i18n from './i18n';
import { resolveTheme, themeAlgorithm, type ThemeMode } from './theme';
import type {
  AppInfoModel,
  InspectProjectResultModel,
  LocalProjectRow,
  LocalProjectSummary,
  NavigationKey,
  SettingsModel,
} from './types';
import { HomeView } from './views/HomeView';
import { SettingsView } from './views/SettingsView';
import { AboutView } from './views/AboutView';

const antLocales: Record<SupportedLanguage, typeof enUS> = {
  'en-US': enUS,
  'zh-CN': zhCN,
};

function App() {
  const { t } = useTranslation();
  const antApp = AntApp.useApp();
  const [navigation, setNavigation] = useState<NavigationKey>('home');
  const [language, setLanguage] = useState<'en-US' | 'zh-CN'>('en-US');
  const [themeMode, setThemeMode] = useState<ThemeMode>('system');
  const [resolvedTheme, setResolvedTheme] = useState(resolveTheme(themeMode));
  const [info, setInfo] = useState<AppInfoModel | null>(null);
  const [projectSearch, setProjectSearch] = useState('');
  const [projects, setProjects] = useState<LocalProjectSummary[]>([]);
  const [projectsLoading, setProjectsLoading] = useState(true);
  const [projectsError, setProjectsError] = useState<string | null>(null);
  const [selectedProjectId, setSelectedProjectId] = useState<string | null>(null);
  const [settings, setSettings] = useState<SettingsModel | null>(null);
  const [settingsLoading, setSettingsLoading] = useState(true);
  const [settingsSaving, setSettingsSaving] = useState(false);
  const [projectActionsOpen, setProjectActionsOpen] = useState(false);
  const [inspectDialogOpen, setInspectDialogOpen] = useState(false);
  const [inspectLoading, setInspectLoading] = useState(false);
  const [inspectResult, setInspectResult] = useState<InspectProjectResultModel | null>(null);
  const [inspectResultOpen, setInspectResultOpen] = useState(false);
  const [inspectForm] = Form.useForm<{ password: string }>();

  useEffect(() => {
    AppInfo().then(setInfo).catch(() => setInfo(null));
  }, []);

  useEffect(() => {
    void i18n.changeLanguage(language);
  }, [language]);

  useEffect(() => {
    const media = window.matchMedia('(prefers-color-scheme: dark)');
    const update = () => setResolvedTheme(resolveTheme(themeMode));
    update();
    media.addEventListener('change', update);
    return () => media.removeEventListener('change', update);
  }, [themeMode]);

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

  const loadProjects = async () => {
    setProjectsLoading(true);
    setProjectsError(null);
    try {
      const nextProjects = await ListLocalProjects();
      setProjects(nextProjects);
    } catch {
      setProjects([]);
      setProjectsError(t('errorLoadingProjects'));
    } finally {
      setProjectsLoading(false);
    }
  };

  useEffect(() => {
    void loadProjects();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

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
          antApp.message.error(t('errorLoadingSettings'));
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
  }, [antApp.message, t]);

  const handleSaveSettings = async (values: SettingsModel) => {
    setSettingsSaving(true);
    try {
      const saved = await SaveSettings(values);
      applySettings(saved);
      antApp.message.success(t('settingsSaved'));
    } catch {
      antApp.message.error(t('errorSavingSettings'));
    } finally {
      setSettingsSaving(false);
    }
  };

  const handleClearRecentPaths = async () => {
    setSettingsSaving(true);
    try {
      const cleared = await ClearRecentPaths();
      applySettings(cleared);
      antApp.message.success(t('settingsSaved'));
    } catch {
      antApp.message.error(t('errorSavingSettings'));
    } finally {
      setSettingsSaving(false);
    }
  };

  const visibleProjects = useMemo<LocalProjectRow[]>(
    () =>
      projects
        .filter((project) => {
          const query = projectSearch.trim().toLowerCase();
          if (query === '') {
            return true;
          }
          return (
            project.projectId.toLowerCase().includes(query) ||
            project.fileName.toLowerCase().includes(query) ||
            project.availabilityStatus.toLowerCase().includes(query)
          );
        })
        .map((project) => ({
          key: project.projectId,
          projectId: project.projectId,
          fileName: project.fileName,
          modifiedTime: project.modifiedAt
            ? new Intl.DateTimeFormat(language, {
                dateStyle: 'medium',
                timeStyle: 'short',
              }).format(new Date(project.modifiedAt))
            : '',
          availabilityStatus: t(project.availabilityStatus),
        })),
    [language, projectSearch, projects, t],
  );

  const selectedProject = useMemo(
    () => projects.find((project) => project.projectId === selectedProjectId) ?? null,
    [projects, selectedProjectId],
  );

  const columns = useMemo<ColumnsType<LocalProjectRow>>(
    () => [
      { title: t('projectId'), dataIndex: 'projectId', key: 'projectId' },
      { title: t('projectName'), dataIndex: 'fileName', key: 'fileName' },
      { title: t('modifiedTime'), dataIndex: 'modifiedTime', key: 'modifiedTime' },
      { title: t('availabilityStatus'), dataIndex: 'availabilityStatus', key: 'availabilityStatus' },
    ],
    [t],
  );

  const handleOpenProjectActions = () => {
    if (!selectedProjectId) {
      return;
    }
    setProjectActionsOpen(true);
  };

  const handleInspectProject = async (values: { password: string }) => {
    if (!selectedProjectId) {
      return;
    }
    setInspectLoading(true);
    try {
      const result = await InspectProject({
        projectId: selectedProjectId,
        password: values.password,
      });
      setInspectDialogOpen(false);
      inspectForm.resetFields();
      setProjectActionsOpen(false);
      setInspectResult(result);
      setInspectResultOpen(true);
    } catch {
      antApp.message.error(t('inspectProjectFailed'));
    } finally {
      setInspectLoading(false);
    }
  };

  return (
    <ConfigProvider
      locale={antLocales[language]}
      theme={{
        algorithm: themeAlgorithm(resolvedTheme),
        token: {
          borderRadius: 6,
          colorPrimary: '#1677ff',
        },
      }}
    >
      <AntApp>
        <Layout className="app-shell">
          <Layout.Sider width={236} className="app-sidebar">
            <div className="app-brand">
              <Typography.Title level={4}>{t('foldersGuard')}</Typography.Title>
              <Typography.Text type="secondary">{t('startSubtitle')}</Typography.Text>
            </div>
            <Menu
              mode="inline"
              selectedKeys={[navigation]}
              onClick={({ key }) => setNavigation(key as NavigationKey)}
              items={[
                { key: 'home', icon: <HomeOutlined />, label: t('home') },
                { key: 'settings', icon: <SettingOutlined />, label: t('settings') },
                { key: 'about', icon: <InfoCircleOutlined />, label: t('about') },
              ]}
            />
          </Layout.Sider>
          <Layout>
            <Layout.Header className="app-header">
              <Space>
                <Button icon={<FolderAddOutlined />} type="primary">
                  {t('createProject')}
                </Button>
                <Button icon={<ImportOutlined />}>{t('importProject')}</Button>
                <Button icon={<ShareAltOutlined />}>{t('loadShare')}</Button>
              </Space>
              <Space>
                <Button onClick={() => setLanguage(language === 'en-US' ? 'zh-CN' : 'en-US')}>
                  {language}
                </Button>
              </Space>
            </Layout.Header>
            <Layout.Content className="app-content">
              {navigation === 'home' && (
                <HomeView
                  columns={columns}
                  loading={projectsLoading}
                  projects={visibleProjects}
                  projectSearch={projectSearch}
                  projectsError={projectsError}
                  selectedProjectId={selectedProjectId}
                  onProjectSearchChange={setProjectSearch}
                  onRefresh={() => void loadProjects()}
                  onSelectProject={setSelectedProjectId}
                  onOpenProjectActions={handleOpenProjectActions}
                  t={t}
                />
              )}
              {navigation === 'settings' && (
                <SettingsView
                  settings={settings}
                  loading={settingsLoading}
                  saving={settingsSaving}
                  onSave={(values) => void handleSaveSettings(values)}
                  onClearRecentPaths={() => void handleClearRecentPaths()}
                  t={t}
                />
              )}
              {navigation === 'about' && <AboutView info={info} t={t} />}
            </Layout.Content>
          </Layout>
        </Layout>
        <Drawer
          title={t('projectActions')}
          open={projectActionsOpen}
          onClose={() => setProjectActionsOpen(false)}
          width={360}
        >
          <Space direction="vertical" size="middle" className="content-stack">
            {selectedProject ? (
              <Typography.Text type="secondary">
                {selectedProject.projectId} / {selectedProject.fileName}
              </Typography.Text>
            ) : null}
            <Button
              block
              type="primary"
              onClick={() => {
                setInspectDialogOpen(true);
              }}
            >
              {t('inspectProject')}
            </Button>
          </Space>
        </Drawer>
        <Modal
          title={t('inspectProject')}
          open={inspectDialogOpen}
          onCancel={() => {
            setInspectDialogOpen(false);
            inspectForm.resetFields();
          }}
          onOk={() => void inspectForm.submit()}
          okText={t('inspectProject')}
          confirmLoading={inspectLoading}
        >
          <Form form={inspectForm} layout="vertical" onFinish={(values) => void handleInspectProject(values)}>
            <Form.Item name="password" label={t('password')} rules={[{ required: true, message: t('passwordRequired') }]}>
              <Input.Password autoComplete="current-password" />
            </Form.Item>
          </Form>
        </Modal>
        <Drawer
          title={t('projectDetails')}
          open={inspectResultOpen}
          onClose={() => setInspectResultOpen(false)}
          width={540}
        >
          {inspectResult ? (
            <Descriptions column={1} bordered size="small">
              <Descriptions.Item label={t('projectId')}>{inspectResult.projectId}</Descriptions.Item>
              <Descriptions.Item label={t('projectName')}>{inspectResult.rootName}</Descriptions.Item>
              <Descriptions.Item label={t('rootFolderId')}>{inspectResult.rootFolderId}</Descriptions.Item>
              <Descriptions.Item label={t('rootName')}>{inspectResult.rootName}</Descriptions.Item>
              <Descriptions.Item label={t('formatVersion')}>{inspectResult.formatVersion}</Descriptions.Item>
              <Descriptions.Item label={t('schemaVersion')}>{inspectResult.schemaVersion}</Descriptions.Item>
              <Descriptions.Item label={t('databaseType')}>{inspectResult.databaseType}</Descriptions.Item>
              <Descriptions.Item label={t('itemCount')}>{inspectResult.items}</Descriptions.Item>
              <Descriptions.Item label={t('folderCount')}>{inspectResult.folders}</Descriptions.Item>
              <Descriptions.Item label={t('fileCount')}>{inspectResult.files}</Descriptions.Item>
              <Descriptions.Item label={t('partCount')}>{inspectResult.parts}</Descriptions.Item>
              <Descriptions.Item label={t('storageObjects')}>{inspectResult.storageObjects}</Descriptions.Item>
            </Descriptions>
          ) : null}
        </Drawer>
      </AntApp>
    </ConfigProvider>
  );
}

export default App;
