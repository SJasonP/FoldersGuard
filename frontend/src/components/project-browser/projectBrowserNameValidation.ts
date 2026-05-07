export function validateProjectItemName(name: string | undefined) {
  if (!name) {
    return false;
  }
  if (name === '.' || name === '..') {
    return false;
  }
  return !name.includes('/') && !name.includes('\\');
}

export function projectItemNameRules(t: (key: string) => string) {
  return [
    { required: true, message: t('invalidItemName') },
    {
      validator(_: unknown, value: string | undefined) {
        if (validateProjectItemName(value)) {
          return Promise.resolve();
        }
        return Promise.reject(new Error(t('invalidItemName')));
      },
    },
  ];
}
