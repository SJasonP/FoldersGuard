import {Layout, Menu, Progress, Space, Typography} from 'antd';
import {HomeOutlined, InfoCircleOutlined, SettingOutlined,} from '@ant-design/icons';
import type {NavigationKey} from '../../types';

type AppShellProps = {
    navigation: NavigationKey;
    onNavigationChange: (navigation: NavigationKey) => void;
    activeOperationLabel: string | null;
    resolvedTheme: 'light' | 'dark';
    children: React.ReactNode;
    t: (key: string) => string;
};

export function AppShell({
                             navigation,
                             onNavigationChange,
                             activeOperationLabel,
                             resolvedTheme,
                             children,
                             t,
                         }: AppShellProps) {
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
                    onClick={({key}) => onNavigationChange(key as NavigationKey)}
                    items={[
                        {key: 'home', icon: <HomeOutlined/>, label: t('home')},
                        {key: 'settings', icon: <SettingOutlined/>, label: t('settings')},
                        {key: 'about', icon: <InfoCircleOutlined/>, label: t('about')},
                    ]}
                />
            </Layout.Sider>
            <Layout>
                {activeOperationLabel ? (
                    <Layout.Header className="app-header">
                        <Space className="operation-status" size="small">
                            <Progress className="operation-progress" percent={100} size="small" status="active"
                                      showInfo={false}/>
                            <Typography.Text>
                                {t('operationRunning')}: {activeOperationLabel}
                            </Typography.Text>
                        </Space>
                    </Layout.Header>
                ) : null}
                <Layout.Content className="app-content">{children}</Layout.Content>
            </Layout>
        </Layout>
    );
}
