import { Descriptions, Drawer } from 'antd';
import type { DecryptProjectResultModel } from '../../types';

type DecryptProjectDrawerProps = {
  open: boolean;
  result: DecryptProjectResultModel | null;
  onClose: () => void;
  t: (key: string) => string;
};

export function DecryptProjectDrawer({ open, result, onClose, t }: DecryptProjectDrawerProps) {
  return (
    <Drawer title={t('decryptProject')} open={open} onClose={onClose} width={540}>
      {result ? (
        <Descriptions column={1} bordered size="small">
          <Descriptions.Item label={t('projectId')}>{result.projectId}</Descriptions.Item>
          <Descriptions.Item label={t('outputPath')}>{result.outputPath}</Descriptions.Item>
          <Descriptions.Item label={t('decryptedFiles')}>{result.decryptedFiles}</Descriptions.Item>
          <Descriptions.Item label={t('restoredFolders')}>{result.restoredFolders}</Descriptions.Item>
          <Descriptions.Item label={t('skippedFolders')}>{result.skippedFolders}</Descriptions.Item>
          <Descriptions.Item label={t('deletedEncryptedFiles')}>{result.deletedEncryptedFiles}</Descriptions.Item>
          <Descriptions.Item label={t('failedEncryptedFiles')}>{result.failedEncryptedFiles}</Descriptions.Item>
        </Descriptions>
      ) : null}
    </Drawer>
  );
}
