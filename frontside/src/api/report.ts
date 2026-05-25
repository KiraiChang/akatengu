import { apiFetch } from './http';
import type { BalanceSheet, IncomeStatement, CashFlowStatement, DirectCashFlowStatement, EquityStatement } from '../types/report';

export const getBalanceSheet = async (reportDate: string): Promise<BalanceSheet> => {
  const response = await apiFetch(`/api/report/balance_sheet?report_date=${reportDate}`);
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢資產負債表失敗');
  }
  return response.json();
};

export const getIncomeStatement = async (beginDate: string, endDate: string): Promise<IncomeStatement> => {
  const response = await apiFetch(`/api/report/income_statement?begin_date=${beginDate}&end_date=${endDate}`);
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢損益表失敗');
  }
  return response.json();
};

export const getEquityStatement = async (beginDate: string, endDate: string): Promise<EquityStatement> => {
  const response = await apiFetch(`/api/report/equity_statement?begin_date=${beginDate}&end_date=${endDate}`);
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢權益變動表失敗');
  }
  return response.json();
};

export const getCashFlowStatement = async (beginDate: string, endDate: string): Promise<CashFlowStatement> => {
  const response = await apiFetch(`/api/report/cash_flow_statement?begin_date=${beginDate}&end_date=${endDate}`);
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢現金流量表失敗');
  }
  return response.json();
};

export const getDirectCashFlowStatement = async (beginDate: string, endDate: string): Promise<DirectCashFlowStatement> => {
  const response = await apiFetch(`/api/report/cash_flow_statement_direct?begin_date=${beginDate}&end_date=${endDate}`);
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢現金流量表（直接法）失敗');
  }
  return response.json();
};
