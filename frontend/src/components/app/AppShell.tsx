import { Button, Layout, Menu, Progress, Space, Typography } from 'antd';
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
  activeOperationLabel: string | null;
  actionsDisabled: boolean;
  resolvedTheme: 'light' | 'dark';
  children: React.ReactNode;
  t: (key: string) => string;
};

export function AppShell({
  navigation,
  onNavigationChange,
  onCreateProject,
  onImportProject,
  onLoadShare,
  activeOperationLabel,
  actionsDisabled,
  resolvedTheme,
  children,
  t,
}: AppShellProps) {
  const operationActive = activeOperationLabel !== null;
  const disabled = actionsDisabled || operationActive;

  return (
    <Layout className={`app-shell app-shell-${resolvedTheme}`}>
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
            <Button icon={<FolderAddOutlined />} type="primary" onClick={onCreateProject} disabled={disabled}>
              {t('createProject')}
            </Button>
            <Button icon={<ImportOutlined />} onClick={onImportProject} disabled={disabled}>
              {t('importProject')}
            </Button>
            <Button icon={<ShareAltOutlined />} onClick={onLoadShare} disabled={disabled}>
              {t('loadShare')}
            </Button>
          </Space>
          {activeOperationLabel ? (
            <Space className="operation-status" size="small">
              <Progress className="operation-progress" percent={100} size="small" status="active" showInfo={false} />
              <Typography.Text>
                {t('operationRunning')}: {activeOperationLabel}
              </Typography.Text>
            </Space>
          ) : null}
        </Layout.Header>
        <Layout.Content className="app-content">{children}</Layout.Content>
      </Layout>
    </Layout>
  );
}
