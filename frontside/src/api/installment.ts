import { apiFetch } from './http';
import { getAggregateVersion, appendEvent } from './aggregate';
import type { Installment, InstallmentPayment, InstallmentCreatedPayload, InstallmentPeriodPaidPayload } from '../types/installment';
import type { PaginatedResponse } from '../types/pagination';

export const getInstallmentPaged = async (params: {
  page:     number;
  pageSize: number;
}): Promise<PaginatedResponse<Installment>> => {
  const qs = new URLSearchParams({
    page:      String(params.page),
    page_size: String(params.pageSize),
  });
  const response = await apiFetch(`/api/installment/paged?${qs}`);
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢失敗');
  }
  return response.json() as Promise<PaginatedResponse<Installment>>;
};

export const getInstallmentPaymentPaged = async (
  installmentId: number,
  params: { page: number; pageSize: number },
): Promise<PaginatedResponse<InstallmentPayment>> => {
  const qs = new URLSearchParams({
    page:       String(params.page),
    page_size:  String(params.pageSize),
    payment_id: String(installmentId),
  });
  const response = await apiFetch(`/api/installment/payment?${qs}`);
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢失敗');
  }
  return response.json() as Promise<PaginatedResponse<InstallmentPayment>>;
};

export const createInstallment = async (payload: InstallmentCreatedPayload): Promise<void> => {
  const version = await getAggregateVersion('TRANSACTION', '');
  await appendEvent<InstallmentCreatedPayload>({
    aggregate_type:   'TRANSACTION',
    aggregate_id:     '',
    expected_version: version,
    event_type:       'installment.created',
    payload,
    metadata:         {},
  });
};

export const payInstallmentPeriod = async (payload: InstallmentPeriodPaidPayload): Promise<void> => {
  const version = await getAggregateVersion('TRANSACTION', '');
  await appendEvent<InstallmentPeriodPaidPayload>({
    aggregate_type:   'TRANSACTION',
    aggregate_id:     '',
    expected_version: version,
    event_type:       'installment.period_paid',
    payload,
    metadata:         {},
  });
};