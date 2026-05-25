<script lang="ts">
  import { getTemplates, getTemplate, updateTemplate, deleteTemplate } from '../api/template';
  import { getAccountAll } from '../api/account';
  import { getLedgerAccountAll } from '../api/ledger';
  import AccountSelect from '../components/AccountSelect.svelte';
  import LedgerSelect from '../components/LedgerSelect.svelte';
  import { CASH_FLOW_CATEGORIES, CASH_FLOW_CATEGORY_LABELS } from '../types/account';
  import type { Account, CashFlowCategory } from '../types/account';
  import type { LedgerAccount } from '../types/ledger';
  import type { TransactionTemplate } from '../types/template';

  type StringEditLineField = 'account_id' | 'ledger_id' | 'debit' | 'credit';

  interface EditLine {
    id:                 number;
    account_id:         string;
    ledger_id:          string;
    debit:              string;
    credit:             string;
    cash_flow_category: CashFlowCategory | null;
  }

  let lineSeq = 1;

  function emptyLine(): EditLine {
    return { id: lineSeq++, account_id: '', ledger_id: '', debit: '', credit: '', cash_flow_category: null };
  }

  let templates = $state<TransactionTemplate[]>([]);
  let isLoading = $state(false);
  let error     = $state('');

  let allAccounts = $state<Account[]>([]);
  let allLedgers  = $state<LedgerAccount[]>([]);

  let showEditModal = $state(false);
  let editId        = $state<number | null>(null);
  let editName      = $state('');
  let editDesc      = $state('');
  let editTag       = $state('');
  let editLines     = $state<EditLine[]>([]);
  let detailLoading = $state(false);
  let isSaving      = $state(false);
  let saveError     = $state('');

  let deletingId  = $state<number | null>(null);
  let isDeleting  = $state(false);
  let deleteError = $state('');

  const activeAccounts = $derived(allAccounts.filter(a => a.is_active));
  const activeLedgers  = $derived(allLedgers.filter(l => l.is_active));

  const totalDebit  = $derived(editLines.reduce((s, l) => s + (parseFloat(l.debit)  || 0), 0));
  const totalCredit = $derived(editLines.reduce((s, l) => s + (parseFloat(l.credit) || 0), 0));
  const isBalanced  = $derived(Math.abs(totalDebit - totalCredit) < 0.001 && totalDebit > 0);

  $effect(() => { void load(); });

  async function load(): Promise<void> {
    isLoading = true;
    error     = '';
    try {
      templates = await getTemplates();
    } catch (err) {
      error = err instanceof Error ? err.message : '載入失敗';
    } finally {
      isLoading = false;
    }
  }

  async function openEditModal(tpl: TransactionTemplate): Promise<void> {
    editId        = tpl.id;
    editName      = tpl.name;
    editDesc      = tpl.description ?? '';
    editTag       = tpl.tag ?? '';
    editLines     = [];
    saveError     = '';
    detailLoading = true;
    showEditModal = true;

    if (allAccounts.length === 0) {
      try {
        [allAccounts, allLedgers] = await Promise.all([getAccountAll(), getLedgerAccountAll()]);
      } catch { /* use empty lists */ }
    }

    try {
      const detail = await getTemplate(tpl.id);
      editLines = detail.entries.length > 0
        ? detail.entries.map(e => ({
            id:                 lineSeq++,
            account_id:         e.account_id,
            ledger_id:          e.ledger_id ? String(e.ledger_id) : '',
            debit:              parseFloat(e.debit)  > 0 ? e.debit  : '',
            credit:             parseFloat(e.credit) > 0 ? e.credit : '',
            cash_flow_category: e.cash_flow_category ?? null,
          }))
        : [emptyLine(), emptyLine()];
    } catch {
      editLines = [emptyLine(), emptyLine()];
    } finally {
      detailLoading = false;
    }
  }

  function closeEditModal(): void {
    showEditModal = false;
    editId = null;
  }

  function addLine(): void { editLines = [...editLines, emptyLine()]; }

  function removeLine(id: number): void {
    if (editLines.length <= 2) return;
    editLines = editLines.filter(l => l.id !== id);
  }

  function updateLine(id: number, field: StringEditLineField, value: string): void {
    editLines = editLines.map(l => l.id === id ? { ...l, [field]: value } : l);
  }

  function updateLineCashFlow(id: number, value: CashFlowCategory | null): void {
    editLines = editLines.map(l => l.id === id ? { ...l, cash_flow_category: value } : l);
  }

  function selectLedger(lineId: number, ledgerIdStr: string): void {
    const ledger = activeLedgers.find(l => String(l.ledger_id) === ledgerIdStr);
    const newAccountId = ledgerIdStr === '' ? '' : (ledger?.account_id ?? '');
    const acct = newAccountId ? (allAccounts.find(a => a.account_id === newAccountId) ?? null) : null;
    editLines = editLines.map(l =>
      l.id === lineId
        ? {
            ...l,
            ledger_id:          ledgerIdStr,
            account_id:         ledgerIdStr === '' ? '' : (ledger?.account_id ?? l.account_id),
            cash_flow_category: acct?.cash_flow_category ?? null,
          }
        : l,
    );
  }

  async function handleSave(e: Event): Promise<void> {
    e.preventDefault();
    if (!editId || !editName.trim()) return;
    isSaving  = true;
    saveError = '';
    try {
      const validLines = editLines.filter(l => l.account_id !== '');
      await updateTemplate(editId, {
        name:        editName.trim(),
        description: editDesc.trim() || null,
        tag:         editTag.trim()  || null,
        entries: validLines.map((l, i) => ({
          sort_order:         i + 1,
          account_id:         l.account_id,
          ledger_id:          l.ledger_id ? parseInt(l.ledger_id, 10) : null,
          debit:              parseFloat(l.debit)  || 0,
          credit:             parseFloat(l.credit) || 0,
          note:               null,
          cash_flow_category: l.cash_flow_category ?? null,
        })),
      });
      closeEditModal();
      await load();
    } catch (err) {
      saveError = err instanceof Error ? err.message : '儲存失敗';
    } finally {
      isSaving = false;
    }
  }

  async function handleDelete(id: number): Promise<void> {
    isDeleting  = true;
    deleteError = '';
    try {
      await deleteTemplate(id);
      deletingId = null;
      await load();
    } catch (err) {
      deleteError = err instanceof Error ? err.message : '刪除失敗';
    } finally {
      isDeleting = false;
    }
  }

  function fmtNum(n: number): string {
    return n.toLocaleString('zh-TW', { minimumFractionDigits: 2 });
  }
</script>

<div class="content-header">
  <h1 class="content-title">範本管理</h1>
  <span class="content-date">共 {templates.length} 個範本</span>
</div>

{#if error}
  <p class="query-error" role="alert">{error}</p>
{/if}

<section class="section">
  <div class="table-wrap">
    <table class="data-table" aria-label="分錄範本列表">
      <thead>
        <tr>
          <th>範本名稱</th>
          <th>說明</th>
          <th>標籤</th>
          <th>更新者</th>
          <th>更新時間</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        {#if isLoading}
          <tr><td colspan="6" class="table-empty">載入中…</td></tr>
        {:else if templates.length === 0}
          <tr><td colspan="6" class="table-empty">尚無範本；可在傳票管理頁面展開分錄後點擊「儲存為範本」建立</td></tr>
        {:else}
          {#each templates as tpl (tpl.id)}
            <tr>
              <td>{tpl.name}</td>
              <td style="color:#7a8096">{tpl.description ?? '—'}</td>
              <td>
                {#if tpl.tag}
                  <span class="tpl-tag">{tpl.tag}</span>
                {:else}
                  —
                {/if}
              </td>
              <td class="mono" style="font-size:11px;color:#5c6278">{tpl.updated_by ?? '—'}</td>
              <td class="mono" style="font-size:11px;color:#5c6278">{tpl.updated_at ?? '—'}</td>
              <td style="white-space:nowrap">
                {#if deletingId === tpl.id}
                  <span style="font-size:11px;color:#c07070;letter-spacing:0.04em;margin-right:6px;">確認刪除？</span>
                  <button
                    class="btn-ghost"
                    style="padding:2px 10px;font-size:11px;color:#c07070;border-color:rgba(192,112,112,0.4);"
                    onclick={() => handleDelete(tpl.id)}
                    disabled={isDeleting}
                  >{isDeleting ? '…' : '確認'}</button>
                  <button
                    class="btn-ghost"
                    style="padding:2px 10px;font-size:11px;margin-left:4px;"
                    onclick={() => { deletingId = null; deleteError = ''; }}
                    disabled={isDeleting}
                  >取消</button>
                {:else}
                  <button
                    class="btn-ghost"
                    style="padding:2px 10px;font-size:11px;"
                    onclick={() => openEditModal(tpl)}
                  >編輯</button>
                  <button
                    class="btn-ghost"
                    style="padding:2px 10px;font-size:11px;margin-left:4px;"
                    onclick={() => { deletingId = tpl.id; deleteError = ''; }}
                  >刪除</button>
                {/if}
              </td>
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  </div>
  {#if deleteError}
    <p class="query-error" style="margin-top:8px;" role="alert">{deleteError}</p>
  {/if}
</section>

{#if showEditModal}
  <div class="modal-overlay" role="dialog" aria-modal="true">
    <div class="modal-panel modal-panel--wide">
      <header class="modal-header">
        <h2 class="modal-title">編輯範本</h2>
        <button class="modal-close" onclick={closeEditModal} aria-label="關閉">×</button>
      </header>

      <form class="modal-body" onsubmit={handleSave}>
        {#if saveError}
          <p class="query-error" style="margin-bottom:16px;" role="alert">{saveError}</p>
        {/if}

        <div class="je-form-header--4col" style="margin-bottom:24px;">
          <div class="form-group" style="margin:0">
            <label class="form-label" for="tpl-edit-name">名稱 *</label>
            <input id="tpl-edit-name" class="form-input" type="text" bind:value={editName} required />
          </div>
          <div class="form-group" style="margin:0">
            <label class="form-label" for="tpl-edit-desc">說明</label>
            <input id="tpl-edit-desc" class="form-input" type="text" bind:value={editDesc} placeholder="選填" />
          </div>
          <div class="form-group" style="margin:0">
            <label class="form-label" for="tpl-edit-tag">標籤</label>
            <input id="tpl-edit-tag" class="form-input" type="text" bind:value={editTag} placeholder="選填" />
          </div>
        </div>

        {#if detailLoading}
          <div style="padding:24px;text-align:center;font-size:12px;color:#3d4258;">載入分錄中…</div>
        {:else}
          <div class="je-lines">
            <div class="je-lines-header je-lines-header--6col">
              <span>金融帳戶</span>
              <span>會計科目 *</span>
              <span>現金流量</span>
              <span class="num">借方金額</span>
              <span class="num">貸方金額</span>
              <span></span>
            </div>

            {#each editLines as line (line.id)}
              <div class="je-line je-line--6col">
                <div class="je-line-cell">
                  <LedgerSelect
                    ledgers={activeLedgers}
                    value={line.ledger_id}
                    onselect={(id) => selectLedger(line.id, id)}
                  />
                </div>
                <div class="je-line-cell">
                  <AccountSelect
                    accounts={activeAccounts}
                    value={line.account_id}
                    placeholder="選擇科目…"
                    onselect={(id) => {
                      updateLine(line.id, 'account_id', id);
                      const acct = activeAccounts.find(a => a.account_id === id) ?? null;
                      updateLineCashFlow(line.id, acct?.cash_flow_category ?? null);
                    }}
                  />
                </div>
                <div class="je-line-cell">
                  <select
                    class="je-cf-select"
                    value={line.cash_flow_category ?? ''}
                    onchange={(e) => {
                      const v = (e.target as HTMLSelectElement).value;
                      updateLineCashFlow(line.id, v ? v as CashFlowCategory : null);
                    }}
                  >
                    <option value="">—</option>
                    {#each CASH_FLOW_CATEGORIES as c}
                      <option value={c}>{CASH_FLOW_CATEGORY_LABELS[c]}</option>
                    {/each}
                  </select>
                </div>
                <div class="je-line-cell">
                  <input
                    class="je-amount-input"
                    type="number"
                    min="0"
                    step="0.01"
                    placeholder="0.00"
                    value={line.debit}
                    oninput={(e) => {
                      updateLine(line.id, 'debit', (e.target as HTMLInputElement).value);
                      if ((e.target as HTMLInputElement).value) updateLine(line.id, 'credit', '');
                    }}
                  />
                </div>
                <div class="je-line-cell">
                  <input
                    class="je-amount-input"
                    type="number"
                    min="0"
                    step="0.01"
                    placeholder="0.00"
                    value={line.credit}
                    oninput={(e) => {
                      updateLine(line.id, 'credit', (e.target as HTMLInputElement).value);
                      if ((e.target as HTMLInputElement).value) updateLine(line.id, 'debit', '');
                    }}
                  />
                </div>
                <div class="je-line-cell" style="padding:6px 8px;">
                  <button
                    type="button"
                    class="je-remove-btn"
                    onclick={() => removeLine(line.id)}
                    disabled={editLines.length <= 2}
                    aria-label="移除此行"
                  >×</button>
                </div>
              </div>
            {/each}

            <div class="je-totals">
              <div class="je-total-item">
                <span class="je-total-label">借方合計</span>
                <span class="je-total-value {isBalanced ? 'balanced' : totalDebit > 0 ? 'unbalanced' : ''}">
                  {fmtNum(totalDebit)}
                </span>
              </div>
              <div class="je-total-item">
                <span class="je-total-label">貸方合計</span>
                <span class="je-total-value {isBalanced ? 'balanced' : totalCredit > 0 ? 'unbalanced' : ''}">
                  {fmtNum(totalCredit)}
                </span>
              </div>
            </div>
          </div>

          <button type="button" class="je-add-line-btn" onclick={addLine}>＋ 新增一行</button>
        {/if}

        <div class="je-actions">
          {#if !isBalanced && (totalDebit > 0 || totalCredit > 0)}
            <span class="je-balance-hint">借貸不平衡，差額：{fmtNum(Math.abs(totalDebit - totalCredit))}</span>
          {/if}
          <button type="button" class="btn-ghost" onclick={closeEditModal} disabled={isSaving}>取消</button>
          <button
            type="submit"
            class="btn-primary"
            disabled={isSaving || !editName.trim() || detailLoading}
          >
            {isSaving ? '儲存中…' : '儲存修改'}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}
