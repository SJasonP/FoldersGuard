const bytesPerMB = 1024 * 1024;
const minimumSplitPartSizeMB = 5;

export function bytesToPartSizeMB(bytes: number | null | undefined) {
    const value = Math.trunc(bytes ?? 0);
    if (value <= 0) {
        return 0;
    }
    return Math.max(0, Math.trunc(value / bytesPerMB));
}

export function partSizeMBToSettingsBytes(mb: number | null | undefined) {
    const value = Math.trunc(mb ?? 0);
    if (value < minimumSplitPartSizeMB) {
        return 0;
    }
    return value * bytesPerMB;
}
