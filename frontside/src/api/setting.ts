import { apiFetch } from './http';
import type {
  LedgerAccountTypeConfigResult,
  AssetTypeAccountConfigResult,
  UpdateAssetTypePayload,
  LedgerAccountType,
  AssetType,
} from '../types/setting';

export const getLedgerAccountTypeConfigs = async (): Promise<LedgerAccountTypeConfigResult[]> => {
  const res = await apiFetch('/api/setting/ledger-account-type');
  if (!res.ok) {
    const p = await res.json();
    throw new Error(p.detail ?? p.title ?? '查詢失敗');
  }
  return res.json();
};

export const updateLedgerAccountTypeConfig = async (
  type: LedgerAccountType,
  accountId: string,
): Promise<void> => {
  const res = await apiFetch(`/api/setting/ledger-account-type/${type}`, {
    method:  'PUT',
    headers: { 'Content-Type': 'application/json' },
    body:    JSON.stringify({ account_id: accountId }),
  });
  if (!res.ok) {
    const p = await res.json();
    throw new Error(p.detail ?? p.title ?? '更新失敗');
  }
};

export const getAssetTypeConfigs = async (): Promise<AssetTypeAccountConfigResult[]> => {
  const res = await apiFetch('/api/setting/asset-type');
  if (!res.ok) {
    const p = await res.json();
    throw new Error(p.detail ?? p.title ?? '查詢失敗');
  }
  return res.json();
};

export const updateAssetTypeConfig = async (
  type: AssetType,
  payload: UpdateAssetTypePayload,
): Promise<void> => {
  const res = await apiFetch(`/api/setting/asset-type/${type}`, {
    method:  'PUT',
    headers: { 'Content-Type': 'application/json' },
    body:    JSON.stringify(payload),
  });
  if (!res.ok) {
    const p = await res.json();
    throw new Error(p.detail ?? p.title ?? '更新失敗');
  }
};
