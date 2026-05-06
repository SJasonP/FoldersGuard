import { Button, Form, InputNumber, Select, Space, Switch, Typography } from 'antd';
import type { SettingsModel } from '../types';

type SettingsViewProps = {
  settings: SettingsModel | null;
  loading: boolean;
  saving: boolean;
  onSave: (values: SettingsModel) => void;
  onClearRecentPaths: () => void;
  t: (key: string) => string;
};

export function SettingsView({
  settings,
  loading,
  saving,
  onSave,
  onClearRecentPaths,
  t,
}: SettingsViewProps) {
  const [form] = Form.useForm<SettingsModel>();

  if (settings && form.getFieldValue('language') === undefined) {
    form.setFieldsValue(settings);
  }

  return (
    <Space direction="vertical" size="large" className="content-stack">
      <Typography.Title level={2}>{t('settings')}</Typography.Title>
      <Form<SettingsModel>
        form={form}
        layout="vertical"
        initialValues={settings ?? undefined}
        disabled={loading}
        onFinish={onSave}
      >
        <div className="settings-grid">
          <Form.Item name="operationGuideFormat" label={t('operationGuideFormat')} rules={[{ required: true }]}>
            <Select
              options={[
                { value: 'txt', label: 'txt' },
                { value: 'md', label: 'md' },
              ]}
            />
          </Form.Item>
          <Form.Item name="defaultMaxPartSize" label={t('defaultMaxPartSize')}>
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="sourceCleanupMode" label={t('sourceCleanupMode')} rules={[{ required: true }]}>
            <Select
              options={[
                { value: 'ask', label: t('sourceCleanupAsk') },
                { value: 'keep', label: t('sourceCleanupKeep') },
                { value: 'delete', label: t('sourceCleanupDelete') },
              ]}
            />
          </Form.Item>
          <Form.Item name="theme" label={t('theme')} rules={[{ required: true }]}>
            <Select
              options={[
                { value: 'system', label: t('themeSystem') },
                { value: 'light', label: t('themeLight') },
                { value: 'dark', label: t('themeDark') },
              ]}
            />
          </Form.Item>
          <Form.Item name="language" label={t('language')} rules={[{ required: true }]}>
            <Select
              options={[
                { value: 'system', label: t('languageSystem') },
                { value: 'en-US', label: t('languageEnglishUS') },
                { value: 'zh-CN', label: t('languageSimplifiedChinese') },
              ]}
            />
          </Form.Item>
        </div>
        <div className="settings-switches">
          <Form.Item name="rememberRecentPaths" label={t('rememberRecentPaths')} valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name="windowStatePersistence" label={t('windowStatePersistence')} valuePropName="checked">
            <Switch />
          </Form.Item>
        </div>
        <Space wrap>
          <Button type="primary" htmlType="submit" loading={saving}>
            {t('saveSettings')}
          </Button>
          <Button onClick={onClearRecentPaths} disabled={loading || saving}>
            {t('clearRecentPaths')}
          </Button>
        </Space>
      </Form>
      {settings ? (
        <Typography.Text type="secondary">
          {t('recentPathCount')}: {settings.recentPaths.length}
        </Typography.Text>
      ) : null}
    </Space>
  );
}
