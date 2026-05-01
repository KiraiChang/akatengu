import { apiFetch } from './http';
import { getAggregateVersion, appendEvent } from './aggregate';
import type { LedgerAccount, LedgerBalance, LedgerAccountCreatePayload, LedgerAccountUpdatePayload } from '../types/ledger';

let _cache: LedgerAccount[] | null = null;

export const getLedgerAccountAll = async (force = false): Promise<LedgerAccount[]> => {
  if (_cache && !force) return _cache;
  const response = await apiFetch('/api/ledger/all');
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢失敗');
  }
  _cache = await response.json() as LedgerAccount[];
  return _cache;
};

export const invalidateLedgerCache = (): void => { _cache = null; };

export const getLedgerBalances = async (): Promise<LedgerBalance[]> => {
  const response = await apiFetch('/api/ledger/all_balance');
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢餘額失敗');
  }
  return response.json();
};

export const createLedgerAccount = async (
  payload: LedgerAccountCreatePayload,
): Promise<void> => {
  const version = await getAggregateVersion('ACCOUNT', payload.account_id);
  await appendEvent({
    aggregate_type:   'ACCOUNT',
    aggregate_id:     payload.account_id,
    expected_version: version,
    event_type:       'ledger_account.created',
    payload,
    metadata:         {},
  });
};

export const updateLedgerAccount = async (
  payload: LedgerAccountUpdatePayload,
): Promise<void> => {
  const version = await getAggregateVersion('ACCOUNT', payload.account_id);
  await appendEvent({
    aggregate_type:   'ACCOUNT',
    aggregate_id:     payload.account_id,
    expected_version: version,
    event_type:       'ledger_account.updated',
    payload,
    metadata:         {},
  });
};
