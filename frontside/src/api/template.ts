import { apiFetch } from './http';
import type { TransactionTemplate, TransactionTemplateDetail, SaveTemplateRequest } from '../types/template';

export const getTemplates = async (q?: string): Promise<TransactionTemplate[]> => {
  const url = q ? `/api/template?q=${encodeURIComponent(q)}` : '/api/template';
  const response = await apiFetch(url);
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '載入範本列表失敗');
  }
  return response.json();
};

export const getTemplate = async (id: number): Promise<TransactionTemplateDetail> => {
  const response = await apiFetch(`/api/template/${id}`);
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '載入範本失敗');
  }
  return response.json();
};

export const createTemplate = async (req: SaveTemplateRequest): Promise<TransactionTemplateDetail> => {
  const response = await apiFetch('/api/template', {
    method:  'POST',
    headers: { 'Content-Type': 'application/json' },
    body:    JSON.stringify(req),
  });
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '建立範本失敗');
  }
  return response.json();
};

export const updateTemplate = async (id: number, req: SaveTemplateRequest): Promise<void> => {
  const response = await apiFetch(`/api/template/${id}`, {
    method:  'PUT',
    headers: { 'Content-Type': 'application/json' },
    body:    JSON.stringify(req),
  });
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '更新範本失敗');
  }
};

export const deleteTemplate = async (id: number): Promise<void> => {
  const response = await apiFetch(`/api/template/${id}`, { method: 'DELETE' });
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '刪除範本失敗');
  }
};
