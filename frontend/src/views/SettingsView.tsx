import {Button, Form, InputNumber, Select, Space, Typography} from 'antd';
import type {SettingsModel} from '../types';

type SettingsViewProps = {
    settings: SettingsModel | null;
    loading: boolean;
    saving: boolean;
    disabled: boolean;
    onSave: (values: SettingsModel) => void;
    t: (key: string) => string;
};

export function SettingsView({
                                 settings,
                                 loading,
                                 saving,
                                 disabled,
                                 onSave,
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
                disabled={disabled || loading}
                onFinish={onSave}
            >
                <div className="settings-grid">
                    <Form.Item name="operationGuideFormat" label={t('operationGuideFormat')} rules={[{required: true}]}>
                        <Select
                            options={[
                                {value: 'txt', label: 'txt'},
                                {value: 'md', label: 'md'},
                            ]}
                        />
                    </Form.Item>
                    <Form.Item name="defaultMaxPartSize" label={t('defaultMaxPartSize')}>
                        <InputNumber min={0} precision={0} style={{width: '100%'}}
                                     placeholder={t('partSizeDisabledHint')}/>
                    </Form.Item>
                    <Form.Item name="theme" label={t('theme')} rules={[{required: true}]}>
                        <Select
                            options={[
                                {value: 'system', label: t('themeSystem')},
                                {value: 'light', label: t('themeLight')},
                                {value: 'dark', label: t('themeDark')},
                            ]}
                        />
                    </Form.Item>
                    <Form.Item name="language" label={t('language')} rules={[{required: true}]}>
                        <Select
                            options={[
                                {value: 'system', label: t('languageSystem')},
                                {value: 'en-US', label: t('languageEnglishUS')},
                                {value: 'zh-CN', label: t('languageSimplifiedChinese')},
                            ]}
                        />
                    </Form.Item>
                </div>
                <Space wrap>
                    <Button type="primary" htmlType="submit" loading={saving} disabled={disabled}>
                        {t('saveSettings')}
                    </Button>
                </Space>
            </Form>
        </Space>
    );
}
