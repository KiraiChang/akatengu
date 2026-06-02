import { apiFetch } from './http';
import type {
  FixedAssetCategory,
  FixedAssetCategoryCreatePayload,
  FixedAssetCategoryUpdatePayload,
} from '../types/fixedAssetCategory';

export const getAllFixedAssetCategories = async (): Promise<FixedAssetCategory[]> => {
  const response = await apiFetch('/api/fixed_asset/category/all');
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢失敗');
  }
  return response.json() as Promise<FixedAssetCategory[]>;
};

export const createFixedAssetCategory = async (payload: FixedAssetCategoryCreatePayload): Promise<void> => {
  const response = await apiFetch('/api/fixed_asset/category', {
    method:  'POST',
    headers: { 'Content-Type': 'application/json' },
    body:    JSON.stringify(payload),
  });
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '新增失敗');
  }
};

export const updateFixedAssetCategory = async (
  categoryUUID: string,
  payload: FixedAssetCategoryUpdatePayload,
): Promise<void> => {
  const response = await apiFetch(`/api/fixed_asset/category/${categoryUUID}`, {
    method:  'PUT',
    headers: { 'Content-Type': 'application/json' },
    body:    JSON.stringify(payload),
  });
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '修改失敗');
  }
};

export const deleteFixedAssetCategory = async (categoryUUID: string, expectedVersion: number): Promise<void> => {
  const response = await apiFetch(`/api/fixed_asset/category/${categoryUUID}`, {
    method:  'DELETE',
    headers: { 'Content-Type': 'application/json' },
    body:    JSON.stringify({ expected_version: expectedVersion }),
  });
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '關閉失敗');
  }
};
