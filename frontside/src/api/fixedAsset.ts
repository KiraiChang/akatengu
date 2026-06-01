import { apiFetch } from './http';
import { getAggregateVersion, appendEvent } from './aggregate';
import type {
  FixedAsset,
  FixedAssetDepreciation,
  AssetPurchasedPayload,
  AssetPurchasedWithInstallmentPayload,
  AssetDepreciatedPayload,
  AssetDisposedPayload,
} from '../types/fixedAsset';

export const getAllFixedAssets = async (): Promise<FixedAsset[]> => {
  const response = await apiFetch('/api/fixed_asset/all');
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢失敗');
  }
  return response.json() as Promise<FixedAsset[]>;
};

export const getFixedAssetDepreciations = async (assetId: number): Promise<FixedAssetDepreciation[]> => {
  const response = await apiFetch(`/api/fixed_asset/${assetId}/depreciations`);
  if (!response.ok) {
    const problem = await response.json();
    throw new Error(problem.detail ?? problem.title ?? '查詢失敗');
  }
  return response.json() as Promise<FixedAssetDepreciation[]>;
};

export const purchaseFixedAsset = async (payload: AssetPurchasedPayload): Promise<void> => {
  const version = await getAggregateVersion('TRANSACTION', '');
  await appendEvent<AssetPurchasedPayload>({
    aggregate_type:   'TRANSACTION',
    aggregate_id:     '',
    expected_version: version,
    event_type:       'asset.purchased',
    payload,
    metadata:         {},
  });
};

export const purchaseFixedAssetWithInstallment = async (payload: AssetPurchasedWithInstallmentPayload): Promise<void> => {
  const version = await getAggregateVersion('TRANSACTION', '');
  await appendEvent<AssetPurchasedWithInstallmentPayload>({
    aggregate_type:   'TRANSACTION',
    aggregate_id:     '',
    expected_version: version,
    event_type:       'asset.purchased_with_installment',
    payload,
    metadata:         {},
  });
};

export const depreciateFixedAsset = async (payload: AssetDepreciatedPayload): Promise<void> => {
  const version = await getAggregateVersion('TRANSACTION', '');
  await appendEvent<AssetDepreciatedPayload>({
    aggregate_type:   'TRANSACTION',
    aggregate_id:     '',
    expected_version: version,
    event_type:       'asset.depreciated',
    payload,
    metadata:         {},
  });
};

export const disposeFixedAsset = async (payload: AssetDisposedPayload): Promise<void> => {
  const version = await getAggregateVersion('TRANSACTION', '');
  await appendEvent<AssetDisposedPayload>({
    aggregate_type:   'TRANSACTION',
    aggregate_id:     '',
    expected_version: version,
    event_type:       'asset.disposed',
    payload,
    metadata:         {},
  });
};
