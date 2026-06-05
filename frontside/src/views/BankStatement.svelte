<script lang="ts">
  import { getImportsPaged, importCSV, importExcel, getTemplates } from '../api/bankStatement';
  import { getLedgerAccountAll } from '../api/ledger';
  import type { BankStatementImport, BankCsvTemplate } from '../types/bankStatement';
  import type { LedgerAccount } from '../types/ledger';

  let imports    = $state<BankStatementImport[]>([]);
  let total      = $state(0);
  let page       = $state(1);
  const pageSize = 20;
  let isLoading  = $state(false);
  let error      = $state('');

  let filterLedgerId = $state('0');

  let ledgers   = $state<LedgerAccount[]>([]);
  let templates = $state<BankCsvTemplate[]>([]);

  let showUpload   = $state(false);
  let fUploadType  = $state<'csv' | 'xlsx'>('csv');
  let fLedgerId    = $state('');
  let fDate        = $state('');
  let fTemplateId  = $state('');
  let fFile        = $state<File | null>(null);
  let fNote        = $state('');
  let fPassword    = $state('');
  let isUploading  = $state(false);
  let uploadError  = $state('');

  $effect(() => { void init(); });

  async function init(): Promise<void> {
    try {
      [ledgers, templates] = await Promise.all([getLedgerAccountAll(), getTemplates()]);
    } catch { /* non-fatal, lists stay empty */ }
    await load();
  }

  async function load(): Promise<void> {
    isLoading = true; error = '';
    try {
      const res = await getImportsPaged(parseInt(filterLedgerId, 10) || 0, page, pageSize);
      imports = res.data;
      total   = res.total;
    } catch (e) {
      error = e instanceof Error ? e.message : '載入失敗';
    } finally {
      isLoading = false;
    }
  }

  function openUpload(type: 'csv' | 'xlsx'): void {
    fUploadType = type;
    fLedgerId = ''; fDate = ''; fTemplateId = ''; fFile = null; fNote = ''; fPassword = '';
    uploadError = ''; showUpload = true;
  }

  function closeUpload(): void { showUpload = false; fPassword = ''; }

  function handleFileChange(e: Event): void {
    fFile = (e.target as HTMLInputElement).files?.[0] ?? null;
  }

  async function handleUpload(e: Event): Promise<void> {
    e.preventDefault();
    if (!fFile) { uploadError = `請選擇${fUploadType === 'xlsx' ? 'Excel' : 'CSV'}檔案`; return; }
    isUploading = true; uploadError = '';
    try {
      const fd = new FormData();
      fd.append('ledger_id',      fLedgerId);
      fd.append('statement_date', fDate);
      fd.append('template_id',    fTemplateId);
      fd.append('file',           fFile);
      if (fNote.trim()) fd.append('note', fNote.trim());
      if (fUploadType === 'xlsx') {
        if (fPassword.trim()) fd.append('password', fPassword.trim());
        await importExcel(fd);
      } else {
        await importCSV(fd);
      }
      closeUpload();
      await load();
    } catch (e) {
      uploadError = e instanceof Error ? e.message : '上傳失敗';
    } finally {
      isUploading = false;
    }
  }

  function goToResult(imp: BankStatementImport): void {
    window.location.hash = `#/home/bank-statement/${imp.import_id}/result`;
  }

  const activeLedgers   = $derived(ledgers.filter(l => l.is_active));
  const activeTemplates = $derived(templates.filter(t => t.is_active));
  const totalPages      = $derived(Math.ceil(total / pageSize));

  function fmtLedger(ledgerId: number): string {
    const l = ledgers.find(l => l.ledger_id === ledgerId);
    return l ? `${l.institution} ${l.name}` : String(ledgerId);
  }
</script>

<div class="content-header">
  <h1 class="content-title">對帳單匯入</h1>
  <div style="display:flex;gap:8px">
    <button class="btn-ghost" onclick={() => openUpload('csv')}>＋ 上傳 CSV</button>
    <button class="btn-primary" onclick={() => openUpload('xlsx')}>＋ 上傳 Excel</button>
  </div>
</div>

{#if error}
  <p class="query-error" role="alert">{error}</p>
{/if}

<section class="section">
  <div class="bs-filter-row">
    <label class="form-label" for="bs-filter-ledger" style="margin:0;white-space:nowrap">篩選帳戶</label>
    <select
      id="bs-filter-ledger"
      class="form-input bs-filter-select"
      bind:value={filterLedgerId}
      onchange={() => { page = 1; void load(); }}
    >
      <option value="0">全部帳戶</option>
      {#each activeLedgers as l (l.ledger_id)}
        <option value={String(l.ledger_id)}>{l.institution} {l.name}</option>
      {/each}
    </select>
  </div>

  <div class="table-wrap">
    <table class="data-table" aria-label="對帳單匯入清單">
      <thead>
        <tr>
          <th>帳戶</th>
          <th>對帳日期</th>
          <th>來源檔名</th>
          <th>狀態</th>
          <th>備註</th>
          <th>更新時間</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        {#if isLoading}
          <tr><td colspan="7" class="table-empty">載入中…</td></tr>
        {:else if imports.length === 0}
          <tr><td colspan="7" class="table-empty">尚無匯入紀錄</td></tr>
        {:else}
          {#each imports as imp (imp.import_id)}
            <tr>
              <td>{fmtLedger(imp.ledger_id)}</td>
              <td class="mono" style="font-size:12px">{imp.statement_date}</td>
              <td style="color:#7a8096;font-size:12px">{imp.filename ?? '—'}</td>
              <td>
                <span class="bs-import-status bs-import-status--{imp.status}">
                  {imp.status}
                </span>
              </td>
              <td style="color:#7a8096;font-size:12px">{imp.note ?? '—'}</td>
              <td class="mono" style="font-size:11px;color:#5c6278">{imp.updated_at}</td>
              <td>
                <button
                  class="btn-ghost"
                  style="padding:2px 10px;font-size:11px;"
                  onclick={() => goToResult(imp)}
                >查看明細</button>
              </td>
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  </div>

  {#if totalPages > 1}
    <div class="bs-pagination">
      <button
        class="btn-ghost"
        style="padding:2px 10px;font-size:12px;"
        disabled={page <= 1}
        onclick={() => { page -= 1; void load(); }}
      >‹ 上一頁</button>
      <span style="font-size:12px;color:#7a8096">{page} / {totalPages}</span>
      <button
        class="btn-ghost"
        style="padding:2px 10px;font-size:12px;"
        disabled={page >= totalPages}
        onclick={() => { page += 1; void load(); }}
      >下一頁 ›</button>
    </div>
  {/if}
</section>

{#if showUpload}
  <div class="modal-overlay" role="dialog" aria-modal="true" aria-labelledby="bs-upload-title">
    <div class="modal-panel">
      <header class="modal-header">
        <h2 class="modal-title" id="bs-upload-title">{fUploadType === 'xlsx' ? '上傳 Excel 對帳單' : '上傳 CSV 對帳單'}</h2>
        <button class="modal-close" onclick={closeUpload} aria-label="關閉">×</button>
      </header>

      <form class="modal-body" onsubmit={handleUpload}>
        {#if uploadError}
          <p class="query-error" style="margin-bottom:16px;" role="alert">{uploadError}</p>
        {/if}

        <div class="form-group">
          <label class="form-label" for="bs-up-ledger">帳戶 *</label>
          <select id="bs-up-ledger" class="form-input" bind:value={fLedgerId} required>
            <option value="">請選擇…</option>
            {#each activeLedgers as l (l.ledger_id)}
              <option value={String(l.ledger_id)}>{l.institution} {l.name}</option>
            {/each}
          </select>
        </div>

        <div class="form-group">
          <label class="form-label" for="bs-up-date">對帳日期 *</label>
          <input id="bs-up-date" class="form-input" type="date" bind:value={fDate} required />
        </div>

        <div class="form-group">
          <label class="form-label" for="bs-up-tmpl">欄位範本 *</label>
          <select id="bs-up-tmpl" class="form-input" bind:value={fTemplateId} required>
            <option value="">請選擇…</option>
            {#each activeTemplates as t (t.template_id)}
              <option value={String(t.template_id)}>{t.template_name}</option>
            {/each}
          </select>
        </div>

        <div class="form-group">
          <label class="form-label" for="bs-up-file">{fUploadType === 'xlsx' ? 'Excel 檔案' : 'CSV 檔案'} *</label>
          <input
            id="bs-up-file"
            class="form-input"
            type="file"
            accept={fUploadType === 'xlsx' ? '.xlsx,.xls' : '.csv'}
            onchange={handleFileChange}
            required
          />
        </div>

        {#if fUploadType === 'xlsx'}
          <div class="form-group">
            <label class="form-label" for="bs-up-pw">Excel 密碼（選填）</label>
            <input id="bs-up-pw" class="form-input" type="password" placeholder="加密檔案才需填入" bind:value={fPassword} />
          </div>
        {/if}

        <div class="form-group">
          <label class="form-label" for="bs-up-note">備註</label>
          <input id="bs-up-note" class="form-input" type="text" placeholder="選填" bind:value={fNote} />
        </div>

        <div class="je-actions">
          <button type="button" class="btn-ghost" onclick={closeUpload} disabled={isUploading}>取消</button>
          <button type="submit" class="btn-primary" disabled={isUploading}>
            {isUploading ? '上傳中…' : '上傳'}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}
