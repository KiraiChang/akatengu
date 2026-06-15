<script lang="ts">
  import {
    getPdfTemplates, getPdfTemplateLedgers, createPdfTemplate, updatePdfTemplate, deactivatePdfTemplate,
  } from '../api/bankStatement';
  import { getLedgerAccountAll } from '../api/ledger';
  import type { BankPdfTemplate, CreateBankPdfTemplateRequest, BankPdfTemplateLedgerItem } from '../types/bankStatement';
  import type { LedgerAccount } from '../types/ledger';

  const BANK_TYPES   = ['SINOPAC', 'TBD'] as const;
  const ACCOUNT_TYPES = ['BANK_ACCOUNT', 'CREDIT_CARD', 'LOAN'] as const;

  interface LedgerRow {
    tpl_ledger_uuid?: string;
    ledger_uuid:      string;
    account_type:     string;
  }

  interface PdfTemplateForm {
    template_name: string;
    bank_type:     string;
    ledgers:       LedgerRow[];
  }

  function emptyForm(): PdfTemplateForm {
    return { template_name: '', bank_type: 'SINOPAC', ledgers: [{ ledger_uuid: '', account_type: 'BANK_ACCOUNT' }] };
  }

  let templates        = $state<BankPdfTemplate[]>([]);
  let ledgers          = $state<LedgerAccount[]>([]);
  let isLoading        = $state(false);
  let error            = $state('');

  let showModal        = $state(false);
  let modalMode        = $state<'new' | 'edit'>('new');
  let editTarget       = $state<BankPdfTemplate | null>(null);
  let form             = $state<PdfTemplateForm>(emptyForm());
  let isSaving         = $state(false);
  let saveError        = $state('');
  let isLoadingLedgers = $state(false);

  let deactivateTarget = $state<BankPdfTemplate | null>(null);
  let isDeactivating   = $state(false);
  let deactivateError  = $state('');

  $effect(() => { void init(); });

  async function init(): Promise<void> {
    isLoading = true; error = '';
    try {
      [templates, ledgers] = await Promise.all([getPdfTemplates(), getLedgerAccountAll()]);
    } catch (e) {
      error = e instanceof Error ? e.message : '載入失敗';
    } finally {
      isLoading = false;
    }
  }

  const activeLedgers = $derived(ledgers.filter(l => l.is_active));

  function openNew(): void {
    modalMode = 'new'; editTarget = null; form = emptyForm(); saveError = ''; showModal = true;
  }

  async function openEdit(t: BankPdfTemplate): Promise<void> {
    modalMode = 'edit'; editTarget = t;
    form = { template_name: t.template_name, bank_type: t.bank_type, ledgers: [] };
    saveError = ''; isLoadingLedgers = true; showModal = true;
    try {
      const existing = await getPdfTemplateLedgers(t.template_uuid);
      const sorted = [...existing].sort((a, b) => a.sort_order - b.sort_order);
      form.ledgers = sorted.length > 0
        ? sorted.map(l => ({ tpl_ledger_uuid: l.tpl_ledger_uuid, ledger_uuid: l.ledger_uuid, account_type: l.account_type }))
        : [{ ledger_uuid: '', account_type: 'BANK_ACCOUNT' }];
    } catch {
      form.ledgers = [{ ledger_uuid: '', account_type: 'BANK_ACCOUNT' }];
    } finally {
      isLoadingLedgers = false;
    }
  }

  function closeModal(): void { showModal = false; }

  function addLedgerRow(): void {
    form.ledgers = [...form.ledgers, { ledger_uuid: '', account_type: 'BANK_ACCOUNT' }];
  }

  function removeLedgerRow(i: number): void {
    form.ledgers = form.ledgers.filter((_, idx) => idx !== i);
  }

  async function handleSave(e: Event): Promise<void> {
    e.preventDefault();
    isSaving = true; saveError = '';
    try {
      const ledgerItems: BankPdfTemplateLedgerItem[] = form.ledgers.map((row, i) => ({
        ...(row.tpl_ledger_uuid ? { tpl_ledger_uuid: row.tpl_ledger_uuid } : {}),
        ledger_uuid:  row.ledger_uuid,
        account_type: row.account_type,
        sort_order:   i,
      }));
      const req: CreateBankPdfTemplateRequest = {
        template_name: form.template_name.trim(),
        bank_type:     form.bank_type,
        ledgers:       ledgerItems,
      };
      if (modalMode === 'new') {
        await createPdfTemplate(req);
      } else if (editTarget) {
        await updatePdfTemplate(editTarget.template_uuid, { ...req, expected_version: editTarget.version });
      }
      closeModal();
      templates = await getPdfTemplates();
    } catch (e) {
      saveError = e instanceof Error ? e.message : '儲存失敗';
    } finally {
      isSaving = false;
    }
  }

  async function handleDeactivate(): Promise<void> {
    if (!deactivateTarget) return;
    isDeactivating = true; deactivateError = '';
    try {
      await deactivatePdfTemplate(deactivateTarget.template_uuid, deactivateTarget.version);
      deactivateTarget = null;
      templates = await getPdfTemplates();
    } catch (e) {
      deactivateError = e instanceof Error ? e.message : '停用失敗';
    } finally {
      isDeactivating = false;
    }
  }

  const isFormValid = $derived(
    form.template_name.trim() !== '' &&
    form.ledgers.length > 0 &&
    form.ledgers.every(r => r.ledger_uuid !== ''),
  );
</script>

<div class="content-header">
  <h1 class="content-title">PDF 範本管理</h1>
  <button class="btn-primary" onclick={openNew}>＋ 新增範本</button>
</div>

{#if error}
  <p class="query-error" role="alert">{error}</p>
{/if}

<section class="section">
  <div class="table-wrap">
    <table class="data-table" aria-label="PDF 範本列表">
      <thead>
        <tr>
          <th>範本名稱</th>
          <th>銀行類型</th>
          <th>狀態</th>
          <th>版本</th>
          <th>更新時間</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        {#if isLoading}
          <tr><td colspan="6" class="table-empty">載入中…</td></tr>
        {:else if templates.length === 0}
          <tr><td colspan="6" class="table-empty">尚無 PDF 範本</td></tr>
        {:else}
          {#each templates as tpl (tpl.template_id)}
            <tr>
              <td>{tpl.template_name}</td>
              <td class="mono" style="font-size:12px">{tpl.bank_type}</td>
              <td>
                <span class="bs-status-badge" class:bs-status-badge--inactive={!tpl.is_active}>
                  {tpl.is_active ? '啟用' : '停用'}
                </span>
              </td>
              <td style="text-align:center;color:#7a8096;font-size:12px">{tpl.version}</td>
              <td class="mono" style="font-size:11px;color:#5c6278">{tpl.updated_at}</td>
              <td style="white-space:nowrap">
                <button
                  class="btn-ghost"
                  style="padding:2px 10px;font-size:11px;"
                  onclick={() => openEdit(tpl)}
                  disabled={!tpl.is_active}
                >編輯</button>
                {#if tpl.is_active}
                  {#if deactivateTarget?.template_id === tpl.template_id}
                    <span style="font-size:11px;color:#c07070;margin-right:4px;">確認停用？</span>
                    <button
                      class="btn-ghost"
                      style="padding:2px 10px;font-size:11px;color:#c07070;border-color:rgba(192,112,112,0.4);"
                      onclick={handleDeactivate}
                      disabled={isDeactivating}
                    >{isDeactivating ? '…' : '確認'}</button>
                    <button
                      class="btn-ghost"
                      style="padding:2px 10px;font-size:11px;margin-left:4px;"
                      onclick={() => { deactivateTarget = null; deactivateError = ''; }}
                      disabled={isDeactivating}
                    >取消</button>
                  {:else}
                    <button
                      class="btn-ghost"
                      style="padding:2px 10px;font-size:11px;margin-left:4px;"
                      onclick={() => { deactivateTarget = tpl; deactivateError = ''; }}
                    >停用</button>
                  {/if}
                {/if}
              </td>
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  </div>
  {#if deactivateError}
    <p class="query-error" style="margin-top:8px;" role="alert">{deactivateError}</p>
  {/if}
</section>

{#if showModal}
  <div class="modal-overlay" role="dialog" aria-modal="true" aria-labelledby="bspdf-tpl-title">
    <div class="modal-panel modal-panel--wide">
      <header class="modal-header">
        <h2 class="modal-title" id="bspdf-tpl-title">
          {modalMode === 'new' ? '新增 PDF 範本' : '編輯 PDF 範本'}
        </h2>
        <button class="modal-close" onclick={closeModal} aria-label="關閉">×</button>
      </header>

      <form class="modal-body" onsubmit={handleSave}>
        {#if saveError}
          <p class="query-error" style="margin-bottom:16px;" role="alert">{saveError}</p>
        {/if}

        <div class="bs-form-grid">
          <div class="form-group">
            <label class="form-label" for="bspdf-name">範本名稱 *</label>
            <input id="bspdf-name" class="form-input" type="text" bind:value={form.template_name} required />
          </div>
          <div class="form-group">
            <label class="form-label" for="bspdf-bank-type">銀行類型 *</label>
            <select id="bspdf-bank-type" class="form-input" bind:value={form.bank_type}>
              {#each BANK_TYPES as bt}
                <option value={bt}>{bt}</option>
              {/each}
            </select>
          </div>
        </div>

        <p class="bs-section-hint">帳本對應</p>

        {#if isLoadingLedgers}
          <p style="font-size:12px;color:#7a8096;margin-bottom:8px;">載入帳本對應中…</p>
        {/if}

        <table class="bs-ledger-table">
          <thead>
            <tr>
              <th>帳戶</th>
              <th>科目類型</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {#each form.ledgers as row, i (i)}
              <tr>
                <td>
                  <select class="form-input" style="font-size:12px" bind:value={row.ledger_uuid} required>
                    <option value="">請選擇…</option>
                    {#each activeLedgers as l (l.ledger_uuid)}
                      <option value={l.ledger_uuid}>{l.institution} {l.name}</option>
                    {/each}
                  </select>
                </td>
                <td>
                  <select class="form-input" style="font-size:12px" bind:value={row.account_type}>
                    {#each ACCOUNT_TYPES as at}
                      <option value={at}>{at}</option>
                    {/each}
                  </select>
                </td>
                <td>
                  <button
                    type="button"
                    class="btn-ghost"
                    style="padding:2px 8px;font-size:11px;color:#c07070;border-color:rgba(192,112,112,0.4);"
                    onclick={() => removeLedgerRow(i)}
                    disabled={form.ledgers.length <= 1}
                    aria-label="移除此行"
                  >－</button>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
        <button
          type="button"
          class="btn-ghost"
          style="padding:2px 12px;font-size:12px;margin-bottom:16px;"
          onclick={addLedgerRow}
        >＋ 新增帳本</button>

        <div class="je-actions">
          <button type="button" class="btn-ghost" onclick={closeModal} disabled={isSaving}>取消</button>
          <button type="submit" class="btn-primary" disabled={isSaving || !isFormValid}>
            {isSaving ? '儲存中…' : (modalMode === 'new' ? '建立範本' : '儲存修改')}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}
