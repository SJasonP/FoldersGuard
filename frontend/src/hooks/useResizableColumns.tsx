import {useEffect, useMemo, useState} from 'react';
import type {ReactNode} from 'react';

type ColumnWidthMap = Record<string, number>;

type ResizableColumnInput<Row> = {
    title: ReactNode;
    key: string;
    dataIndex?: keyof Row;
    minWidth?: number;
    defaultWidth: number;
};

type UseResizableColumnsResult<Row> = {
    columnWidths: ColumnWidthMap;
    resizeTitle: (column: ResizableColumnInput<Row>) => ReactNode;
    scrollX: number;
};

export function useResizableColumns<Row>(
    storageKey: string,
    columns: ResizableColumnInput<Row>[],
): UseResizableColumnsResult<Row> {
    const defaultWidths = useMemo(
        () => Object.fromEntries(columns.map((column) => [column.key, column.defaultWidth])),
        [columns],
    );
    const minWidths = useMemo(
        () => Object.fromEntries(columns.map((column) => [column.key, column.minWidth ?? 120])),
        [columns],
    );
    const [columnWidths, setColumnWidths] = useState<ColumnWidthMap>(() => readColumnWidths(storageKey, defaultWidths, minWidths));

    useEffect(() => {
        setColumnWidths((current) => {
            const next = {...defaultWidths, ...current};
            for (const [key, minWidth] of Object.entries(minWidths)) {
                next[key] = Math.max(next[key] ?? defaultWidths[key] ?? minWidth, minWidth);
            }
            return next;
        });
    }, [defaultWidths, minWidths]);

    useEffect(() => {
        window.localStorage.setItem(storageKey, JSON.stringify(columnWidths));
    }, [columnWidths, storageKey]);

    const resizeTitle = (column: ResizableColumnInput<Row>) => (
        <span className="resizable-column-title">
            <span className="resizable-column-label">{column.title}</span>
            <span
                aria-hidden="true"
                className="resizable-column-handle"
                onClick={(event) => event.stopPropagation()}
                onMouseDown={(event) => {
                    event.preventDefault();
                    event.stopPropagation();
                    const startX = event.clientX;
                    const startWidth = columnWidths[column.key] ?? column.defaultWidth;
                    const minWidth = minWidths[column.key] ?? 120;
                    const onMouseMove = (moveEvent: MouseEvent) => {
                        const nextWidth = Math.max(minWidth, startWidth + moveEvent.clientX - startX);
                        setColumnWidths((current) => ({...current, [column.key]: nextWidth}));
                    };
                    const onMouseUp = () => {
                        window.removeEventListener('mousemove', onMouseMove);
                        window.removeEventListener('mouseup', onMouseUp);
                    };
                    window.addEventListener('mousemove', onMouseMove);
                    window.addEventListener('mouseup', onMouseUp);
                }}
            />
        </span>
    );

    return {
        columnWidths,
        resizeTitle,
        scrollX: Object.values(columnWidths).reduce((total, width) => total + width, 64),
    };
}

function readColumnWidths(storageKey: string, defaultWidths: ColumnWidthMap, minWidths: ColumnWidthMap): ColumnWidthMap {
    try {
        const raw = window.localStorage.getItem(storageKey);
        if (!raw) {
            return defaultWidths;
        }
        const parsed = JSON.parse(raw) as Record<string, unknown>;
        const widths = {...defaultWidths};
        for (const key of Object.keys(defaultWidths)) {
            const value = parsed[key];
            if (typeof value === 'number' && Number.isFinite(value)) {
                widths[key] = Math.max(value, minWidths[key] ?? 120);
            }
        }
        return widths;
    } catch {
        return defaultWidths;
    }
}
