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

export const getExchangeRates = async (currency?: string): Promise<ExchangeRate[]> => {
  const query = currency ? `?currency=${encodeURIComponent(currency)}` : '';
  const res = await apiFetch(`/api/exchange-rate${query}`);
  if (!res.ok) {
    const p = await res.json();
    throw new Error(p.detail ?? p.title ?? '查詢匯率失敗');
  }
  return res.json();
};
