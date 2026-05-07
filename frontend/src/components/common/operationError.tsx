import { Collapse, Typography } from 'antd';
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
