import { apiFetch } from './http';
import type {
  BankCsvTemplate,
  CreateBankCsvTemplateRequest,
  UpdateBankCsvTemplateRequest,
  PagedImportsResponse,
  ImportResultResponse,
  ReviewItem,
  ApproveRequest,
  BankPdfTemplate,
  BankPdfTemplateLedgerResponse,
  CreateBankPdfTemplateRequest,
  UpdateBankPdfTemplateRequest,
} from '../types/bankStatement';

export const getTemplates = async (): Promise<BankCsvTemplate[]> => {
  const response = await apiFetch('/api/bank-statement/template');
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '載入範本列表失敗');
  }
  return response.json();
};

export const createTemplate = async (req: CreateBankCsvTemplateRequest): Promise<void> => {
  const response = await apiFetch('/api/bank-statement/template', {
    method:  'POST',
    headers: { 'Content-Type': 'application/json' },
    body:    JSON.stringify(req),
  });
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '建立範本失敗');
  }
};

export const updateTemplate = async (templateId: string, req: UpdateBankCsvTemplateRequest): Promise<void> => {
  const response = await apiFetch(`/api/bank-statement/template/${templateId}`, {
    method:  'PUT',
    headers: { 'Content-Type': 'application/json' },
    body:    JSON.stringify(req),
  });
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '更新範本失敗');
  }
};

export const deactivateTemplate = async (templateId: string, expectedVersion: number): Promise<void> => {
  const response = await apiFetch(`/api/bank-statement/template/${templateId}`, {
    method:  'DELETE',
    headers: { 'Content-Type': 'application/json' },
    body:    JSON.stringify({ expected_version: expectedVersion }),
  });
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '停用範本失敗');
  }
};

export const getImportsPaged = async (
  ledgerId: number,
  page: number,
  size: number,
): Promise<PagedImportsResponse> => {
  const params = new URLSearchParams({ page: String(page), size: String(size) });
  if (ledgerId > 0) params.set('ledger_id', String(ledgerId));
  const response = await apiFetch(`/api/bank-statement/paged?${params}`);
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '載入匯入清單失敗');
  }
  return response.json();
};

export const importCSV = async (formData: FormData): Promise<void> => {
  const response = await apiFetch('/api/bank-statement/import', {
    method: 'POST',
    body:   formData,
  });
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '上傳失敗');
  }
};

export const importExcel = async (formData: FormData): Promise<void> => {
  const response = await apiFetch('/api/bank-statement/import/excel', {
    method: 'POST',
    body:   formData,
  });
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '上傳失敗');
  }
};

export const getImportResult = async (importId: number): Promise<ImportResultResponse> => {
  const response = await apiFetch(`/api/bank-statement/${importId}/result`);
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '載入明細失敗');
  }
  return response.json();
};

export const autoMatch = async (importId: number): Promise<void> => {
  const response = await apiFetch(`/api/bank-statement/${importId}/auto-match`, { method: 'POST' });
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '自動比對失敗');
  }
};

export const reMatch = async (importId: number): Promise<void> => {
  const response = await apiFetch(`/api/bank-statement/${importId}/re-match`, { method: 'POST' });
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '重新比對失敗');
  }
};

export const manualMatch = async (
  importId: number,
  bankTxnId: number,
  entryId: number,
): Promise<void> => {
  const response = await apiFetch(`/api/bank-statement/${importId}/txn/${bankTxnId}/match`, {
    method:  'PUT',
    headers: { 'Content-Type': 'application/json' },
    body:    JSON.stringify({ entry_id: entryId }),
  });
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '手動比對失敗');
  }
};

export const getReviewItems = async (importId: number): Promise<ReviewItem[]> => {
  const response = await apiFetch(`/api/bank-statement/${importId}/review`);
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '載入審查項目失敗');
  }
  return response.json();
};

export const approveTxn = async (
  importId: number,
  bankTxnId: number,
  req: ApproveRequest,
): Promise<void> => {
  const response = await apiFetch(`/api/bank-statement/${importId}/txn/${bankTxnId}/approve`, {
    method:  'POST',
    headers: { 'Content-Type': 'application/json' },
    body:    JSON.stringify(req),
  });
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '核准失敗');
  }
};

export const ignoreTxn = async (importId: number, bankTxnId: number): Promise<void> => {
  const response = await apiFetch(`/api/bank-statement/${importId}/txn/${bankTxnId}/ignore`, {
    method: 'POST',
  });
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '忽略失敗');
  }
};

export const completeImport = async (
  importId: number,
  importUuid: string,
  expectedVersion: number,
): Promise<void> => {
  const response = await apiFetch(`/api/bank-statement/${importId}/complete`, {
    method:  'POST',
    headers: { 'Content-Type': 'application/json' },
    body:    JSON.stringify({ import_uuid: importUuid, expected_version: expectedVersion }),
  });
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '完成匯入失敗');
  }
};

// ── PDF Template ─────────────────────────

export const getPdfTemplateLedgers = async (
  templateUuid: string,
): Promise<BankPdfTemplateLedgerResponse[]> => {
  const response = await apiFetch(`/api/bank-pdf-template/${templateUuid}/ledgers`);
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '載入範本帳本失敗');
  }
  return response.json();
};

export const getPdfTemplates = async (): Promise<BankPdfTemplate[]> => {
  const response = await apiFetch('/api/bank-pdf-template');
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '載入 PDF 範本列表失敗');
  }
  return response.json();
};

export const createPdfTemplate = async (req: CreateBankPdfTemplateRequest): Promise<void> => {
  const response = await apiFetch('/api/bank-pdf-template', {
    method:  'POST',
    headers: { 'Content-Type': 'application/json' },
    body:    JSON.stringify(req),
  });
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '建立 PDF 範本失敗');
  }
};

export const updatePdfTemplate = async (
  templateUuid: string,
  req: UpdateBankPdfTemplateRequest,
): Promise<void> => {
  const response = await apiFetch(`/api/bank-pdf-template/${templateUuid}`, {
    method:  'PUT',
    headers: { 'Content-Type': 'application/json' },
    body:    JSON.stringify(req),
  });
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '更新 PDF 範本失敗');
  }
};

export const deactivatePdfTemplate = async (
  templateUuid: string,
  expectedVersion: number,
): Promise<void> => {
  const response = await apiFetch(`/api/bank-pdf-template/${templateUuid}`, {
    method:  'DELETE',
    headers: { 'Content-Type': 'application/json' },
    body:    JSON.stringify({ expected_version: expectedVersion }),
  });
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '停用 PDF 範本失敗');
  }
};

export const importPDF = async (formData: FormData): Promise<void> => {
  const response = await apiFetch('/api/bank-statement/import/pdf', {
    method: 'POST',
    body:   formData,
  });
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '上傳 PDF 失敗');
  }
};
