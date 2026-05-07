import { Collapse, Space, Typography } from 'antd';
import type { HookAPI as ModalHookAPI } from 'antd/es/modal/useModal';

const secretPatterns = [
  /(password\s*[:=]\s*)([^\s,;]+)/gi,
  /(passphrase\s*[:=]\s*)([^\s,;]+)/gi,
  /(database[_\s-]*key\s*[:=]\s*)([^\s,;]+)/gi,
  /(key[_\s-]*material\s*[:=]\s*)([^\s,;]+)/gi,
];

export function technicalErrorMessage(error: unknown) {
  const message = error instanceof Error ? error.message : String(error ?? '');
  return secretPatterns.reduce((current, pattern) => current.replace(pattern, '$1[redacted]'), message).trim();
}

export function showOperationError(
  modalApi: ModalHookAPI,
  title: string,
  error: unknown,
  t: (key: string) => string,
) {
  const details = technicalErrorMessage(error);
  modalApi.error({
    title,
    content: details ? (
      <Collapse
        ghost
        items={[
          {
            key: 'technical-details',
            label: t('technicalDetails'),
            children: <Typography.Text code>{details}</Typography.Text>,
          },
        ]}
      />
    ) : undefined,
  });
}

export function showStartupError(
  modalApi: ModalHookAPI,
  title: string,
  dataDirectory: string,
  error: unknown,
  t: (key: string) => string,
) {
  const details = technicalErrorMessage(error);
  modalApi.error({
    title,
    closable: false,
    maskClosable: false,
    content: (
      <Space direction="vertical" size="middle">
        <Space direction="vertical" size={4}>
          <Typography.Text>{t('dataDirectory')}</Typography.Text>
          <Typography.Text code>{dataDirectory}</Typography.Text>
        </Space>
        {details ? (
          <Collapse
            ghost
            items={[
              {
                key: 'technical-details',
                label: t('underlyingError'),
                children: <Typography.Text code>{details}</Typography.Text>,
              },
            ]}
          />
        ) : null}
      </Space>
    ),
  });
}
