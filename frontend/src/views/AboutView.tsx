import { Space, Typography } from 'antd';
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';
import type { AppInfoModel } from '../types';

type AboutViewProps = {
  info: AppInfoModel | null;
  t: (key: string) => string;
};

export function AboutView({ info, t }: AboutViewProps) {
  return (
    <Space direction="vertical" size="middle" className="content-stack">
      <Typography.Title level={2}>{t('about')}</Typography.Title>
      {info && (
        <div className="about-grid">
          <Typography.Text>{t('productName')}</Typography.Text>
          <Typography.Text code>{info.productName}</Typography.Text>
          <Typography.Text>{t('productVersion')}</Typography.Text>
          <Typography.Text code>{info.productVersion}</Typography.Text>
          <Typography.Text>{t('formatVersion')}</Typography.Text>
          <Typography.Text code>{info.formatVersion}</Typography.Text>
          <Typography.Text>{t('dataDirectory')}</Typography.Text>
          <Typography.Text code>{info.dataDir}</Typography.Text>
          <Typography.Text>{t('copyrightNotice')}</Typography.Text>
          <Typography.Text code>{info.copyrightNotice}</Typography.Text>
          <Typography.Text>{t('projectLink')}</Typography.Text>
          <Typography.Link onClick={() => BrowserOpenURL(info.projectLink)}>{info.projectLink}</Typography.Link>
          <Typography.Text>{t('openSourceLicenses')}</Typography.Text>
          <Typography.Link onClick={() => BrowserOpenURL(info.thirdPartyLink)}>{t('viewOpenSourceLicenses')}</Typography.Link>
        </div>
      )}
    </Space>
  );
}
