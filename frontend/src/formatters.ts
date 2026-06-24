import i18n from './i18n';

function currentLanguage() {
    return i18n.language === 'zh-CN' ? 'zh-CN' : 'en-US';
}

export function formatDateTime(value: Date | string | null | undefined) {
    if (!value) {
        return '';
    }
    const date = value instanceof Date ? value : new Date(value);
    if (Number.isNaN(date.getTime())) {
        return '';
    }
    return new Intl.DateTimeFormat(currentLanguage(), {
        dateStyle: 'medium',
        timeStyle: 'short',
    }).format(date);
}

export function formatNumber(value: number | null | undefined) {
    return new Intl.NumberFormat(currentLanguage()).format(value ?? 0);
}

export function formatDuration(seconds: number | null | undefined) {
    const total = Math.max(0, Math.round(seconds ?? 0));
    const hours = Math.floor(total / 3600);
    const minutes = Math.floor((total % 3600) / 60);
    const secs = total % 60;
    if (hours > 0) {
        return `${hours}:${String(minutes).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
    }
    return `${minutes}:${String(secs).padStart(2, '0')}`;
}

export function formatThroughput(bytesPerSecond: number | null | undefined) {
    return `${formatFileSize(bytesPerSecond)}/s`;
}

export function formatFileSize(bytes: number | null | undefined) {
    const value = Math.max(0, bytes ?? 0);
    if (value < 1024) {
        return `${formatNumber(value)} B`;
    }
    const units = ['KB', 'MB', 'GB', 'TB', 'PB'];
    let size = value / 1024;
    let unitIndex = 0;
    while (size >= 1024 && unitIndex < units.length - 1) {
        size /= 1024;
        unitIndex++;
    }
    return `${new Intl.NumberFormat(currentLanguage(), {maximumFractionDigits: 1}).format(size)} ${units[unitIndex]}`;
}
