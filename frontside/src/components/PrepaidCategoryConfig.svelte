<script lang="ts">
  import { getAllPrepaidCategories, createPrepaidCategory, updatePrepaidCategory, deletePrepaidCategory } from '../api/prepaidCategory';
  import { getAccountAll } from '../api/account';
  import AccountSelect from './AccountSelect.svelte';
  import type { PrepaidCategory } from '../types/prepaidCategory';
  import type { Account } from '../types/account';

  interface CatForm {
    name:               string;
    account_id:         string;
    expense_account_id: string;
  }

  let categories    = $state<PrepaidCategory[]>([]);
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
  let editingCat    = $state<PrepaidCategory | null>(null);
  let editForm      = $state<CatForm>(emptyForm());

  let showCloseModal = $state(false);
  let isClosing      = $state(false);
  let closeError     = $state('');
  let closingCat     = $state<PrepaidCategory | null>(null);

  const isCreateValid = $derived(
    createForm.name.trim()        !== '' &&
    createForm.account_id         !== '' &&
    createForm.expense_account_id !== ''
  );

  const isEditValid = $derived(
    editForm.name.trim()        !== '' &&
    editForm.account_id         !== '' &&
    editForm.expense_account_id !== ''
  );

  $effect(() => {
    void load();
  });

  function emptyForm(): CatForm {
    return { name: '', account_id: '', expense_account_id: '' };
  }

  async function load(): Promise<void> {
    isLoading = true;
    error     = '';
    try {
      categories = await getAllPrepaidCategories();
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
      await createPrepaidCategory({
        name:               createForm.name.trim(),
        account_id:         createForm.account_id,
        expense_account_id: createForm.expense_account_id,
      });
      showCreateModal = false;
      await load();
    } catch (err) {
      createError = err instanceof Error ? err.message : '新增失敗，請稍後再試。';
    } finally {
      isCreating = false;
    }
  }

  async function openEditModal(cat: PrepaidCategory): Promise<void> {
    editingCat = cat;
    editForm   = {
      name:               cat.name,
      account_id:         cat.account_id,
      expense_account_id: cat.expense_account_id,
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
      await updatePrepaidCategory(editingCat.category_uuid, {
        expected_version:   editingCat.version,
        name:               editForm.name.trim(),
        account_id:         editForm.account_id,
        expense_account_id: editForm.expense_account_id,
      });
      showEditModal = false;
      await load();
    } catch (err) {
      editError = err instanceof Error ? err.message : '修改失敗，請稍後再試。';
    } finally {
      isUpdating = false;
    }
  }

  function openCloseModal(cat: PrepaidCategory): void {
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
      await deletePrepaidCategory(closingCat.category_uuid, closingCat.version);
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
    <h2 class="section-title">預付費用類別</h2>
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

  <div class="table-wrap pp-cat-table-wrap">
    <table class="data-table" aria-label="預付費用類別">
      <thead>
        <tr>
          <th>名稱</th>
          <th>預付費用科目</th>
          <th>費用科目</th>
          <th class="hidden md:table-cell">更新者</th>
          <th class="hidden md:table-cell">更新時間</th>
          <th>狀態</th>
        </tr>
      </thead>
      <tbody>
        {#if isLoading && categories.length === 0}
          <tr><td colspan="6" class="table-empty">載入中...</td></tr>
        {:else if categories.length === 0}
          <tr><td colspan="6" class="table-empty">無資料</td></tr>
        {:else}
          {#each categories as cat (cat.id)}
            <tr>
              <td>{cat.name}</td>
              <td class="mono" style="font-size:11px;">{cat.account_id}</td>
              <td class="mono" style="font-size:11px;">{cat.expense_account_id}</td>
              <td class="hidden md:table-cell">{cat.updated_by ?? '—'}</td>
              <td class="hidden md:table-cell">{cat.updated_at ?? '—'}</td>
              <td>
                <span class="badge {cat.is_active ? 'pp-cat-status-active' : 'pp-cat-status-closed'}">{cat.is_active ? '使用中' : '已關閉'}</span>
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

  <div class="pp-cat-card-list">
    {#if isLoading && categories.length === 0}
      <div class="table-empty">載入中...</div>
    {:else if categories.length === 0}
      <div class="table-empty">無資料</div>
    {:else}
      {#each categories as cat (cat.id)}
        <div class="pp-cat-card">
          <div class="pp-cat-card-head">
            <span class="pp-cat-card-title">{cat.name}</span>
            <span class="badge {cat.is_active ? 'pp-cat-status-active' : 'pp-cat-status-closed'}">{cat.is_active ? '使用中' : '已關閉'}</span>
          </div>
          <div class="pp-cat-card-accts">
            <span>預付 {cat.account_id}</span>
            <span>費用 {cat.expense_account_id}</span>
          </div>
          {#if cat.is_active}
            <div class="pp-cat-card-footer">
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
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="pp-cat-create-title">
      <div class="modal-header">
        <h2 class="modal-title" id="pp-cat-create-title">新增預付費用類別</h2>
        <button class="modal-close" onclick={() => { showCreateModal = false; }} aria-label="關閉">×</button>
      </div>
      <form class="modal-body" onsubmit={handleCreate}>
        {#if createError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{createError}</p>
        {/if}
        <div class="form-group">
          <label class="form-label" for="pp-cat-name">名稱 *</label>
          <input id="pp-cat-name" class="form-input" type="text" bind:value={createForm.name} placeholder="例：辦公室租金" required />
        </div>
        <div class="form-group">
          <span class="form-label">預付費用科目 *</span>
          <AccountSelect {accounts} value={createForm.account_id} placeholder="選擇預付費用資產科目…" onselect={(id) => { createForm.account_id = id; }} />
        </div>
        <div class="form-group">
          <span class="form-label">費用科目 *</span>
          <AccountSelect {accounts} value={createForm.expense_account_id} placeholder="選擇攤提費用科目…" onselect={(id) => { createForm.expense_account_id = id; }} />
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
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="pp-cat-edit-title">
      <div class="modal-header">
        <h2 class="modal-title" id="pp-cat-edit-title">修改預付費用類別</h2>
        <button class="modal-close" onclick={() => { showEditModal = false; }} aria-label="關閉">×</button>
      </div>
      <form class="modal-body" onsubmit={handleEdit}>
        {#if editError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{editError}</p>
        {/if}
        <div class="form-group">
          <label class="form-label" for="pp-cat-edit-name">名稱 *</label>
          <input id="pp-cat-edit-name" class="form-input" type="text" bind:value={editForm.name} required />
        </div>
        <div class="form-group">
          <span class="form-label">預付費用科目 *</span>
          <AccountSelect {accounts} value={editForm.account_id} placeholder="選擇預付費用資產科目…" onselect={(id) => { editForm.account_id = id; }} />
        </div>
        <div class="form-group">
          <span class="form-label">費用科目 *</span>
          <AccountSelect {accounts} value={editForm.expense_account_id} placeholder="選擇攤提費用科目…" onselect={(id) => { editForm.expense_account_id = id; }} />
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
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="pp-cat-close-title">
      <div class="modal-header">
        <h2 class="modal-title" id="pp-cat-close-title">關閉預付費用類別</h2>
        <button class="modal-close" onclick={() => { showCloseModal = false; }} aria-label="關閉">×</button>
      </div>
      <form class="modal-body" onsubmit={handleClose}>
        {#if closeError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{closeError}</p>
        {/if}
        <p style="font-size:13px;color:#c0bdb4;margin-bottom:16px;">確定要關閉類別「{closingCat?.name}」？關閉後將無法用於新增預付費用。</p>
        <div class="modal-footer" style="padding:0;margin-top:8px;">
          <button type="button" class="btn-ghost" onclick={() => { showCloseModal = false; }} disabled={isClosing}>取消</button>
          <button type="submit" class="btn-primary" disabled={isClosing}>{isClosing ? '處理中…' : '確認關閉'}</button>
        </div>
      </form>
    </div>
  </div>
{/if}
