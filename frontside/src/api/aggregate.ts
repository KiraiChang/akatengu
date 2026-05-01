import { apiFetch } from './http';
import type { AppendEventCmd } from '../types/event';

export const getAggregateVersion = async (aggregateType: string, aggregateId: string): Promise<number> => {
  const response = await apiFetch(`/api/aggerate/${aggregateType}?aggerate_id=${aggregateId}`);
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '取得版本失敗');
  }
  return response.json();
};

export const appendEvent = async <T>(cmd: AppendEventCmd<T>): Promise<void> => {
  const response = await apiFetch('/api/event/append', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(cmd),
  });
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '事件寫入失敗');
  }
};
