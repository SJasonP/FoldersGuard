import { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  App as AntApp,
  Button,
  ConfigProvider,
  Empty,
  Flex,
  Input,
  Layout,
  Menu,
  Space,
  Table,
  Typography,
  theme,
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
import { AppInfo } from '../wailsjs/go/main/App';
import type { SupportedLanguage } from './i18n';
import i18n from './i18n';
import { resolveTheme, themeAlgorithm, type ThemeMode } from './theme';

type NavigationKey = 'home' | 'settings' | 'about';

type LocalProject = {
  key: string;
  projectId: string;
  fileName: string;
  modifiedTime: string;
  availabilityStatus: string;
};

type AppInfoModel = Awaited<ReturnType<typeof AppInfo>>;

const antLocales: Record<SupportedLanguage, typeof enUS> = {
  'en-US': enUS,
  'zh-CN': zhCN,
};

function App() {
  const { t } = useTranslation();
  const [navigation, setNavigation] = useState<NavigationKey>('home');
  const [language, setLanguage] = useState<SupportedLanguage>('en-US');
  const [themeMode] = useState<ThemeMode>('system');
  const [resolvedTheme, setResolvedTheme] = useState(resolveTheme(themeMode));
  const [info, setInfo] = useState<AppInfoModel | null>(null);

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

  const projects = useMemo<LocalProject[]>(() => [], []);
  const columns = useMemo<ColumnsType<LocalProject>>(
    () => [
      { title: t('projectId'), dataIndex: 'projectId', key: 'projectId' },
      { title: t('projectName'), dataIndex: 'fileName', key: 'fileName' },
      { title: t('modifiedTime'), dataIndex: 'modifiedTime', key: 'modifiedTime' },
      { title: t('availabilityStatus'), dataIndex: 'availabilityStatus', key: 'availabilityStatus' },
    ],
    [t],
  );

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
                <Space direction="vertical" size="large" className="content-stack">
                  <Flex justify="space-between" align="center" gap={16}>
                    <Typography.Title level={2}>{t('localProjects')}</Typography.Title>
                    <Space>
                      <Input.Search placeholder={t('searchProjects')} />
                      <Button icon={<ReloadOutlined />}>{t('refresh')}</Button>
                    </Space>
                  </Flex>
                  <Table
                    columns={columns}
                    dataSource={projects}
                    locale={{ emptyText: <Empty description={t('noProjects')} /> }}
                    pagination={false}
                  />
                </Space>
              )}
              {navigation === 'settings' && (
                <Space direction="vertical" size="middle" className="content-stack">
                  <Typography.Title level={2}>{t('settings')}</Typography.Title>
                  <Typography.Text type="secondary">{t('startSubtitle')}</Typography.Text>
                </Space>
              )}
              {navigation === 'about' && (
                <Space direction="vertical" size="middle" className="content-stack">
                  <Typography.Title level={2}>{t('about')}</Typography.Title>
                  {info && (
                    <div className="about-grid">
                      <Typography.Text>{t('appId')}</Typography.Text>
                      <Typography.Text code>{info.appId}</Typography.Text>
                      <Typography.Text>{t('formatVersion')}</Typography.Text>
                      <Typography.Text code>{info.nativeFormatVersion}</Typography.Text>
                      <Typography.Text>{t('schemaVersion')}</Typography.Text>
                      <Typography.Text code>{info.schemaVersion}</Typography.Text>
                      <Typography.Text>{t('dataDirectory')}</Typography.Text>
                      <Typography.Text code>{info.dataDir}</Typography.Text>
                      <Typography.Text>{t('cliExecutable')}</Typography.Text>
                      <Typography.Text code>{info.cliExecutableName}</Typography.Text>
                      <Typography.Text>{t('cliAlias')}</Typography.Text>
                      <Typography.Text code>{info.cliShortAlias}</Typography.Text>
                    </div>
                  )}
                </Space>
              )}
            </Layout.Content>
          </Layout>
        </Layout>
      </AntApp>
    </ConfigProvider>
  );
}

export default App;
