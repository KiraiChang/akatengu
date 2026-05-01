import { apiFetch } from './http';
import { getAggregateVersion, appendEvent } from './aggregate';
import type { Transaction, Entry, TransactionCreatedPayload } from '../types/transaction';
import type { PaginatedResponse } from '../types/pagination';

export interface GetTransactionPagedParams {
  page: number;
  pageSize: number;
}

export const getTransactionPaged = async (
  params: GetTransactionPagedParams,
): Promise<PaginatedResponse<Transaction>> => {
  const qs = new URLSearchParams({
    page: String(params.page),
    page_size: String(params.pageSize),
  });

  const response = await apiFetch(`/api/txn/paged?${qs}`);
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢失敗');
  }
  return response.json();
};

export const getEntries = async (txnId: number): Promise<Entry[]> => {
  const response = await apiFetch(`/api/txn/${txnId}`);
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢失敗');
  }
  return response.json();
};

export const createTransaction = async (payload: TransactionCreatedPayload): Promise<void> => {
  const version = await getAggregateVersion('TRANSACTION', '');
  await appendEvent({
    aggregate_type:   'TRANSACTION',
    aggregate_id:     '',
    expected_version: version,
    event_type:       'transaction.created',
    payload,
    metadata:         {},
  });
};
