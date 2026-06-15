import { apiFetch } from './http';
import { getAggregateVersion } from './aggregate';
import type { CFReviewRow, CFActivityCategory } from '../types/cfCategory';

export const getCFReview = async (dateFrom?: string, dateTo?: string): Promise<CFReviewRow[]> => {
  const qs = new URLSearchParams();
  if (dateFrom) qs.set('date_from', dateFrom);
  if (dateTo)   qs.set('date_to',   dateTo);
  const url = `/api/cf-categories${qs.toString() ? `?${qs}` : ''}`;
  const response = await apiFetch(url);
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢失敗');
  }
  return response.json();
};

export const updateCFCategory = async (
  txnUUID: string,
  entries: { entry_uuid: string; cf_category: CFActivityCategory }[],
): Promise<void> => {
  const expectedVersion = await getAggregateVersion('TRANSACTION', txnUUID);
  const response = await apiFetch(`/api/txn/${txnUUID}/cf-category`, {
    method:  'PUT',
    headers: { 'Content-Type': 'application/json' },
    body:    JSON.stringify({ expected_version: expectedVersion, entries }),
  });
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '更新失敗');
  }
};
