import { authStore } from '../stores/auth.svelte';
import type { JournalEntry, CreateJournalEntryRequest } from '../types/journalEntry';
import type { PaginatedResponse } from '../types/pagination';

export const createJournalEntry = async (
  data: CreateJournalEntryRequest,
): Promise<JournalEntry> => {
  const response = await fetch('/api/journal-entry', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization:  `Bearer ${authStore.token}`,
    },
    body: JSON.stringify(data),
  });
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '新增失敗');
  }
  return response.json();
};

export const getJournalEntriesPaged = async (params: {
  page: number;
  pageSize: number;
}): Promise<PaginatedResponse<JournalEntry>> => {
  const qs = new URLSearchParams({
    page:      String(params.page),
    page_size: String(params.pageSize),
  });
  const response = await fetch(`/api/journal-entry/paged?${qs}`, {
    headers: { Authorization: `Bearer ${authStore.token}` },
  });
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢失敗');
  }
  return response.json();
};
