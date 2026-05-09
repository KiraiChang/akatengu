import { apiFetch } from './http';
import { getAggregateVersion, appendEvent } from './aggregate';
import type { Investment, InvestmentLot, InvestmentPosition, InvestmentCreatedPayload, InvestmentUpdatedPayload, InvestmentBoughtPayload, InvestmentSoldPayload, InvestmentLotDisposal, InvestmentMovement } from '../types/investment';
import type { PaginatedResponse } from '../types/pagination';

export const getInvestmentPaged = async (params: {
  page:     number;
  pageSize: number;
}): Promise<PaginatedResponse<Investment>> => {
  const qs = new URLSearchParams({
    page:      String(params.page),
    page_size: String(params.pageSize),
  });
  const response = await apiFetch(`/api/investment/paged?${qs}`);
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢失敗');
  }
  return response.json();
};

export const createInvestment = async (payload: InvestmentCreatedPayload): Promise<void> => {
  const aggregateId = '';
  const version = await getAggregateVersion('TRANSACTION', aggregateId);
  await appendEvent<InvestmentCreatedPayload>({
    aggregate_type:   'TRANSACTION',
    aggregate_id:     aggregateId,
    expected_version: version,
    event_type:       'investment.created',
    payload,
    metadata:         {},
  });
};

export const updateInvestment = async (payload: InvestmentUpdatedPayload): Promise<void> => {
  const aggregateId = '';
  const version = await getAggregateVersion('TRANSACTION', aggregateId);
  await appendEvent<InvestmentUpdatedPayload>({
    aggregate_type:   'TRANSACTION',
    aggregate_id:     aggregateId,
    expected_version: version,
    event_type:       'investment.updated',
    payload,
    metadata:         {},
  });
};

export const getOpenLotsPaged = async (
  investmentId: number,
  params: { page: number; pageSize: number },
): Promise<PaginatedResponse<InvestmentLot>> => {
  const qs = new URLSearchParams({
    page:      String(params.page),
    page_size: String(params.pageSize),
    txn_id:    String(investmentId),
  });
  const response = await apiFetch(`/api/investment/lot/paged?${qs}`);
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢失敗');
  }
  return response.json();
};

export const getPosition = async (investmentId: number): Promise<InvestmentPosition | null> => {
  const qs = new URLSearchParams({ investment_id: String(investmentId) });
  const response = await apiFetch(`/api/investment/position?${qs}`);
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢失敗');
  }
  return response.json();
};

export const buyInvestment = async (payload: InvestmentBoughtPayload): Promise<void> => {
  const aggregateId = '';
  const version = await getAggregateVersion('TRANSACTION', aggregateId);
  await appendEvent<InvestmentBoughtPayload>({
    aggregate_type:   'TRANSACTION',
    aggregate_id:     aggregateId,
    expected_version: version,
    event_type:       'investment.bought',
    payload,
    metadata:         {},
  });
};

export const getLotDisposalsPaged = async (
  lotId: number,
  params: { page: number; pageSize: number },
): Promise<PaginatedResponse<InvestmentLotDisposal>> => {
  const qs = new URLSearchParams({
    page:      String(params.page),
    page_size: String(params.pageSize),
    lot_id:    String(lotId),
  });
  const response = await apiFetch(`/api/investment/lot_disposal/paged?${qs}`);
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢失敗');
  }
  return response.json();
};

export const getMovementPaged = async (
  investmentId: number,
  params: { page: number; pageSize: number },
): Promise<PaginatedResponse<InvestmentMovement>> => {
  const qs = new URLSearchParams({
    page:          String(params.page),
    page_size:     String(params.pageSize),
    investment_id: String(investmentId),
  });
  const response = await apiFetch(`/api/investment/movement/paged?${qs}`);
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢失敗');
  }
  return response.json();
};

export const sellInvestment = async (payload: InvestmentSoldPayload): Promise<void> => {
  const aggregateId = '';
  const version = await getAggregateVersion('TRANSACTION', aggregateId);
  await appendEvent<InvestmentSoldPayload>({
    aggregate_type:   'TRANSACTION',
    aggregate_id:     aggregateId,
    expected_version: version,
    event_type:       'investment.sold',
    payload,
    metadata:         {},
  });
};