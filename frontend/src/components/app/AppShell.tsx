import { Button, Layout, Menu, Space, Typography } from 'antd';
import {
  FolderAddOutlined,
  HomeOutlined,
  ImportOutlined,
  InfoCircleOutlined,
  SettingOutlined,
  ShareAltOutlined,
} from '@ant-design/icons';
import type { NavigationKey } from '../../types';

type AppShellProps = {
  navigation: NavigationKey;
  onNavigationChange: (navigation: NavigationKey) => void;
  onCreateProject: () => void;
  onImportProject: () => void;
  onLoadShare: () => void;
  language: 'en-US' | 'zh-CN';
  onToggleLanguage: () => void;
  children: React.ReactNode;
  t: (key: string) => string;
};

export function AppShell({
  navigation,
  onNavigationChange,
  onCreateProject,
  onImportProject,
  onLoadShare,
  language,
  onToggleLanguage,
  children,
  t,
}: AppShellProps) {
  return (
    <Layout className="app-shell">
      <Layout.Sider width={236} className="app-sidebar">
        <div className="app-brand">
          <Typography.Title level={4}>{t('foldersGuard')}</Typography.Title>
          <Typography.Text type="secondary">{t('startSubtitle')}</Typography.Text>
        </div>
        <Menu
          mode="inline"
          selectedKeys={[navigation]}
          onClick={({ key }) => onNavigationChange(key as NavigationKey)}
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
            <Button icon={<FolderAddOutlined />} type="primary" onClick={onCreateProject}>
              {t('createProject')}
            </Button>
            <Button icon={<ImportOutlined />} onClick={onImportProject}>
              {t('importProject')}
            </Button>
            <Button icon={<ShareAltOutlined />} onClick={onLoadShare}>
              {t('loadShare')}
            </Button>
          </Space>
          <Space>
            <Button onClick={onToggleLanguage}>{language}</Button>
          </Space>
        </Layout.Header>
        <Layout.Content className="app-content">{children}</Layout.Content>
      </Layout>
    </Layout>
  );
}
