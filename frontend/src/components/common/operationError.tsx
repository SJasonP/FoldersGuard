import {Space, Typography} from 'antd';
import type {HookAPI as ModalHookAPI} from 'antd/es/modal/useModal';

const secretPatterns = [
    /(password\s*[:=]\s*)([^\s,;]+)/gi,
    /(passphrase\s*[:=]\s*)([^\s,;]+)/gi,
    /(database[_\s-]*key\s*[:=]\s*)([^\s,;]+)/gi,
    /(key[_\s-]*material\s*[:=]\s*)([^\s,;]+)/gi,
];

const codedErrorMessages: Array<[string, string]> = [
    ['FG_INVALID_PASSWORD', 'errorPasswordIncorrect'],
    ['FG_OUTPUT_FOLDER_NOT_EMPTY', 'errorOutputFolderNotEmpty'],
    ['FG_OUTPUT_INSIDE_SOURCE', 'errorOutputInsideSource'],
    ['FG_OUTPUT_CONTAINS_SOURCE', 'errorOutputContainsSource'],
    ['FG_SOURCE_TARGET_SAME', 'errorSourceTargetSame'],
];

function rawErrorMessage(error: unknown): string {
    if (error instanceof Error) {
        return error.message;
    }
    if (typeof error === 'string') {
        return error;
    }
    if (error && typeof error === 'object') {
        const record = error as Record<string, unknown>;
        for (const key of ['message', 'error', 'details', 'detail', 'reason']) {
            const value = record[key];
            if (typeof value === 'string' && value.trim()) {
                return value;
            }
        }
        try {
            return JSON.stringify(error);
        } catch {
            return Object.prototype.toString.call(error);
        }
    }
    return String(error ?? '');
}

export function technicalErrorMessage(error: unknown) {
    const message = rawErrorMessage(error);
    return secretPatterns.reduce((current, pattern) => current.replace(pattern, '$1[redacted]'), message).trim();
}

function userFacingErrorMessage(details: string, t: (key: string) => string) {
    for (const [code, messageKey] of codedErrorMessages) {
        if (details.includes(code)) {
            return t(messageKey);
        }
    }

    const lower = details.toLowerCase();
    if (lower.includes('database password is incorrect') || lower.includes('file is not a database') || lower.includes('file is encrypted or is not a database')) {
        return t('errorPasswordIncorrect');
    }
    if (lower.includes('folder is not empty')) {
        return t('errorOutputFolderNotEmpty');
    }
    if (lower.includes('output path must be outside the source folder')) {
        return t('errorOutputInsideSource');
    }
    if (lower.includes('output path must not contain the source folder')) {
        return t('errorOutputContainsSource');
    }
    if (lower.includes('source and target paths must be different')) {
        return t('errorSourceTargetSame');
    }
    return t('errorOperationFailed');
}

function operationErrorMessage(error: unknown, t: (key: string) => string) {
    const details = technicalErrorMessage(error);
    return {
        details,
        userMessage: details ? userFacingErrorMessage(details, t) : '',
    };
}

export function showOperationError(
    modalApi: ModalHookAPI,
    title: string,
    error: unknown,
    t: (key: string) => string,
) {
    const {details, userMessage} = operationErrorMessage(error, t);
    modalApi.error({
        title,
        content: details ? <Typography.Paragraph>{userMessage}</Typography.Paragraph> : undefined,
    });
}

export function showStartupError(
    modalApi: ModalHookAPI,
    title: string,
    dataDirectory: string,
    t: (key: string) => string,
) {
    modalApi.error({
        title,
        closable: false,
        maskClosable: false,
        content: (
            <Space direction="vertical" size="middle">
                <Space direction="vertical" size={4}>
                    <Typography.Text>{t('dataDirectory')}</Typography.Text>
                    <Typography.Text code>{dataDirectory}</Typography.Text>
                </Space>
            </Space>
        ),
    });
}
