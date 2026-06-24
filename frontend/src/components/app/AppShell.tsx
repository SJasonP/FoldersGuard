import {Layout, Menu, Typography} from 'antd';
import {HomeOutlined, InfoCircleOutlined, SettingOutlined,} from '@ant-design/icons';
import type {NavigationKey} from '../../types';
import type {OperationProgress as OperationProgressData} from '../../hooks/useOperationProgress';
import {OperationProgress} from './OperationProgress';

type AppShellProps = {
    navigation: NavigationKey;
    onNavigationChange: (navigation: NavigationKey) => void;
    activeOperationLabel: string | null;
    operationProgress: OperationProgressData | null;
    resolvedTheme: 'light' | 'dark';
    children: React.ReactNode;
    t: (key: string, options?: Record<string, unknown>) => string;
};

export function AppShell({
                             navigation,
                             onNavigationChange,
                             activeOperationLabel,
                             operationProgress,
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
                <Layout.Content className="app-content">{children}</Layout.Content>
            </Layout>
            {activeOperationLabel ? (
                <OperationProgress
                    label={activeOperationLabel}
                    progress={operationProgress}
                    resolvedTheme={resolvedTheme}
                    t={t}
                />
            ) : null}
        </Layout>
    );
}
