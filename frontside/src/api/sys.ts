import { apiFetch } from './http';
import type { SysAccount } from '../types/sys';

export const getSysAccounts = async (): Promise<SysAccount[]> => {
  const response = await apiFetch('/api/sys/account');
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢失敗');
  }
  return response.json();
};

export const updateSysAccount = async (data: SysAccount): Promise<void> => {
  const response = await apiFetch('/api/sys/account', {
    method:  'POST',
    headers: { 'Content-Type': 'application/json' },
    body:    JSON.stringify(data),
  });
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '更新失敗');
  }
};
