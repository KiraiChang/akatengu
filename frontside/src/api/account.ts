import { authStore } from '../stores/auth.svelte';
import type { Account } from '../types/account';
import type { PaginatedResponse } from '../types/pagination';

export interface GetAccountPagedParams {
  page: number;
  pageSize: number;
}

export const getAccountPaged = async (
  params: GetAccountPagedParams,
): Promise<PaginatedResponse<Account>> => {
  const qs = new URLSearchParams({
    page: String(params.page),
    page_size: String(params.pageSize),
  });

  const response = await fetch(`/api/account/paged?${qs}`, {
    headers: {
      Authorization: `Bearer ${authStore.token}`,
    },
  });

  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢失敗');
  }

  return response.json();
};
