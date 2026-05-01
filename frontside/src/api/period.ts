import { apiFetch } from './http';
import { getAggregateVersion, appendEvent } from './aggregate';
import type {
  PeriodClosing,
  PeriodType,
  PeriodMonthStartedPayload,
  PeriodMonthClosedPayload,
  PeriodMonthReopenedPayload,
  PeriodAnnualStartedPayload,
  PeriodAnnualClosedPayload,
  PeriodAnnualReopenedPayload,
} from '../types/period';
import type { PaginatedResponse } from '../types/pagination';

export interface GetPeriodPagedParams {
  page:     number;
  pageSize: number;
}

export const getPeriodPaged = async (
  periodType: PeriodType,
  params: GetPeriodPagedParams,
): Promise<PaginatedResponse<PeriodClosing>> => {
  const qs = new URLSearchParams({
    page:      String(params.page),
    page_size: String(params.pageSize),
  });

  const response = await apiFetch(`/api/period/${periodType}?${qs}`);
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢失敗');
  }
  return response.json();
};

// ── 月結事件 ──────────────────────────────────

export const startMonthPeriod = async (payload: PeriodMonthStartedPayload): Promise<void> => {
  const version = await getAggregateVersion('ACCOUNT', '');
  await appendEvent({
    aggregate_type:   'ACCOUNT',
    aggregate_id:     '',
    expected_version: version,
    event_type:       'period.month_started',
    payload,
    metadata:         {},
  });
};

export const closeMonthPeriod = async (payload: PeriodMonthClosedPayload): Promise<void> => {
  const version = await getAggregateVersion('ACCOUNT', '');
  await appendEvent({
    aggregate_type:   'ACCOUNT',
    aggregate_id:     '',
    expected_version: version,
    event_type:       'period.month_closed',
    payload,
    metadata:         {},
  });
};

export const reopenMonthPeriod = async (payload: PeriodMonthReopenedPayload): Promise<void> => {
  const version = await getAggregateVersion('ACCOUNT', '');
  await appendEvent({
    aggregate_type:   'ACCOUNT',
    aggregate_id:     '',
    expected_version: version,
    event_type:       'period.month_reopened',
    payload,
    metadata:         {},
  });
};

// ── 年結事件 ──────────────────────────────────

export const startAnnualPeriod = async (payload: PeriodAnnualStartedPayload): Promise<void> => {
  const version = await getAggregateVersion('ACCOUNT', '');
  await appendEvent({
    aggregate_type:   'ACCOUNT',
    aggregate_id:     '',
    expected_version: version,
    event_type:       'period.annual_started',
    payload,
    metadata:         {},
  });
};

export const closeAnnualPeriod = async (payload: PeriodAnnualClosedPayload): Promise<void> => {
  const version = await getAggregateVersion('ACCOUNT', '');
  await appendEvent({
    aggregate_type:   'ACCOUNT',
    aggregate_id:     '',
    expected_version: version,
    event_type:       'period.annual_closed',
    payload,
    metadata:         {},
  });
};

export const reopenAnnualPeriod = async (payload: PeriodAnnualReopenedPayload): Promise<void> => {
  const version = await getAggregateVersion('ACCOUNT', '');
  await appendEvent({
    aggregate_type:   'ACCOUNT',
    aggregate_id:     '',
    expected_version: version,
    event_type:       'period.annual_reopened',
    payload,
    metadata:         {},
  });
};
