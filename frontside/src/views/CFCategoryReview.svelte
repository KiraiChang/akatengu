<script lang="ts">
  import { getCFReview, updateCFCategory } from '../api/cfCategory';
  import { getAccountAll } from '../api/account';
  import { CASH_FLOW_CATEGORY_LABELS } from '../types/account';
  import type { Account } from '../types/account';
  import type { CFReviewRow, CFReviewTxn, CFActivityCategory } from '../types/cfCategory';

  const CF_ACTIVITY_OPTIONS: { value: CFActivityCategory; label: string }[] = [
    { value: 'OPERATING', label: CASH_FLOW_CATEGORY_LABELS['OPERATING'] },
    { value: 'INVESTING',  label: CASH_FLOW_CATEGORY_LABELS['INVESTING'] },
    { value: 'FINANCING',  label: CASH_FLOW_CATEGORY_LABELS['FINANCING'] },
  ];

  const today = new Date().toISOString().slice(0, 10);
  const monthStart = `${today.slice(0, 7)}-01`;

  let allAccounts         = $state<Account[]>([]);
  let rows                = $state<CFReviewRow[]>([]);
  let isLoading           = $state(false);
  let error               = $state('');
  let filterFrom          = $state(monthStart);
  let filterTo            = $state(today);
  let showOnlyUnconfirmed = $state(true);
  let editState           = $state<Record<string, CFActivityCategory | null>>({});
  let savingTxn           = $state<string | null>(null);
  let saveErrors          = $state<Record<string, string>>({});

  async function load(): Promise<void> {
    isLoading = true;
    error     = '';
    editState = {};
    saveErrors = {};
    try {
      rows = await getCFReview(filterFrom || undefined, filterTo || undefined);
    } catch (e) {
      error = e instanceof Error ? e.message : '載入失敗';
    } finally {
      isLoading = false;
    }
  }

  $effect(() => { void load(); });
  $effect(() => { void getAccountAll().then(a => { allAccounts = a; }).catch(() => {}); });

  const grouped = $derived((): CFReviewTxn[] => {
    const map = new Map<string, CFReviewTxn>();
    for (const row of rows) {
      if (!map.has(row.txn_uuid)) {
        map.set(row.txn_uuid, {
          txn_id:       row.txn_id,
          txn_uuid:     row.txn_uuid,
          txn_date:     row.txn_date,
          description:  row.description,
          total_amount: row.total_amount,
          currency:     row.currency,
          entries:      [],
        });
      }
      map.get(row.txn_uuid)!.entries.push(row);
    }
    let txns = [...map.values()];
    if (showOnlyUnconfirmed) {
      txns = txns.filter(t =>
        t.entries.some(e => !e.is_confirmed && e.cf_category !== 'CASH'),
      );
    }
    return txns;
  });

  function getEditValue(entryUuid: string, original: CFReviewRow['cf_category']): string {
    if (entryUuid in editState) {
      return editState[entryUuid] ?? '';
    }
    if (original === 'CASH' || original === null) return '';
    return original;
  }

  function isDirtyTxn(txn: CFReviewTxn): boolean {
    return txn.entries.some(e => e.entry_uuid in editState);
  }

  function hasPendingNull(txn: CFReviewTxn): boolean {
    return txn.entries.some(e => {
      if (e.cf_category === 'CASH') return false;
      const val = getEditValue(e.entry_uuid, e.cf_category);
      return val === '';
    });
  }

  async function handleSave(txn: CFReviewTxn): Promise<void> {
    if (hasPendingNull(txn)) {
      saveErrors = { ...saveErrors, [txn.txn_uuid]: '所有非現金分錄需設定 CF 分類後才能儲存' };
      return;
    }

    const entriesToSave = txn.entries
      .filter(e => e.cf_category !== 'CASH')
      .map(e => ({
        entry_uuid:  e.entry_uuid,
        cf_category: (getEditValue(e.entry_uuid, e.cf_category) || e.cf_category) as CFActivityCategory,
      }))
      .filter(e => e.cf_category);

    if (entriesToSave.length === 0) return;

    savingTxn  = txn.txn_uuid;
    saveErrors = { ...saveErrors, [txn.txn_uuid]: '' };
    try {
      await updateCFCategory(txn.txn_uuid, entriesToSave);
      const next = { ...editState };
      txn.entries.forEach(e => delete next[e.entry_uuid]);
      editState = next;
      await load();
    } catch (e) {
      saveErrors = { ...saveErrors, [txn.txn_uuid]: e instanceof Error ? e.message : '儲存失敗' };
    } finally {
      savingTxn = null;
    }
  }

  function fmtAmount(s: string): string {
    const n = parseFloat(s);
    if (isNaN(n) || n === 0) return '';
    return n.toLocaleString('zh-TW', { minimumFractionDigits: 2 });
  }

  function cfLabel(cat: string | null): string {
    if (!cat) return '未分類';
    return CASH_FLOW_CATEGORY_LABELS[cat as keyof typeof CASH_FLOW_CATEGORY_LABELS] ?? cat;
  }
</script>

<div class="content-header">
  <h1 class="content-title">CF 分類審視</h1>
</div>

<section class="section">
  <form
    class="cfc-filter-bar"
    onsubmit={(e) => { e.preventDefault(); void load(); }}
  >
    <div class="form-group">
      <label class="form-label" for="cfc-from">開始日期</label>
      <input id="cfc-from" class="form-input" type="date" bind:value={filterFrom} />
    </div>
    <div class="form-group">
      <label class="form-label" for="cfc-to">結束日期</label>
      <input id="cfc-to" class="form-input" type="date" bind:value={filterTo} />
    </div>
    <button type="submit" class="btn-primary" disabled={isLoading}>
      {isLoading ? '查詢中…' : '查詢'}
    </button>
    <label class="cfc-toggle-label">
      <input type="checkbox" bind:checked={showOnlyUnconfirmed} />
      僅顯示未確認
    </label>
  </form>

  {#if error}
    <p class="query-error" role="alert">{error}</p>
  {:else if isLoading}
    <div class="table-empty" style="padding:40px;text-align:center;">載入中…</div>
  {:else if grouped().length === 0}
    <div class="table-empty" style="padding:40px;text-align:center;">
      {showOnlyUnconfirmed ? '所有分錄皆已確認' : '此期間無資料'}
    </div>
  {:else}
    <div class="cfc-txn-list">
      {#each grouped() as txn (txn.txn_uuid)}
        {@const dirty = isDirtyTxn(txn)}
        {@const isSaving = savingTxn === txn.txn_uuid}
        {@const txnError = saveErrors[txn.txn_uuid] ?? ''}

        <div class="cfc-txn-group {dirty ? 'cfc-txn-group--dirty' : ''}">
          <div class="cfc-txn-header">
            <span>{txn.txn_date}</span>
            <span class="cfc-txn-desc">{txn.description}</span>
            <span class="num">{parseFloat(txn.total_amount).toLocaleString('zh-TW', { minimumFractionDigits: 2 })}</span>
            <span class="cfc-txn-currency">{txn.currency}</span>
          </div>

          <div class="cfc-entry-header">
            <span>科目</span>
            <span class="num">借方</span>
            <span class="num">貸方</span>
            <span>CF 分類</span>
            <span>狀態</span>
          </div>

          {#each txn.entries as entry (entry.entry_uuid)}
            {@const isCash = entry.cf_category === 'CASH'}
            {@const account = allAccounts.find(a => a.account_id === entry.account_id) ?? null}
            <div class="cfc-entry-row {isCash ? 'cfc-entry-row--cash' : ''}">
              <span class="cfc-entry-cell">{account?.name ?? entry.account_id}</span>
              <span class="cfc-entry-cell num">{fmtAmount(entry.debit)}</span>
              <span class="cfc-entry-cell num">{fmtAmount(entry.credit)}</span>
              <span class="cfc-entry-cell">
                {#if isCash}
                  <span class="cfc-cf-badge cfc-cf-badge--cash">{cfLabel(entry.cf_category)}</span>
                {:else}
                  {@const currentVal = getEditValue(entry.entry_uuid, entry.cf_category)}
                  {@const entryDirty = entry.entry_uuid in editState}
                  <select
                    class="cfc-cf-select {entryDirty ? 'cfc-dirty' : ''}"
                    value={currentVal}
                    onchange={(e) => {
                      const v = (e.target as HTMLSelectElement).value as CFActivityCategory | '';
                      editState = { ...editState, [entry.entry_uuid]: v || null };
                    }}
                  >
                    <option value="">— 未分類 —</option>
                    {#each CF_ACTIVITY_OPTIONS as opt}
                      <option value={opt.value}>{opt.label}</option>
                    {/each}
                  </select>
                {/if}
              </span>
              <span class="cfc-entry-cell">
                {#if entry.is_confirmed}
                  <span class="cfc-confirmed">已確認</span>
                {:else}
                  <span class="cfc-unconfirmed">待確認</span>
                {/if}
              </span>
            </div>
          {/each}

          <div class="cfc-save-row">
            {#if txnError}
              <span class="cfc-save-error">{txnError}</span>
            {/if}
            <button
              type="button"
              class="btn-primary"
              disabled={isSaving || (!dirty && !hasPendingNull(txn))}
              onclick={() => handleSave(txn)}
            >
              {isSaving ? '儲存中…' : '儲存'}
            </button>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</section>
