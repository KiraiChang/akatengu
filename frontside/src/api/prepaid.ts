import { apiFetch } from './http';
import { getAggregateVersion, appendEvent } from './aggregate';
import type {
  Prepaid,
  PrepaidAmortization,
  PrepaidCreatedPayload,
  PrepaidAmortizedPayload,
  PrepaidDisposedPayload,
} from '../types/prepaid';

export const getAllPrepaids = async (): Promise<Prepaid[]> => {
  const response = await apiFetch('/api/prepaid/all');
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢失敗');
  }
  return response.json() as Promise<Prepaid[]>;
};

export const getPrepaidAmortizations = async (prepaidId: number): Promise<PrepaidAmortization[]> => {
  const response = await apiFetch(`/api/prepaid/${prepaidId}/amortizations`);
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢失敗');
  }
  return response.json() as Promise<PrepaidAmortization[]>;
};

export const createPrepaid = async (payload: PrepaidCreatedPayload): Promise<void> => {
  const version = await getAggregateVersion('TRANSACTION', '');
  await appendEvent<PrepaidCreatedPayload>({
    aggregate_type:   'TRANSACTION',
    aggregate_id:     '',
    expected_version: version,
    event_type:       'prepaid.created',
    payload,
    metadata:         {},
  });
};

export const amortizePrepaid = async (payload: PrepaidAmortizedPayload): Promise<void> => {
  const version = await getAggregateVersion('TRANSACTION', '');
  await appendEvent<PrepaidAmortizedPayload>({
    aggregate_type:   'TRANSACTION',
    aggregate_id:     '',
    expected_version: version,
    event_type:       'prepaid.amortized',
    payload,
    metadata:         {},
  });
};

export const disposePrepaid = async (payload: PrepaidDisposedPayload): Promise<void> => {
  const version = await getAggregateVersion('TRANSACTION', '');
  await appendEvent<PrepaidDisposedPayload>({
    aggregate_type:   'TRANSACTION',
    aggregate_id:     '',
    expected_version: version,
    event_type:       'prepaid.disposed',
    payload,
    metadata:         {},
  });
};
