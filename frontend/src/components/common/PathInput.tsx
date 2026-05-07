import { FolderOpenOutlined } from '@ant-design/icons';
import { Button, Input, Space } from 'antd';
import { SelectPath } from '../../../wailsjs/go/main/App';
import { main } from '../../../wailsjs/go/models';

type PathDialogKind = 'open-directory' | 'open-file' | 'save-file';

type PathDialogFilter = {
  displayName: string;
  pattern: string;
};

type PathInputProps = {
  value?: string;
  onChange?: (value: string) => void;
  placeholder?: string;
  dialogKind: PathDialogKind;
  dialogTitle: string;
  buttonLabel?: string;
  secondaryDialogKind?: PathDialogKind;
  secondaryDialogTitle?: string;
  secondaryButtonLabel?: string;
  defaultFilename?: string;
  filters?: PathDialogFilter[];
  t: (key: string) => string;
};

export function PathInput({
  value,
  onChange,
  placeholder,
  dialogKind,
  dialogTitle,
  buttonLabel,
  secondaryDialogKind,
  secondaryDialogTitle,
  secondaryButtonLabel,
  defaultFilename,
  filters,
  t,
}: PathInputProps) {
  const openDialog = async (kind: PathDialogKind, title: string) => {
    const selected = await SelectPath(new main.SelectPathRequest({
      kind,
      title,
      defaultFilename: defaultFilename ?? '',
      filters: filters ?? [],
    }));
    if (selected) {
      onChange?.(selected);
    }
  };

  return (
    <Space.Compact style={{ width: '100%' }}>
      <Input value={value} onChange={(event) => onChange?.(event.target.value)} placeholder={placeholder} />
      <Button icon={<FolderOpenOutlined />} onClick={() => void openDialog(dialogKind, dialogTitle)}>
        {buttonLabel ?? t('browse')}
      </Button>
      {secondaryDialogKind && secondaryDialogTitle ? (
        <Button icon={<FolderOpenOutlined />} onClick={() => void openDialog(secondaryDialogKind, secondaryDialogTitle)}>
          {secondaryButtonLabel ?? t('browse')}
        </Button>
      ) : null}
    </Space.Compact>
  );
}
