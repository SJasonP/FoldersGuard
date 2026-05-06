import { Descriptions, Drawer } from 'antd';
import type { VerifyProjectResultModel } from '../../types';

type VerifyProjectDrawerProps = {
  open: boolean;
  result: VerifyProjectResultModel | null;
  onClose: () => void;
  t: (key: string) => string;
};

export function VerifyProjectDrawer({ open, result, onClose, t }: VerifyProjectDrawerProps) {
  return (
    <Drawer title={t('verifyProject')} open={open} onClose={onClose} width={540}>
      {result ? (
        <Descriptions column={1} bordered size="small">
          <Descriptions.Item label={t('projectId')}>{result.projectId}</Descriptions.Item>
          <Descriptions.Item label={t('verifyProjectStatus')}>
            {result.status === 'ok' ? t('verificationOk') : t('verificationFailed')}
          </Descriptions.Item>
          <Descriptions.Item label={t('checkedObjects')}>{result.checkedObjects}</Descriptions.Item>
          <Descriptions.Item label={t('missingObjects')}>{result.missingObjects}</Descriptions.Item>
          <Descriptions.Item label={t('tamperedObjects')}>{result.tamperedObjects}</Descriptions.Item>
          <Descriptions.Item label={t('extraObjects')}>{result.extraObjects}</Descriptions.Item>
        </Descriptions>
      ) : null}
    </Drawer>
  );
}
