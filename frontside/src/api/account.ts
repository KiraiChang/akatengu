import { apiFetch } from './http';
import { getAggregateVersion, appendEvent } from './aggregate';
import type { Account, CreateAccountRequest, UpdateAccountRequest } from '../types/account';

export const getAccountAll = async (): Promise<Account[]> => {
  const response = await apiFetch('/api/account/all');
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢失敗');
  }
  return response.json();
};

export const createAccount = async (data: CreateAccountRequest): Promise<void> => {
  const version = await getAggregateVersion('ACCOUNT', data.account_id);
  await appendEvent({
    aggregate_type:   'ACCOUNT',
    aggregate_id:     data.account_id,
    expected_version: version,
    event_type:       'account.created',
    payload:          data,
    metadata:         {},
  });
};

export const updateAccount = async (data: UpdateAccountRequest): Promise<void> => {
  const version = await getAggregateVersion('ACCOUNT', data.account_id);
  await appendEvent({
    aggregate_type:   'ACCOUNT',
    aggregate_id:     data.account_id,
    expected_version: version,
    event_type:       'account.updated',
    payload:          data,
    metadata:         {},
  });
};
