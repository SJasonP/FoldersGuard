import { useEffect } from 'react';
import type { FormInstance } from 'antd';

export function useResetFormOnClose<T>(form: FormInstance<T>, open: boolean) {
  useEffect(() => {
    if (!open) {
      form.resetFields();
    }
  }, [form, open]);
}
