import {List, Typography} from 'antd';
import type {FailedItemModel} from '../../types';

type FailuresListProps = {
    failures?: FailedItemModel[] | null;
    t: (key: string) => string;
};

// FailuresList renders the per-item failures reported by a continue-on-error
// operation. It renders nothing when there are no failures. Only the non-secret
// visible file id, base name, and reason are shown.
export function FailuresList({failures, t}: FailuresListProps) {
    if (!failures || failures.length === 0) {
        return null;
    }
    return (
        <div style={{marginTop: 16}}>
            <Typography.Title level={5}>{t('failuresListTitle')}</Typography.Title>
            <List
                size="small"
                bordered
                dataSource={failures}
                renderItem={(item) => (
                    <List.Item>
                        <List.Item.Meta
                            title={item.name ? item.name : item.fileId}
                            description={item.reason}
                        />
                    </List.Item>
                )}
            />
        </div>
    );
}
