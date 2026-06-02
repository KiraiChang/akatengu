<script lang="ts">
  import { getAllFixedAssetCategories, createFixedAssetCategory, updateFixedAssetCategory, deleteFixedAssetCategory } from '../api/fixedAssetCategory';
  import { getAccountAll } from '../api/account';
  import AccountSelect from './AccountSelect.svelte';
  import type { FixedAssetCategory } from '../types/fixedAssetCategory';
  import type { Account } from '../types/account';

  interface CatForm {
    name:                            string;
    asset_account_id:                string;
    accum_depreciation_account_id:   string;
    depreciation_expense_account_id: string;
  }

  let categories    = $state<FixedAssetCategory[]>([]);
  let accounts      = $state<Account[]>([]);
  let isLoading     = $state(false);
  let error         = $state('');

  let showCreateModal = $state(false);
  let isCreating      = $state(false);
  let createError     = $state('');
  let createForm      = $state<CatForm>(emptyForm());

  let showEditModal = $state(false);
  let isUpdating    = $state(false);
  let editError     = $state('');
  let editingCat    = $state<FixedAssetCategory | null>(null);
  let editForm      = $state<CatForm>(emptyForm());

  let showCloseModal = $state(false);
  let isClosing      = $state(false);
  let closeError     = $state('');
  let closingCat     = $state<FixedAssetCategory | null>(null);

  const isCreateValid = $derived(
    createForm.name.trim()                              !== '' &&
    createForm.asset_account_id                         !== '' &&
    createForm.accum_depreciation_account_id            !== '' &&
    createForm.depreciation_expense_account_id          !== ''
  );

  const isEditValid = $derived(
    editForm.name.trim()                              !== '' &&
    editForm.asset_account_id                         !== '' &&
    editForm.accum_depreciation_account_id            !== '' &&
    editForm.depreciation_expense_account_id          !== ''
  );

  $effect(() => {
    void load();
  });

  function emptyForm(): CatForm {
    return { name: '', asset_account_id: '', accum_depreciation_account_id: '', depreciation_expense_account_id: '' };
  }

  async function load(): Promise<void> {
    isLoading = true;
    error     = '';
    try {
      categories = await getAllFixedAssetCategories();
    } catch (err) {
      error = err instanceof Error ? err.message : '查詢失敗，請稍後再試。';
    } finally {
      isLoading = false;
    }
  }

  async function ensureAccounts(): Promise<void> {
    if (accounts.length === 0) { try { accounts = await getAccountAll(); } catch { /* 非致命 */ } }
  }

  async function openCreateModal(): Promise<void> {
    createForm  = emptyForm();
    createError = '';
    await ensureAccounts();
    showCreateModal = true;
  }

  async function handleCreate(e: Event): Promise<void> {
    e.preventDefault();
    if (!isCreateValid) return;
    isCreating  = true;
    createError = '';
    try {
      await createFixedAssetCategory({
        name:                            createForm.name.trim(),
        asset_account_id:                createForm.asset_account_id,
        accum_depreciation_account_id:   createForm.accum_depreciation_account_id,
        depreciation_expense_account_id: createForm.depreciation_expense_account_id,
      });
      showCreateModal = false;
      await load();
    } catch (err) {
      createError = err instanceof Error ? err.message : '新增失敗，請稍後再試。';
    } finally {
      isCreating = false;
    }
  }

  async function openEditModal(cat: FixedAssetCategory): Promise<void> {
    editingCat = cat;
    editForm   = {
      name:                            cat.name,
      asset_account_id:                cat.asset_account_id,
      accum_depreciation_account_id:   cat.accum_depreciation_account_id,
      depreciation_expense_account_id: cat.depreciation_expense_account_id,
    };
    editError = '';
    await ensureAccounts();
    showEditModal = true;
  }

  async function handleEdit(e: Event): Promise<void> {
    e.preventDefault();
    if (!isEditValid || !editingCat) return;
    isUpdating = true;
    editError  = '';
    try {
      await updateFixedAssetCategory(editingCat.category_uuid, {
        expected_version:                editingCat.version,
        name:                            editForm.name.trim(),
        asset_account_id:                editForm.asset_account_id,
        accum_depreciation_account_id:   editForm.accum_depreciation_account_id,
        depreciation_expense_account_id: editForm.depreciation_expense_account_id,
      });
      showEditModal = false;
      await load();
    } catch (err) {
      editError = err instanceof Error ? err.message : '修改失敗，請稍後再試。';
    } finally {
      isUpdating = false;
    }
  }

  function openCloseModal(cat: FixedAssetCategory): void {
    closingCat = cat;
    closeError = '';
    showCloseModal = true;
  }

  async function handleClose(e: Event): Promise<void> {
    e.preventDefault();
    if (!closingCat) return;
    isClosing  = true;
    closeError = '';
    try {
      await deleteFixedAssetCategory(closingCat.category_uuid, closingCat.version);
      showCloseModal = false;
      await load();
    } catch (err) {
      closeError = err instanceof Error ? err.message : '關閉失敗，請稍後再試。';
    } finally {
      isClosing = false;
    }
  }
</script>

<section class="section">
  <header class="section-header">
    <h2 class="section-title">固定資產類別</h2>
    <div style="display:flex;align-items:center;gap:16px;">
      {#if isLoading}
        <span class="query-loading"><span class="spinner" aria-hidden="true"></span>載入中</span>
      {/if}
      <button class="section-action" onclick={openCreateModal}>＋ 新增類別</button>
    </div>
  </header>

  {#if error}
    <p class="query-error" role="alert">{error}</p>
  {/if}

  <div class="table-wrap fa-cat-table-wrap">
    <table class="data-table" aria-label="固定資產類別">
      <thead>
        <tr>
          <th>名稱</th>
          <th>資產科目</th>
          <th>累計折舊科目</th>
          <th>折舊費用科目</th>
          <th class="hidden md:table-cell">更新者</th>
          <th class="hidden md:table-cell">更新時間</th>
          <th>狀態</th>
        </tr>
      </thead>
      <tbody>
        {#if isLoading && categories.length === 0}
          <tr><td colspan="7" class="table-empty">載入中...</td></tr>
        {:else if categories.length === 0}
          <tr><td colspan="7" class="table-empty">無資料</td></tr>
        {:else}
          {#each categories as cat (cat.id)}
            <tr>
              <td>{cat.name}</td>
              <td class="mono" style="font-size:11px;">{cat.asset_account_id}</td>
              <td class="mono" style="font-size:11px;">{cat.accum_depreciation_account_id}</td>
              <td class="mono" style="font-size:11px;">{cat.depreciation_expense_account_id}</td>
              <td class="hidden md:table-cell">{cat.updated_by ?? '—'}</td>
              <td class="hidden md:table-cell">{cat.updated_at ?? '—'}</td>
              <td>
                <span class="badge {cat.is_active ? 'fa-cat-status-active' : 'fa-cat-status-closed'}">{cat.is_active ? '使用中' : '已關閉'}</span>
                {#if cat.is_active}
                  <button type="button" class="btn-ghost" style="padding:1px 6px;font-size:10px;margin-left:4px;" onclick={() => openEditModal(cat)}>修改</button>
                  <button type="button" class="btn-ghost" style="padding:1px 6px;font-size:10px;margin-left:4px;" onclick={() => openCloseModal(cat)}>關閉</button>
                {/if}
              </td>
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  </div>

  <div class="fa-cat-card-list">
    {#if isLoading && categories.length === 0}
      <div class="table-empty">載入中...</div>
    {:else if categories.length === 0}
      <div class="table-empty">無資料</div>
    {:else}
      {#each categories as cat (cat.id)}
        <div class="fa-cat-card">
          <div class="fa-cat-card-head">
            <span class="fa-cat-card-title">{cat.name}</span>
            <span class="badge {cat.is_active ? 'fa-cat-status-active' : 'fa-cat-status-closed'}">{cat.is_active ? '使用中' : '已關閉'}</span>
          </div>
          <div class="fa-cat-card-accts">
            <span>資產 {cat.asset_account_id}</span>
            <span>累折 {cat.accum_depreciation_account_id}</span>
            <span>折費 {cat.depreciation_expense_account_id}</span>
          </div>
          {#if cat.is_active}
            <div class="fa-cat-card-footer">
              <button type="button" class="btn-ghost" style="padding:1px 6px;font-size:10px;" onclick={() => openEditModal(cat)}>修改</button>
              <button type="button" class="btn-ghost" style="padding:1px 6px;font-size:10px;" onclick={() => openCloseModal(cat)}>關閉</button>
            </div>
          {/if}
        </div>
      {/each}
    {/if}
  </div>
</section>

<!-- ── 新增 Modal ──────────────────────────── -->
{#if showCreateModal}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) showCreateModal = false; }}>
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="fa-cat-create-title">
      <div class="modal-header">
        <h2 class="modal-title" id="fa-cat-create-title">新增固定資產類別</h2>
        <button class="modal-close" onclick={() => { showCreateModal = false; }} aria-label="關閉">×</button>
      </div>
      <form class="modal-body" onsubmit={handleCreate}>
        {#if createError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{createError}</p>
        {/if}
        <div class="form-group">
          <label class="form-label" for="fa-cat-name">名稱 *</label>
          <input id="fa-cat-name" class="form-input" type="text" bind:value={createForm.name} placeholder="例：辦公設備" required />
        </div>
        <div class="form-group">
          <span class="form-label">資產科目 *</span>
          <AccountSelect {accounts} value={createForm.asset_account_id} placeholder="選擇資產科目…" onselect={(id) => { createForm.asset_account_id = id; }} />
        </div>
        <div class="form-group">
          <span class="form-label">累計折舊科目 *</span>
          <AccountSelect {accounts} value={createForm.accum_depreciation_account_id} placeholder="選擇累計折舊科目…" onselect={(id) => { createForm.accum_depreciation_account_id = id; }} />
        </div>
        <div class="form-group">
          <span class="form-label">折舊費用科目 *</span>
          <AccountSelect {accounts} value={createForm.depreciation_expense_account_id} placeholder="選擇折舊費用科目…" onselect={(id) => { createForm.depreciation_expense_account_id = id; }} />
        </div>
        <div class="modal-footer" style="padding:0;margin-top:8px;">
          <button type="button" class="btn-ghost" onclick={() => { showCreateModal = false; }} disabled={isCreating}>取消</button>
          <button type="submit" class="btn-primary" disabled={isCreating || !isCreateValid}>{isCreating ? '建立中…' : '建立類別'}</button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- ── 修改 Modal ──────────────────────────── -->
{#if showEditModal}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) showEditModal = false; }}>
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="fa-cat-edit-title">
      <div class="modal-header">
        <h2 class="modal-title" id="fa-cat-edit-title">修改固定資產類別</h2>
        <button class="modal-close" onclick={() => { showEditModal = false; }} aria-label="關閉">×</button>
      </div>
      <form class="modal-body" onsubmit={handleEdit}>
        {#if editError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{editError}</p>
        {/if}
        <div class="form-group">
          <label class="form-label" for="fa-cat-edit-name">名稱 *</label>
          <input id="fa-cat-edit-name" class="form-input" type="text" bind:value={editForm.name} required />
        </div>
        <div class="form-group">
          <span class="form-label">資產科目 *</span>
          <AccountSelect {accounts} value={editForm.asset_account_id} placeholder="選擇資產科目…" onselect={(id) => { editForm.asset_account_id = id; }} />
        </div>
        <div class="form-group">
          <span class="form-label">累計折舊科目 *</span>
          <AccountSelect {accounts} value={editForm.accum_depreciation_account_id} placeholder="選擇累計折舊科目…" onselect={(id) => { editForm.accum_depreciation_account_id = id; }} />
        </div>
        <div class="form-group">
          <span class="form-label">折舊費用科目 *</span>
          <AccountSelect {accounts} value={editForm.depreciation_expense_account_id} placeholder="選擇折舊費用科目…" onselect={(id) => { editForm.depreciation_expense_account_id = id; }} />
        </div>
        <div class="modal-footer" style="padding:0;margin-top:8px;">
          <button type="button" class="btn-ghost" onclick={() => { showEditModal = false; }} disabled={isUpdating}>取消</button>
          <button type="submit" class="btn-primary" disabled={isUpdating || !isEditValid}>{isUpdating ? '更新中…' : '儲存修改'}</button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- ── 關閉 Modal ──────────────────────────── -->
{#if showCloseModal}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) showCloseModal = false; }}>
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="fa-cat-close-title">
      <div class="modal-header">
        <h2 class="modal-title" id="fa-cat-close-title">關閉固定資產類別</h2>
        <button class="modal-close" onclick={() => { showCloseModal = false; }} aria-label="關閉">×</button>
      </div>
      <form class="modal-body" onsubmit={handleClose}>
        {#if closeError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{closeError}</p>
        {/if}
        <p style="font-size:13px;color:#c0bdb4;margin-bottom:16px;">確定要關閉類別「{closingCat?.name}」？關閉後將無法用於新增資產。</p>
        <div class="modal-footer" style="padding:0;margin-top:8px;">
          <button type="button" class="btn-ghost" onclick={() => { showCloseModal = false; }} disabled={isClosing}>取消</button>
          <button type="submit" class="btn-primary" disabled={isClosing}>{isClosing ? '處理中…' : '確認關閉'}</button>
        </div>
      </form>
    </div>
  </div>
{/if}
