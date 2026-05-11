export function displayItemType(type: string, t: (key: string) => string) {
    if (type === 'file') {
        return t('itemTypeFile');
    }
    if (type === 'folder') {
        return t('itemTypeFolder');
    }
    return type;
}
