import { apiFetch } from './http';
import type { DashboardSummary, MonthlyTrendItem, LedgerBalance } from '../types/dashboard';

export const getDashboardSummary = async (): Promise<DashboardSummary> => {
  const res = await apiFetch('/api/dashboard/summary');
  if (!res.ok) {
    const p = await res.json();
    throw new Error(p.detail ?? p.title ?? '查詢儀表板摘要失敗');
  }
  return res.json();
};

export const getMonthlyTrend = async (months = 12): Promise<MonthlyTrendItem[]> => {
  const res = await apiFetch(`/api/dashboard/monthly-trend?months=${months}`);
  if (!res.ok) {
    const p = await res.json();
    throw new Error(p.detail ?? p.title ?? '查詢月度趨勢失敗');
  }
  return res.json();
};

export const getLedgerBalances = async (): Promise<LedgerBalance[]> => {
  const res = await apiFetch('/api/ledger/balances');
  if (!res.ok) {
    const p = await res.json();
    throw new Error(p.detail ?? p.title ?? '查詢帳戶餘額失敗');
  }
  return res.json();
};
