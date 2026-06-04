import { apiFetch } from './http';
import type { AggregateVersionAudit, EventStoreAudit, CheckpointAudit, SnapshotAudit, ExchangeRate } from '../types/audit';
import type { PaginatedResponse } from '../types/pagination';

export const getAggregateVersions = async (): Promise<AggregateVersionAudit[]> => {
  const res = await apiFetch('/api/audit/aggregate-version');
  if (!res.ok) {
    const p = await res.json();
    throw new Error(p.detail ?? p.title ?? '查詢 Aggregate 版本失敗');
  }
  return res.json();
};

export const getEventStorePaged = async (page: number, pageSize: number): Promise<PaginatedResponse<EventStoreAudit>> => {
  const res = await apiFetch(`/api/audit/event?page=${page}&page_size=${pageSize}`);
  if (!res.ok) {
    const p = await res.json();
    throw new Error(p.detail ?? p.title ?? '查詢事件紀錄失敗');
  }
  return res.json();
};

export const getCheckpoints = async (): Promise<CheckpointAudit[]> => {
  const res = await apiFetch('/api/audit/checkpoint');
  if (!res.ok) {
    const p = await res.json();
    throw new Error(p.detail ?? p.title ?? '查詢 Checkpoint 失敗');
  }
  return res.json();
};

export const getSnapshots = async (): Promise<SnapshotAudit[]> => {
  const res = await apiFetch('/api/audit/snapshot');
  if (!res.ok) {
    const p = await res.json();
    throw new Error(p.detail ?? p.title ?? '查詢快照失敗');
  }
  return res.json();
};

export const replayProjection = async (fromEventId = 0, aggregateType?: string): Promise<{ replayed_count: number }> => {
  const res = await apiFetch('/api/audit/replay', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ from_event_id: fromEventId, aggregate_type: aggregateType ?? null }),
  });
  if (!res.ok) {
    const p = await res.json();
    throw new Error(p.detail ?? p.title ?? '重建 Projection 失敗');
  }
  return res.json();
};

export const exportEvents = async (password?: string): Promise<void> => {
  const res = await apiFetch('/api/audit/event/export', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ password: password || null }),
  });
  if (!res.ok) {
    const p = await res.json();
    throw new Error(p.detail ?? p.title ?? '匯出事件失敗');
  }
  const blob = await res.blob();
  const disposition = res.headers.get('Content-Disposition') ?? '';
  const match = /filename=([^\s;]+)/.exec(disposition);
  const filename = match ? match[1] : 'events.json';
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
};

export const importEvents = async (file: File, password?: string): Promise<{ imported: number }> => {
  const fd = new FormData();
  fd.append('file', file);
  if (password) fd.append('password', password);
  const res = await apiFetch('/api/audit/event/import', { method: 'POST', body: fd });
  if (!res.ok) {
    const p = await res.json();
    throw new Error(p.detail ?? p.title ?? '匯入事件失敗');
  }
  return res.json();
};

export const getExchangeRates = async (currency?: string): Promise<ExchangeRate[]> => {
  const query = currency ? `?currency=${encodeURIComponent(currency)}` : '';
  const res = await apiFetch(`/api/exchange-rate${query}`);
  if (!res.ok) {
    const p = await res.json();
    throw new Error(p.detail ?? p.title ?? '查詢匯率失敗');
  }
  return res.json();
};
