import {useEffect, useRef, useState} from 'react';
import {EventsOn} from '../../wailsjs/runtime/runtime';

export type OperationProgress = {
    operationId: string;
    operation: string;
    state: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';
    phase: string;
    phaseIndex: number;
    phaseCount: number;
    determinate: boolean;
    processedBytes: number;
    totalBytes: number;
    processedItems: number;
    totalItems: number;
    currentItem: string;
    bytesPerSecond: number;
    etaSeconds: number;
    error: string;
};

const OPERATION_PROGRESS_EVENT = 'operation:progress';
const terminalStates = new Set(['completed', 'failed', 'cancelled']);

/**
 * useOperationProgress subscribes to backend operation progress events and
 * exposes the latest snapshot for the active operation. FoldersGuard runs one
 * long-running operation at a time, so the most recent non-terminal event
 * describes the current operation; terminal events clear the snapshot.
 *
 * Operations cannot be cancelled, so this hook is display-only.
 */
export function useOperationProgress() {
    const [progress, setProgress] = useState<OperationProgress | null>(null);
    const activeOperationId = useRef<string | null>(null);

    useEffect(() => {
        const unsubscribe = EventsOn(OPERATION_PROGRESS_EVENT, (event: OperationProgress) => {
            if (!event || typeof event.operationId !== 'string') {
                return;
            }
            if (terminalStates.has(event.state)) {
                // Ignore terminal events from a superseded operation.
                if (activeOperationId.current && event.operationId !== activeOperationId.current) {
                    return;
                }
                activeOperationId.current = null;
                setProgress(null);
                return;
            }
            activeOperationId.current = event.operationId;
            setProgress(event);
        });
        return () => {
            unsubscribe();
        };
    }, []);

    return {progress};
}
