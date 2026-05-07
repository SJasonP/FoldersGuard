import { Descriptions, Drawer } from 'antd';
import type { DecryptShareResultModel } from '../../types';
import { formatNumber } from '../../formatters';

type DecryptShareDrawerProps = {
  open: boolean;
  result: DecryptShareResultModel | null;
  onClose: () => void;
  t: (key: string) => string;
};

export function DecryptShareDrawer({ open, result, onClose, t }: DecryptShareDrawerProps) {
  return (
    <Drawer title={t('decryptShare')} open={open} onClose={onClose} width={540}>
      {result ? (
        <Descriptions column={1} bordered size="small">
          <Descriptions.Item label={t('shareId')}>{result.shareId}</Descriptions.Item>
          <Descriptions.Item label={t('outputPath')}>{result.outputPath}</Descriptions.Item>
          <Descriptions.Item label={t('decryptedFiles')}>{formatNumber(result.decryptedFiles)}</Descriptions.Item>
          <Descriptions.Item label={t('restoredFolders')}>{formatNumber(result.restoredFolders)}</Descriptions.Item>
          <Descriptions.Item label={t('deletedEncryptedFiles')}>{formatNumber(result.deletedEncryptedFiles)}</Descriptions.Item>
          <Descriptions.Item label={t('failedEncryptedFiles')}>{formatNumber(result.failedEncryptedFiles)}</Descriptions.Item>
        </Descriptions>
      ) : null}
    </Drawer>
  );
}
