import { apiFetch } from './http';
import type {
  AccountChildrenBalanceResult,
  AccountJournalEntryRow,
  AccountMonthlyBalance,
  PagedResponse,
} from '../types/accountAnalysis';

export const getChildrenBalance = async (accountId: string): Promise<AccountChildrenBalanceResult> => {
  const response = await apiFetch(`/api/account/${encodeURIComponent(accountId)}/children-balance`);
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢子科目餘額失敗');
  }
  return response.json();
};

export const getAccountEntries = async (
  accountId: string,
  from: string,
  to: string,
  page: number,
  pageSize: number,
): Promise<PagedResponse<AccountJournalEntryRow>> => {
  const params = new URLSearchParams({
    from,
    to,
    page:      String(page),
    page_size: String(pageSize),
  });
  const response = await apiFetch(`/api/account/${encodeURIComponent(accountId)}/entries?${params}`);
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢分錄明細失敗');
  }
  return response.json();
};

export const getMonthlyBalance = async (
  accountId: string,
  from: string,
  to: string,
): Promise<AccountMonthlyBalance[]> => {
  const params = new URLSearchParams({ from, to });
  const response = await apiFetch(`/api/account/${encodeURIComponent(accountId)}/monthly-balance?${params}`);
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢月別趨勢失敗');
  }
  return response.json();
};
