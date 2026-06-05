<script lang="ts">
  import { getTemplates, createTemplate, updateTemplate, deactivateTemplate } from '../api/bankStatement';
  import type { BankCsvTemplate, CreateBankCsvTemplateRequest } from '../types/bankStatement';

  interface TemplateForm {
    template_name:      string;
    encoding:           string;
    skip_rows:          string;
    date_column:        string;
    date_format:        string;
    description_column: string;
    debit_column:       string;
    credit_column:      string;
    amount_column:      string;
    balance_column:     string;
    reference_column:   string;
    note:               string;
  }

  function emptyForm(): TemplateForm {
    return {
      template_name: '', encoding: 'UTF-8', skip_rows: '0',
      date_column: '0', date_format: '2006-01-02', description_column: '1',
      debit_column: '', credit_column: '', amount_column: '',
      balance_column: '', reference_column: '', note: '',
    };
  }

  let templates        = $state<BankCsvTemplate[]>([]);
  let isLoading        = $state(false);
  let error            = $state('');

  let showModal        = $state(false);
  let modalMode        = $state<'new' | 'edit'>('new');
  let editTarget       = $state<BankCsvTemplate | null>(null);
  let form             = $state<TemplateForm>(emptyForm());
  let isSaving         = $state(false);
  let saveError        = $state('');

  let deactivateTarget = $state<BankCsvTemplate | null>(null);
  let isDeactivating   = $state(false);
  let deactivateError  = $state('');

  $effect(() => { void load(); });

  async function load(): Promise<void> {
    isLoading = true; error = '';
    try { templates = await getTemplates(); }
    catch (e) { error = e instanceof Error ? e.message : '載入失敗'; }
    finally { isLoading = false; }
  }

  function openNew(): void {
    modalMode = 'new'; editTarget = null; form = emptyForm(); saveError = ''; showModal = true;
  }

  function openEdit(t: BankCsvTemplate): void {
    modalMode = 'edit'; editTarget = t;
    form = {
      template_name:      t.template_name,
      encoding:           t.encoding,
      skip_rows:          String(t.skip_rows),
      date_column:        String(t.date_column),
      date_format:        t.date_format,
      description_column: String(t.description_column),
      debit_column:       t.debit_column     != null ? String(t.debit_column)     : '',
      credit_column:      t.credit_column    != null ? String(t.credit_column)    : '',
      amount_column:      t.amount_column    != null ? String(t.amount_column)    : '',
      balance_column:     t.balance_column   != null ? String(t.balance_column)   : '',
      reference_column:   t.reference_column != null ? String(t.reference_column) : '',
      note:               t.note ?? '',
    };
    saveError = ''; showModal = true;
  }

  function closeModal(): void { showModal = false; }

  function parseOptInt(s: string): number | null {
    const n = parseInt(s, 10);
    return s.trim() === '' || isNaN(n) ? null : n;
  }

  async function handleSave(e: Event): Promise<void> {
    e.preventDefault();
    isSaving = true; saveError = '';
    try {
      const req: CreateBankCsvTemplateRequest = {
        template_name:      form.template_name.trim(),
        encoding:           form.encoding,
        skip_rows:          parseInt(form.skip_rows, 10) || 0,
        date_column:        parseInt(form.date_column, 10) || 0,
        date_format:        form.date_format.trim(),
        description_column: parseInt(form.description_column, 10) || 0,
        debit_column:       parseOptInt(form.debit_column),
        credit_column:      parseOptInt(form.credit_column),
        amount_column:      parseOptInt(form.amount_column),
        balance_column:     parseOptInt(form.balance_column),
        reference_column:   parseOptInt(form.reference_column),
        note:               form.note.trim() || null,
      };
      if (modalMode === 'new') {
        await createTemplate(req);
      } else if (editTarget) {
        await updateTemplate(editTarget.template_uuid, { ...req, expected_version: editTarget.version });
      }
      closeModal();
      await load();
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
      await deactivateTemplate(deactivateTarget.template_uuid, deactivateTarget.version);
      deactivateTarget = null;
      await load();
    } catch (e) {
      deactivateError = e instanceof Error ? e.message : '停用失敗';
    } finally {
      isDeactivating = false;
    }
  }
</script>

<div class="content-header">
  <h1 class="content-title">CSV 範本管理</h1>
  <button class="btn-primary" onclick={openNew}>＋ 新增範本</button>
</div>

{#if error}
  <p class="query-error" role="alert">{error}</p>
{/if}

<section class="section">
  <div class="table-wrap">
    <table class="data-table" aria-label="CSV 範本列表">
      <thead>
        <tr>
          <th>範本名稱</th>
          <th>編碼</th>
          <th>日期格式</th>
          <th>跳過行數</th>
          <th>狀態</th>
          <th>備註</th>
          <th>更新時間</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        {#if isLoading}
          <tr><td colspan="8" class="table-empty">載入中…</td></tr>
        {:else if templates.length === 0}
          <tr><td colspan="8" class="table-empty">尚無 CSV 範本</td></tr>
        {:else}
          {#each templates as tpl (tpl.template_id)}
            <tr>
              <td>{tpl.template_name}</td>
              <td class="mono" style="font-size:12px">{tpl.encoding}</td>
              <td class="mono" style="font-size:12px">{tpl.date_format}</td>
              <td style="text-align:center">{tpl.skip_rows}</td>
              <td>
                <span class="bs-status-badge" class:bs-status-badge--inactive={!tpl.is_active}>
                  {tpl.is_active ? '啟用' : '停用'}
                </span>
              </td>
              <td style="color:#7a8096">{tpl.note ?? '—'}</td>
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
  <div class="modal-overlay" role="dialog" aria-modal="true" aria-labelledby="bs-tpl-title">
    <div class="modal-panel modal-panel--wide">
      <header class="modal-header">
        <h2 class="modal-title" id="bs-tpl-title">
          {modalMode === 'new' ? '新增 CSV 範本' : '編輯 CSV 範本'}
        </h2>
        <button class="modal-close" onclick={closeModal} aria-label="關閉">×</button>
      </header>

      <form class="modal-body" onsubmit={handleSave}>
        {#if saveError}
          <p class="query-error" style="margin-bottom:16px;" role="alert">{saveError}</p>
        {/if}

        <div class="bs-form-grid">
          <div class="form-group">
            <label class="form-label" for="bs-name">範本名稱 *</label>
            <input id="bs-name" class="form-input" type="text" bind:value={form.template_name} required />
          </div>
          <div class="form-group">
            <label class="form-label" for="bs-encoding">編碼</label>
            <select id="bs-encoding" class="form-input" bind:value={form.encoding}>
              <option value="UTF-8">UTF-8</option>
              <option value="UTF-8-BOM">UTF-8 BOM</option>
              <option value="Big5">Big5</option>
              <option value="Shift-JIS">Shift-JIS</option>
              <option value="GBK">GBK</option>
            </select>
          </div>
          <div class="form-group">
            <label class="form-label" for="bs-skip">跳過行數</label>
            <input id="bs-skip" class="form-input" type="text" inputmode="numeric" bind:value={form.skip_rows} />
          </div>
          <div class="form-group">
            <label class="form-label" for="bs-date-fmt">
              日期格式 * <span class="bs-field-hint">（Go 格式，如 2006-01-02）</span>
            </label>
            <input id="bs-date-fmt" class="form-input" type="text" bind:value={form.date_format} required />
          </div>
          <div class="form-group">
            <label class="form-label" for="bs-date-col">日期欄索引（0 起算）</label>
            <input id="bs-date-col" class="form-input" type="text" inputmode="numeric" bind:value={form.date_column} />
          </div>
          <div class="form-group">
            <label class="form-label" for="bs-desc-col">說明欄索引（0 起算）</label>
            <input id="bs-desc-col" class="form-input" type="text" inputmode="numeric" bind:value={form.description_column} />
          </div>
        </div>

        <p class="bs-section-hint">金額欄位：至少填入「借方」「貸方」其中一個，或填入「金額」</p>
        <div class="bs-form-grid">
          <div class="form-group">
            <label class="form-label" for="bs-debit-col">借方欄索引</label>
            <input id="bs-debit-col" class="form-input" type="text" inputmode="numeric" placeholder="選填" bind:value={form.debit_column} />
          </div>
          <div class="form-group">
            <label class="form-label" for="bs-credit-col">貸方欄索引</label>
            <input id="bs-credit-col" class="form-input" type="text" inputmode="numeric" placeholder="選填" bind:value={form.credit_column} />
          </div>
          <div class="form-group">
            <label class="form-label" for="bs-amount-col">金額欄索引</label>
            <input id="bs-amount-col" class="form-input" type="text" inputmode="numeric" placeholder="選填" bind:value={form.amount_column} />
          </div>
          <div class="form-group">
            <label class="form-label" for="bs-bal-col">餘額欄索引</label>
            <input id="bs-bal-col" class="form-input" type="text" inputmode="numeric" placeholder="選填" bind:value={form.balance_column} />
          </div>
          <div class="form-group">
            <label class="form-label" for="bs-ref-col">參考編號欄索引</label>
            <input id="bs-ref-col" class="form-input" type="text" inputmode="numeric" placeholder="選填" bind:value={form.reference_column} />
          </div>
          <div class="form-group">
            <label class="form-label" for="bs-note">備註</label>
            <input id="bs-note" class="form-input" type="text" placeholder="選填" bind:value={form.note} />
          </div>
        </div>

        <div class="je-actions">
          <button type="button" class="btn-ghost" onclick={closeModal} disabled={isSaving}>取消</button>
          <button
            type="submit"
            class="btn-primary"
            disabled={isSaving || !form.template_name.trim() || !form.date_format.trim()}
          >
            {isSaving ? '儲存中…' : (modalMode === 'new' ? '建立範本' : '儲存修改')}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}
