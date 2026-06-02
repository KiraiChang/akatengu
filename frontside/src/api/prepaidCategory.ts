import { apiFetch } from './http';
import type {
  PrepaidCategory,
  PrepaidCategoryCreatePayload,
  PrepaidCategoryUpdatePayload,
} from '../types/prepaidCategory';

export const getAllPrepaidCategories = async (): Promise<PrepaidCategory[]> => {
  const response = await apiFetch('/api/prepaid/category/all');
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢失敗');
  }
  return response.json() as Promise<PrepaidCategory[]>;
};

export const createPrepaidCategory = async (payload: PrepaidCategoryCreatePayload): Promise<void> => {
  const response = await apiFetch('/api/prepaid/category', {
    method:  'POST',
    headers: { 'Content-Type': 'application/json' },
    body:    JSON.stringify(payload),
  });
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '新增失敗');
  }
};

export const updatePrepaidCategory = async (
  categoryUUID: string,
  payload: PrepaidCategoryUpdatePayload,
): Promise<void> => {
  const response = await apiFetch(`/api/prepaid/category/${categoryUUID}`, {
    method:  'PUT',
    headers: { 'Content-Type': 'application/json' },
    body:    JSON.stringify(payload),
  });
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '修改失敗');
  }
};

export const deletePrepaidCategory = async (categoryUUID: string, expectedVersion: number): Promise<void> => {
  const response = await apiFetch(`/api/prepaid/category/${categoryUUID}`, {
    method:  'DELETE',
    headers: { 'Content-Type': 'application/json' },
    body:    JSON.stringify({ expected_version: expectedVersion }),
  });
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '關閉失敗');
  }
};
