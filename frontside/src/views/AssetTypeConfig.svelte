<script lang="ts">
  import { getAssetTypeConfigs, updateAssetTypeConfig } from '../api/setting';
  import { getAccountAll } from '../api/account';
  import AccountSelect from '../components/AccountSelect.svelte';
  import type { AssetTypeAccountConfigResult, AssetType, UpdateAssetTypePayload } from '../types/setting';
  import type { Account } from '../types/account';

  const ASSET_TYPE_LABELS: Record<AssetType, string> = {
    STOCK: '股票',
    FUND:  '基金',
    GOLD:  '黃金',
    FX:    '外匯',
  };
  const ALL_ASSET_TYPES: AssetType[] = ['STOCK', 'FUND', 'GOLD', 'FX'];

  interface EditForm {
    account_id:                  string;
    realized_gain_account_id:    string;
    realized_loss_account_id:    string;
    unrealized_gain_account_id:  string;
    unrealized_loss_account_id:  string;
    oci_account_id:              string;
    fee_account_id:              string;
    tax_account_id:              string;
  }

  function emptyEditForm(): EditForm {
    return {
      account_id:                  '',
      realized_gain_account_id:    '',
      realized_loss_account_id:    '',
      unrealized_gain_account_id:  '',
      unrealized_loss_account_id:  '',
      oci_account_id:              '',
      fee_account_id:              '',
      tax_account_id:              '',
    };
  }

  let configs   = $state<AssetTypeAccountConfigResult[]>([]);
  let accounts  = $state<Account[]>([]);
  let isLoading = $state(false);
  let error     = $state('');

  let showModal    = $state(false);
  let isSaving     = $state(false);
  let saveError    = $state('');
  let editingType  = $state<AssetType>('STOCK');
  let editForm     = $state<EditForm>(emptyEditForm());

  const configMap  = $derived(new Map(configs.map(c => [c.asset_type, c])));
  const accountMap = $derived(new Map(accounts.map(a => [a.account_id, a])));

  const canSubmit = $derived(
    !!editForm.realized_gain_account_id &&
    !!editForm.realized_loss_account_id &&
    !!editForm.unrealized_gain_account_id &&
    !!editForm.unrealized_loss_account_id &&
    !!editForm.fee_account_id &&
    !!editForm.tax_account_id
  );

  $effect(() => {
    void load();
  });

  async function load(): Promise<void> {
    isLoading = true;
    error     = '';
    try {
      [configs, accounts] = await Promise.all([getAssetTypeConfigs(), getAccountAll()]);
    } catch (err) {
      error = err instanceof Error ? err.message : '查詢失敗，請稍後再試。';
    } finally {
      isLoading = false;
    }
  }

  function openEditModal(type: AssetType): void {
    editingType = type;
    const cfg   = configMap.get(type);
    editForm = {
      account_id:                  cfg?.account_id                  ?? '',
      realized_gain_account_id:    cfg?.realized_gain_account_id    ?? '',
      realized_loss_account_id:    cfg?.realized_loss_account_id    ?? '',
      unrealized_gain_account_id:  cfg?.unrealized_gain_account_id  ?? '',
      unrealized_loss_account_id:  cfg?.unrealized_loss_account_id  ?? '',
      oci_account_id:              cfg?.oci_account_id              ?? '',
      fee_account_id:              cfg?.fee_account_id              ?? '',
      tax_account_id:              cfg?.tax_account_id              ?? '',
    };
    saveError = '';
    showModal = true;
  }

  function closeModal(): void {
    showModal = false;
  }

  async function handleSubmit(e: Event): Promise<void> {
    e.preventDefault();
    if (!canSubmit) return;
    isSaving  = true;
    saveError = '';
    try {
      const payload: UpdateAssetTypePayload = {
        account_id:                  editForm.account_id || null,
        realized_gain_account_id:    editForm.realized_gain_account_id,
        realized_loss_account_id:    editForm.realized_loss_account_id,
        unrealized_gain_account_id:  editForm.unrealized_gain_account_id,
        unrealized_loss_account_id:  editForm.unrealized_loss_account_id,
        oci_account_id:              editForm.oci_account_id || null,
        fee_account_id:              editForm.fee_account_id,
        tax_account_id:              editForm.tax_account_id,
      };
      await updateAssetTypeConfig(editingType, payload);
      showModal = false;
      await load();
    } catch (err) {
      saveError = err instanceof Error ? err.message : '更新失敗';
    } finally {
      isSaving = false;
    }
  }

  function acctName(id: string | null | undefined): string {
    if (!id) return '—';
    return accountMap.get(id)?.name ?? id;
  }
</script>

<div class="content-header">
  <h1 class="content-title">資產類型設定</h1>
  <span class="content-date">共 {ALL_ASSET_TYPES.length} 種類型</span>
</div>

{#if error}
  <p class="query-error" role="alert">{error}</p>
{/if}

<section class="section">
  <header class="section-header">
    <h2 class="section-title">資產類型科目對應</h2>
    {#if isLoading}
      <span class="query-loading">
        <span class="spinner" aria-hidden="true"></span>
        載入中
      </span>
    {/if}
  </header>

  <div class="table-wrap">
    <table class="data-table" aria-label="資產類型科目對應">
      <thead>
        <tr>
          <th>資產類型</th>
          <th>投資主科目</th>
          <th>已實現利得</th>
          <th>已實現損失</th>
          <th>未實現利得</th>
          <th>未實現損失</th>
          <th>OCI</th>
          <th>手續費</th>
          <th>稅費</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        {#if isLoading && configs.length === 0}
          <tr><td colspan="10" class="table-empty">載入中...</td></tr>
        {:else}
          {#each ALL_ASSET_TYPES as type (type)}
            {@const cfg = configMap.get(type)}
            <tr>
              <td>{ASSET_TYPE_LABELS[type]}</td>
              <td class="mono" style="font-size:11px;" title={cfg?.account_id ?? ''}>{acctName(cfg?.account_id)}</td>
              <td class="mono" style="font-size:11px;">{cfg?.realized_gain_account_id || '—'}</td>
              <td class="mono" style="font-size:11px;">{cfg?.realized_loss_account_id || '—'}</td>
              <td class="mono" style="font-size:11px;">{cfg?.unrealized_gain_account_id || '—'}</td>
              <td class="mono" style="font-size:11px;">{cfg?.unrealized_loss_account_id || '—'}</td>
              <td class="mono" style="font-size:11px;">{cfg?.oci_account_id || '—'}</td>
              <td class="mono" style="font-size:11px;">{cfg?.fee_account_id || '—'}</td>
              <td class="mono" style="font-size:11px;">{cfg?.tax_account_id || '—'}</td>
              <td>
                <button
                  class="btn-ghost"
                  style="padding:2px 10px;font-size:11px;"
                  onclick={() => openEditModal(type)}
                >編輯</button>
              </td>
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  </div>
</section>

{#if showModal}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) closeModal(); }}>
    <div class="modal modal--wide" role="dialog" aria-modal="true" aria-labelledby="asset-type-modal-title">
      <div class="modal-header">
        <h2 class="modal-title" id="asset-type-modal-title">
          修改資產類型科目對應 — {ASSET_TYPE_LABELS[editingType]}
        </h2>
        <button class="modal-close" onclick={closeModal} aria-label="關閉">×</button>
      </div>

      <form class="modal-body" onsubmit={handleSubmit}>
        {#if saveError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{saveError}</p>
        {/if}

        <div class="form-row">
          <div class="form-group">
            <label class="form-label" for="at-inv-account">投資主科目（選填）</label>
            <AccountSelect
              {accounts}
              value={editForm.account_id}
              placeholder="選填，用於過濾新增投資的關聯科目…"
              onselect={(id) => { editForm.account_id = id; }}
            />
          </div>
          <div class="form-group"></div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label class="form-label" for="at-realized-gain">已實現利得科目 *</label>
            <AccountSelect
              {accounts}
              value={editForm.realized_gain_account_id}
              placeholder="選擇會計科目…"
              onselect={(id) => { editForm.realized_gain_account_id = id; }}
            />
          </div>
          <div class="form-group">
            <label class="form-label" for="at-realized-loss">已實現損失科目 *</label>
            <AccountSelect
              {accounts}
              value={editForm.realized_loss_account_id}
              placeholder="選擇會計科目…"
              onselect={(id) => { editForm.realized_loss_account_id = id; }}
            />
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label class="form-label" for="at-unrealized-gain">未實現利得科目 *</label>
            <AccountSelect
              {accounts}
              value={editForm.unrealized_gain_account_id}
              placeholder="選擇會計科目…"
              onselect={(id) => { editForm.unrealized_gain_account_id = id; }}
            />
          </div>
          <div class="form-group">
            <label class="form-label" for="at-unrealized-loss">未實現損失科目 *</label>
            <AccountSelect
              {accounts}
              value={editForm.unrealized_loss_account_id}
              placeholder="選擇會計科目…"
              onselect={(id) => { editForm.unrealized_loss_account_id = id; }}
            />
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label class="form-label" for="at-oci">OCI 科目（選填）</label>
            <AccountSelect
              {accounts}
              value={editForm.oci_account_id}
              placeholder="選填"
              onselect={(id) => { editForm.oci_account_id = id; }}
            />
          </div>
          <div class="form-group">
            <label class="form-label" for="at-fee">手續費科目 *</label>
            <AccountSelect
              {accounts}
              value={editForm.fee_account_id}
              placeholder="選擇會計科目…"
              onselect={(id) => { editForm.fee_account_id = id; }}
            />
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label class="form-label" for="at-tax">稅費科目 *</label>
            <AccountSelect
              {accounts}
              value={editForm.tax_account_id}
              placeholder="選擇會計科目…"
              onselect={(id) => { editForm.tax_account_id = id; }}
            />
          </div>
          <div class="form-group"></div>
        </div>

        <div class="modal-footer" style="padding:0;margin-top:8px;">
          <button type="button" class="btn-ghost" onclick={closeModal} disabled={isSaving}>取消</button>
          <button
            type="submit"
            class="btn-primary"
            disabled={isSaving || !canSubmit}
          >
            {isSaving ? '儲存中…' : '儲存修改'}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}
