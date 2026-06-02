<script lang="ts">
  import { getSysAccounts, updateSysAccount } from '../api/sys';
  import { getAccountAll } from '../api/account';
  import AccountSelect from '../components/AccountSelect.svelte';
  import type { SysAccount } from '../types/sys';
  import type { Account } from '../types/account';
  import { getAllFixedAssetCategories, createFixedAssetCategory, updateFixedAssetCategory, deleteFixedAssetCategory } from '../api/fixedAssetCategory';
  import { getAllPrepaidCategories, createPrepaidCategory, updatePrepaidCategory, deletePrepaidCategory } from '../api/prepaidCategory';
  import type { FixedAssetCategory } from '../types/fixedAssetCategory';
  import type { PrepaidCategory } from '../types/prepaidCategory';

  let sysAccounts = $state<SysAccount[]>([]);
  let accounts    = $state<Account[]>([]);
  let isLoading   = $state(false);
  let error       = $state('');

  let showModal  = $state(false);
  let isSaving   = $state(false);
  let saveError  = $state('');
  let editForm   = $state<SysAccount>({ sys_code: '', description: '', account_id: '' });

  const accountMap = $derived(new Map(accounts.map(a => [a.account_id, a])));

  // ── 固定資產類別 ──
  interface FaCatForm {
    name:                            string;
    asset_account_id:                string;
    accum_depreciation_account_id:   string;
    depreciation_expense_account_id: string;
  }

  let faCategories          = $state<FixedAssetCategory[]>([]);
  let isFaCatLoading        = $state(false);
  let faCatError            = $state('');
  let showFaCatCreateModal  = $state(false);
  let isCreatingFaCat       = $state(false);
  let faCatCreateError      = $state('');
  let faCatCreateForm       = $state<FaCatForm>(emptyFaCatForm());
  let showFaCatEditModal    = $state(false);
  let isUpdatingFaCat       = $state(false);
  let faCatEditError        = $state('');
  let editingFaCat          = $state<FixedAssetCategory | null>(null);
  let faCatEditForm         = $state<FaCatForm>(emptyFaCatForm());
  let showFaCatCloseModal   = $state(false);
  let isClosingFaCat        = $state(false);
  let faCatCloseError       = $state('');
  let closingFaCat          = $state<FixedAssetCategory | null>(null);

  const isFaCatCreateValid = $derived(
    faCatCreateForm.name.trim()                              !== '' &&
    faCatCreateForm.asset_account_id                         !== '' &&
    faCatCreateForm.accum_depreciation_account_id            !== '' &&
    faCatCreateForm.depreciation_expense_account_id          !== ''
  );

  const isFaCatEditValid = $derived(
    faCatEditForm.name.trim()                              !== '' &&
    faCatEditForm.asset_account_id                         !== '' &&
    faCatEditForm.accum_depreciation_account_id            !== '' &&
    faCatEditForm.depreciation_expense_account_id          !== ''
  );

  // ── 預付費用類別 ──
  interface PpCatForm {
    name:               string;
    account_id:         string;
    expense_account_id: string;
  }

  let ppCategories          = $state<PrepaidCategory[]>([]);
  let isPpCatLoading        = $state(false);
  let ppCatError            = $state('');
  let showPpCatCreateModal  = $state(false);
  let isCreatingPpCat       = $state(false);
  let ppCatCreateError      = $state('');
  let ppCatCreateForm       = $state<PpCatForm>(emptyPpCatForm());
  let showPpCatEditModal    = $state(false);
  let isUpdatingPpCat       = $state(false);
  let ppCatEditError        = $state('');
  let editingPpCat          = $state<PrepaidCategory | null>(null);
  let ppCatEditForm         = $state<PpCatForm>(emptyPpCatForm());
  let showPpCatCloseModal   = $state(false);
  let isClosingPpCat        = $state(false);
  let ppCatCloseError       = $state('');
  let closingPpCat          = $state<PrepaidCategory | null>(null);

  const isPpCatCreateValid = $derived(
    ppCatCreateForm.name.trim()        !== '' &&
    ppCatCreateForm.account_id         !== '' &&
    ppCatCreateForm.expense_account_id !== ''
  );

  const isPpCatEditValid = $derived(
    ppCatEditForm.name.trim()        !== '' &&
    ppCatEditForm.account_id         !== '' &&
    ppCatEditForm.expense_account_id !== ''
  );

  $effect(() => {
    void load();
    void loadFixedAssetCategories();
    void loadPrepaidCategories();
  });

  async function load(): Promise<void> {
    isLoading = true;
    error     = '';
    try {
      [sysAccounts, accounts] = await Promise.all([getSysAccounts(), getAccountAll()]);
    } catch (err) {
      error = err instanceof Error ? err.message : '查詢失敗，請稍後再試。';
    } finally {
      isLoading = false;
    }
  }

  function openEditModal(entry: SysAccount): void {
    editForm   = { sys_code: entry.sys_code, description: entry.description, account_id: entry.account_id };
    saveError  = '';
    showModal  = true;
  }

  function closeModal(): void {
    showModal = false;
  }

  async function handleSubmit(e: Event): Promise<void> {
    e.preventDefault();
    if (!editForm.account_id) return;
    isSaving  = true;
    saveError = '';
    try {
      await updateSysAccount(editForm);
      showModal = false;
      await load();
    } catch (err) {
      saveError = err instanceof Error ? err.message : '更新失敗';
    } finally {
      isSaving = false;
    }
  }

  function emptyFaCatForm(): FaCatForm {
    return { name: '', asset_account_id: '', accum_depreciation_account_id: '', depreciation_expense_account_id: '' };
  }

  async function loadFixedAssetCategories(): Promise<void> {
    isFaCatLoading = true;
    faCatError     = '';
    try {
      faCategories = await getAllFixedAssetCategories();
    } catch (err) {
      faCatError = err instanceof Error ? err.message : '查詢失敗，請稍後再試。';
    } finally {
      isFaCatLoading = false;
    }
  }

  function openFaCatCreateModal(): void {
    faCatCreateForm  = emptyFaCatForm();
    faCatCreateError = '';
    showFaCatCreateModal = true;
  }

  async function handleFaCatCreate(e: Event): Promise<void> {
    e.preventDefault();
    if (!isFaCatCreateValid) return;
    isCreatingFaCat  = true;
    faCatCreateError = '';
    try {
      await createFixedAssetCategory({
        name:                            faCatCreateForm.name.trim(),
        asset_account_id:                faCatCreateForm.asset_account_id,
        accum_depreciation_account_id:   faCatCreateForm.accum_depreciation_account_id,
        depreciation_expense_account_id: faCatCreateForm.depreciation_expense_account_id,
      });
      showFaCatCreateModal = false;
      await loadFixedAssetCategories();
    } catch (err) {
      faCatCreateError = err instanceof Error ? err.message : '新增失敗，請稍後再試。';
    } finally {
      isCreatingFaCat = false;
    }
  }

  function openFaCatEditModal(cat: FixedAssetCategory): void {
    editingFaCat  = cat;
    faCatEditForm = {
      name:                            cat.name,
      asset_account_id:                cat.asset_account_id,
      accum_depreciation_account_id:   cat.accum_depreciation_account_id,
      depreciation_expense_account_id: cat.depreciation_expense_account_id,
    };
    faCatEditError = '';
    showFaCatEditModal = true;
  }

  async function handleFaCatEdit(e: Event): Promise<void> {
    e.preventDefault();
    if (!isFaCatEditValid || !editingFaCat) return;
    isUpdatingFaCat  = true;
    faCatEditError   = '';
    try {
      await updateFixedAssetCategory(editingFaCat.category_uuid, {
        expected_version:                editingFaCat.version,
        name:                            faCatEditForm.name.trim(),
        asset_account_id:                faCatEditForm.asset_account_id,
        accum_depreciation_account_id:   faCatEditForm.accum_depreciation_account_id,
        depreciation_expense_account_id: faCatEditForm.depreciation_expense_account_id,
      });
      showFaCatEditModal = false;
      await loadFixedAssetCategories();
    } catch (err) {
      faCatEditError = err instanceof Error ? err.message : '修改失敗，請稍後再試。';
    } finally {
      isUpdatingFaCat = false;
    }
  }

  function openFaCatCloseModal(cat: FixedAssetCategory): void {
    closingFaCat    = cat;
    faCatCloseError = '';
    showFaCatCloseModal = true;
  }

  async function handleFaCatClose(e: Event): Promise<void> {
    e.preventDefault();
    if (!closingFaCat) return;
    isClosingFaCat  = true;
    faCatCloseError = '';
    try {
      await deleteFixedAssetCategory(closingFaCat.category_uuid, closingFaCat.version);
      showFaCatCloseModal = false;
      await loadFixedAssetCategories();
    } catch (err) {
      faCatCloseError = err instanceof Error ? err.message : '關閉失敗，請稍後再試。';
    } finally {
      isClosingFaCat = false;
    }
  }

  function emptyPpCatForm(): PpCatForm {
    return { name: '', account_id: '', expense_account_id: '' };
  }

  async function loadPrepaidCategories(): Promise<void> {
    isPpCatLoading = true;
    ppCatError     = '';
    try {
      ppCategories = await getAllPrepaidCategories();
    } catch (err) {
      ppCatError = err instanceof Error ? err.message : '查詢失敗，請稍後再試。';
    } finally {
      isPpCatLoading = false;
    }
  }

  function openPpCatCreateModal(): void {
    ppCatCreateForm  = emptyPpCatForm();
    ppCatCreateError = '';
    showPpCatCreateModal = true;
  }

  async function handlePpCatCreate(e: Event): Promise<void> {
    e.preventDefault();
    if (!isPpCatCreateValid) return;
    isCreatingPpCat  = true;
    ppCatCreateError = '';
    try {
      await createPrepaidCategory({
        name:               ppCatCreateForm.name.trim(),
        account_id:         ppCatCreateForm.account_id,
        expense_account_id: ppCatCreateForm.expense_account_id,
      });
      showPpCatCreateModal = false;
      await loadPrepaidCategories();
    } catch (err) {
      ppCatCreateError = err instanceof Error ? err.message : '新增失敗，請稍後再試。';
    } finally {
      isCreatingPpCat = false;
    }
  }

  function openPpCatEditModal(cat: PrepaidCategory): void {
    editingPpCat  = cat;
    ppCatEditForm = {
      name:               cat.name,
      account_id:         cat.account_id,
      expense_account_id: cat.expense_account_id,
    };
    ppCatEditError = '';
    showPpCatEditModal = true;
  }

  async function handlePpCatEdit(e: Event): Promise<void> {
    e.preventDefault();
    if (!isPpCatEditValid || !editingPpCat) return;
    isUpdatingPpCat  = true;
    ppCatEditError   = '';
    try {
      await updatePrepaidCategory(editingPpCat.category_uuid, {
        expected_version:   editingPpCat.version,
        name:               ppCatEditForm.name.trim(),
        account_id:         ppCatEditForm.account_id,
        expense_account_id: ppCatEditForm.expense_account_id,
      });
      showPpCatEditModal = false;
      await loadPrepaidCategories();
    } catch (err) {
      ppCatEditError = err instanceof Error ? err.message : '修改失敗，請稍後再試。';
    } finally {
      isUpdatingPpCat = false;
    }
  }

  function openPpCatCloseModal(cat: PrepaidCategory): void {
    closingPpCat    = cat;
    ppCatCloseError = '';
    showPpCatCloseModal = true;
  }

  async function handlePpCatClose(e: Event): Promise<void> {
    e.preventDefault();
    if (!closingPpCat) return;
    isClosingPpCat  = true;
    ppCatCloseError = '';
    try {
      await deletePrepaidCategory(closingPpCat.category_uuid, closingPpCat.version);
      showPpCatCloseModal = false;
      await loadPrepaidCategories();
    } catch (err) {
      ppCatCloseError = err instanceof Error ? err.message : '關閉失敗，請稍後再試。';
    } finally {
      isClosingPpCat = false;
    }
  }
</script>

<div class="content-header">
  <h1 class="content-title">系統設定</h1>
  <span class="content-date">共 {sysAccounts.length} 項設定</span>
</div>

{#if error}
  <p class="query-error" role="alert">{error}</p>
{/if}

<section class="section">
  <header class="section-header">
    <h2 class="section-title">系統科目對應</h2>
    {#if isLoading}
      <span class="query-loading">
        <span class="spinner" aria-hidden="true"></span>
        載入中
      </span>
    {/if}
  </header>

  <div class="table-wrap">
    <table class="data-table" aria-label="系統科目對應">
      <thead>
        <tr>
          <th>用途</th>
          <th class="mono">系統代碼</th>
          <th>對應科目編號</th>
          <th>對應科目名稱</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        {#if isLoading && sysAccounts.length === 0}
          <tr><td colspan="5" class="table-empty">載入中...</td></tr>
        {:else if sysAccounts.length === 0}
          <tr><td colspan="5" class="table-empty">無資料</td></tr>
        {:else}
          {#each sysAccounts as entry (entry.sys_code)}
            {@const acct = accountMap.get(entry.account_id)}
            <tr>
              <td>{entry.description}</td>
              <td class="mono">{entry.sys_code}</td>
              <td class="mono">{entry.account_id || '—'}</td>
              <td>{acct?.name ?? '—'}</td>
              <td>
                <button
                  class="btn-ghost"
                  style="padding:2px 10px;font-size:11px;"
                  onclick={() => openEditModal(entry)}
                >編輯</button>
              </td>
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  </div>
</section>

<section class="section">
  <header class="section-header">
    <h2 class="section-title">固定資產類別</h2>
    <div style="display:flex;align-items:center;gap:16px;">
      {#if isFaCatLoading}
        <span class="query-loading"><span class="spinner" aria-hidden="true"></span>載入中</span>
      {/if}
      <button class="section-action" onclick={openFaCatCreateModal}>＋ 新增類別</button>
    </div>
  </header>

  {#if faCatError}
    <p class="query-error" role="alert">{faCatError}</p>
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
        {#if isFaCatLoading && faCategories.length === 0}
          <tr><td colspan="7" class="table-empty">載入中...</td></tr>
        {:else if faCategories.length === 0}
          <tr><td colspan="7" class="table-empty">無資料</td></tr>
        {:else}
          {#each faCategories as cat (cat.id)}
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
                  <button type="button" class="btn-ghost" style="padding:1px 6px;font-size:10px;margin-left:4px;" onclick={() => openFaCatEditModal(cat)}>修改</button>
                  <button type="button" class="btn-ghost" style="padding:1px 6px;font-size:10px;margin-left:4px;" onclick={() => openFaCatCloseModal(cat)}>關閉</button>
                {/if}
              </td>
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  </div>

  <div class="fa-cat-card-list">
    {#if isFaCatLoading && faCategories.length === 0}
      <div class="table-empty">載入中...</div>
    {:else if faCategories.length === 0}
      <div class="table-empty">無資料</div>
    {:else}
      {#each faCategories as cat (cat.id)}
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
              <button type="button" class="btn-ghost" style="padding:1px 6px;font-size:10px;" onclick={() => openFaCatEditModal(cat)}>修改</button>
              <button type="button" class="btn-ghost" style="padding:1px 6px;font-size:10px;" onclick={() => openFaCatCloseModal(cat)}>關閉</button>
            </div>
          {/if}
        </div>
      {/each}
    {/if}
  </div>
</section>

<section class="section">
  <header class="section-header">
    <h2 class="section-title">預付費用類別</h2>
    <div style="display:flex;align-items:center;gap:16px;">
      {#if isPpCatLoading}
        <span class="query-loading"><span class="spinner" aria-hidden="true"></span>載入中</span>
      {/if}
      <button class="section-action" onclick={openPpCatCreateModal}>＋ 新增類別</button>
    </div>
  </header>

  {#if ppCatError}
    <p class="query-error" role="alert">{ppCatError}</p>
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
        {#if isPpCatLoading && ppCategories.length === 0}
          <tr><td colspan="6" class="table-empty">載入中...</td></tr>
        {:else if ppCategories.length === 0}
          <tr><td colspan="6" class="table-empty">無資料</td></tr>
        {:else}
          {#each ppCategories as cat (cat.id)}
            <tr>
              <td>{cat.name}</td>
              <td class="mono" style="font-size:11px;">{cat.account_id}</td>
              <td class="mono" style="font-size:11px;">{cat.expense_account_id}</td>
              <td class="hidden md:table-cell">{cat.updated_by ?? '—'}</td>
              <td class="hidden md:table-cell">{cat.updated_at ?? '—'}</td>
              <td>
                <span class="badge {cat.is_active ? 'pp-cat-status-active' : 'pp-cat-status-closed'}">{cat.is_active ? '使用中' : '已關閉'}</span>
                {#if cat.is_active}
                  <button type="button" class="btn-ghost" style="padding:1px 6px;font-size:10px;margin-left:4px;" onclick={() => openPpCatEditModal(cat)}>修改</button>
                  <button type="button" class="btn-ghost" style="padding:1px 6px;font-size:10px;margin-left:4px;" onclick={() => openPpCatCloseModal(cat)}>關閉</button>
                {/if}
              </td>
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  </div>

  <div class="pp-cat-card-list">
    {#if isPpCatLoading && ppCategories.length === 0}
      <div class="table-empty">載入中...</div>
    {:else if ppCategories.length === 0}
      <div class="table-empty">無資料</div>
    {:else}
      {#each ppCategories as cat (cat.id)}
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
              <button type="button" class="btn-ghost" style="padding:1px 6px;font-size:10px;" onclick={() => openPpCatEditModal(cat)}>修改</button>
              <button type="button" class="btn-ghost" style="padding:1px 6px;font-size:10px;" onclick={() => openPpCatCloseModal(cat)}>關閉</button>
            </div>
          {/if}
        </div>
      {/each}
    {/if}
  </div>
</section>

{#if showModal}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) closeModal(); }}>
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="sys-modal-title">
      <div class="modal-header">
        <h2 class="modal-title" id="sys-modal-title">修改系統科目對應</h2>
        <button class="modal-close" onclick={closeModal} aria-label="關閉">×</button>
      </div>

      <form class="modal-body" onsubmit={handleSubmit}>
        {#if saveError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{saveError}</p>
        {/if}

        <div class="form-group">
          <label class="form-label" for="sys_code">用途</label>
          <p style="font-size:13px;color:#dedad3;margin:0;padding:8px 0;">{editForm.description}</p>
        </div>

        <div class="form-group">
          <label class="form-label" for="sys_code">系統代碼</label>
          <p class="mono" style="font-size:11px;color:#5c6278;margin:0;padding:8px 0;">{editForm.sys_code}</p>
        </div>

        <div class="form-group">
          <label class="form-label" for="account_id">對應科目 *</label>
          <AccountSelect
            {accounts}
            value={editForm.account_id}
            placeholder="選擇會計科目…"
            onselect={(id) => { editForm.account_id = id; }}
          />
        </div>

        <div class="modal-footer" style="padding:0;margin-top:8px;">
          <button type="button" class="btn-ghost" onclick={closeModal} disabled={isSaving}>取消</button>
          <button
            type="submit"
            class="btn-primary"
            disabled={isSaving || !editForm.account_id}
          >
            {isSaving ? '儲存中…' : '儲存修改'}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- ── 固定資產類別：新增 Modal ───────────── -->
{#if showFaCatCreateModal}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) showFaCatCreateModal = false; }}>
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="fa-cat-create-title">
      <div class="modal-header">
        <h2 class="modal-title" id="fa-cat-create-title">新增固定資產類別</h2>
        <button class="modal-close" onclick={() => { showFaCatCreateModal = false; }} aria-label="關閉">×</button>
      </div>
      <form class="modal-body" onsubmit={handleFaCatCreate}>
        {#if faCatCreateError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{faCatCreateError}</p>
        {/if}
        <div class="form-group">
          <label class="form-label" for="fa-cat-name">名稱 *</label>
          <input id="fa-cat-name" class="form-input" type="text" bind:value={faCatCreateForm.name} placeholder="例：辦公設備" required />
        </div>
        <div class="form-group">
          <span class="form-label">資產科目 *</span>
          <AccountSelect {accounts} value={faCatCreateForm.asset_account_id} placeholder="選擇資產科目…" onselect={(id) => { faCatCreateForm.asset_account_id = id; }} />
        </div>
        <div class="form-group">
          <span class="form-label">累計折舊科目 *</span>
          <AccountSelect {accounts} value={faCatCreateForm.accum_depreciation_account_id} placeholder="選擇累計折舊科目…" onselect={(id) => { faCatCreateForm.accum_depreciation_account_id = id; }} />
        </div>
        <div class="form-group">
          <span class="form-label">折舊費用科目 *</span>
          <AccountSelect {accounts} value={faCatCreateForm.depreciation_expense_account_id} placeholder="選擇折舊費用科目…" onselect={(id) => { faCatCreateForm.depreciation_expense_account_id = id; }} />
        </div>
        <div class="modal-footer" style="padding:0;margin-top:8px;">
          <button type="button" class="btn-ghost" onclick={() => { showFaCatCreateModal = false; }} disabled={isCreatingFaCat}>取消</button>
          <button type="submit" class="btn-primary" disabled={isCreatingFaCat || !isFaCatCreateValid}>{isCreatingFaCat ? '建立中…' : '建立類別'}</button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- ── 固定資產類別：修改 Modal ───────────── -->
{#if showFaCatEditModal}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) showFaCatEditModal = false; }}>
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="fa-cat-edit-title">
      <div class="modal-header">
        <h2 class="modal-title" id="fa-cat-edit-title">修改固定資產類別</h2>
        <button class="modal-close" onclick={() => { showFaCatEditModal = false; }} aria-label="關閉">×</button>
      </div>
      <form class="modal-body" onsubmit={handleFaCatEdit}>
        {#if faCatEditError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{faCatEditError}</p>
        {/if}
        <div class="form-group">
          <label class="form-label" for="fa-cat-edit-name">名稱 *</label>
          <input id="fa-cat-edit-name" class="form-input" type="text" bind:value={faCatEditForm.name} required />
        </div>
        <div class="form-group">
          <span class="form-label">資產科目 *</span>
          <AccountSelect {accounts} value={faCatEditForm.asset_account_id} placeholder="選擇資產科目…" onselect={(id) => { faCatEditForm.asset_account_id = id; }} />
        </div>
        <div class="form-group">
          <span class="form-label">累計折舊科目 *</span>
          <AccountSelect {accounts} value={faCatEditForm.accum_depreciation_account_id} placeholder="選擇累計折舊科目…" onselect={(id) => { faCatEditForm.accum_depreciation_account_id = id; }} />
        </div>
        <div class="form-group">
          <span class="form-label">折舊費用科目 *</span>
          <AccountSelect {accounts} value={faCatEditForm.depreciation_expense_account_id} placeholder="選擇折舊費用科目…" onselect={(id) => { faCatEditForm.depreciation_expense_account_id = id; }} />
        </div>
        <div class="modal-footer" style="padding:0;margin-top:8px;">
          <button type="button" class="btn-ghost" onclick={() => { showFaCatEditModal = false; }} disabled={isUpdatingFaCat}>取消</button>
          <button type="submit" class="btn-primary" disabled={isUpdatingFaCat || !isFaCatEditValid}>{isUpdatingFaCat ? '更新中…' : '儲存修改'}</button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- ── 固定資產類別：關閉 Modal ───────────── -->
{#if showFaCatCloseModal}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) showFaCatCloseModal = false; }}>
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="fa-cat-close-title">
      <div class="modal-header">
        <h2 class="modal-title" id="fa-cat-close-title">關閉固定資產類別</h2>
        <button class="modal-close" onclick={() => { showFaCatCloseModal = false; }} aria-label="關閉">×</button>
      </div>
      <form class="modal-body" onsubmit={handleFaCatClose}>
        {#if faCatCloseError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{faCatCloseError}</p>
        {/if}
        <p style="font-size:13px;color:#c0bdb4;margin-bottom:16px;">確定要關閉類別「{closingFaCat?.name}」？關閉後將無法用於新增資產。</p>
        <div class="modal-footer" style="padding:0;margin-top:8px;">
          <button type="button" class="btn-ghost" onclick={() => { showFaCatCloseModal = false; }} disabled={isClosingFaCat}>取消</button>
          <button type="submit" class="btn-primary" disabled={isClosingFaCat}>{isClosingFaCat ? '處理中…' : '確認關閉'}</button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- ── 預付費用類別：新增 Modal ───────────── -->
{#if showPpCatCreateModal}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) showPpCatCreateModal = false; }}>
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="pp-cat-create-title">
      <div class="modal-header">
        <h2 class="modal-title" id="pp-cat-create-title">新增預付費用類別</h2>
        <button class="modal-close" onclick={() => { showPpCatCreateModal = false; }} aria-label="關閉">×</button>
      </div>
      <form class="modal-body" onsubmit={handlePpCatCreate}>
        {#if ppCatCreateError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{ppCatCreateError}</p>
        {/if}
        <div class="form-group">
          <label class="form-label" for="pp-cat-name">名稱 *</label>
          <input id="pp-cat-name" class="form-input" type="text" bind:value={ppCatCreateForm.name} placeholder="例：辦公室租金" required />
        </div>
        <div class="form-group">
          <span class="form-label">預付費用科目 *</span>
          <AccountSelect {accounts} value={ppCatCreateForm.account_id} placeholder="選擇預付費用資產科目…" onselect={(id) => { ppCatCreateForm.account_id = id; }} />
        </div>
        <div class="form-group">
          <span class="form-label">費用科目 *</span>
          <AccountSelect {accounts} value={ppCatCreateForm.expense_account_id} placeholder="選擇攤提費用科目…" onselect={(id) => { ppCatCreateForm.expense_account_id = id; }} />
        </div>
        <div class="modal-footer" style="padding:0;margin-top:8px;">
          <button type="button" class="btn-ghost" onclick={() => { showPpCatCreateModal = false; }} disabled={isCreatingPpCat}>取消</button>
          <button type="submit" class="btn-primary" disabled={isCreatingPpCat || !isPpCatCreateValid}>{isCreatingPpCat ? '建立中…' : '建立類別'}</button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- ── 預付費用類別：修改 Modal ───────────── -->
{#if showPpCatEditModal}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) showPpCatEditModal = false; }}>
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="pp-cat-edit-title">
      <div class="modal-header">
        <h2 class="modal-title" id="pp-cat-edit-title">修改預付費用類別</h2>
        <button class="modal-close" onclick={() => { showPpCatEditModal = false; }} aria-label="關閉">×</button>
      </div>
      <form class="modal-body" onsubmit={handlePpCatEdit}>
        {#if ppCatEditError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{ppCatEditError}</p>
        {/if}
        <div class="form-group">
          <label class="form-label" for="pp-cat-edit-name">名稱 *</label>
          <input id="pp-cat-edit-name" class="form-input" type="text" bind:value={ppCatEditForm.name} required />
        </div>
        <div class="form-group">
          <span class="form-label">預付費用科目 *</span>
          <AccountSelect {accounts} value={ppCatEditForm.account_id} placeholder="選擇預付費用資產科目…" onselect={(id) => { ppCatEditForm.account_id = id; }} />
        </div>
        <div class="form-group">
          <span class="form-label">費用科目 *</span>
          <AccountSelect {accounts} value={ppCatEditForm.expense_account_id} placeholder="選擇攤提費用科目…" onselect={(id) => { ppCatEditForm.expense_account_id = id; }} />
        </div>
        <div class="modal-footer" style="padding:0;margin-top:8px;">
          <button type="button" class="btn-ghost" onclick={() => { showPpCatEditModal = false; }} disabled={isUpdatingPpCat}>取消</button>
          <button type="submit" class="btn-primary" disabled={isUpdatingPpCat || !isPpCatEditValid}>{isUpdatingPpCat ? '更新中…' : '儲存修改'}</button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- ── 預付費用類別：關閉 Modal ───────────── -->
{#if showPpCatCloseModal}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) showPpCatCloseModal = false; }}>
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="pp-cat-close-title">
      <div class="modal-header">
        <h2 class="modal-title" id="pp-cat-close-title">關閉預付費用類別</h2>
        <button class="modal-close" onclick={() => { showPpCatCloseModal = false; }} aria-label="關閉">×</button>
      </div>
      <form class="modal-body" onsubmit={handlePpCatClose}>
        {#if ppCatCloseError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{ppCatCloseError}</p>
        {/if}
        <p style="font-size:13px;color:#c0bdb4;margin-bottom:16px;">確定要關閉類別「{closingPpCat?.name}」？關閉後將無法用於新增預付費用。</p>
        <div class="modal-footer" style="padding:0;margin-top:8px;">
          <button type="button" class="btn-ghost" onclick={() => { showPpCatCloseModal = false; }} disabled={isClosingPpCat}>取消</button>
          <button type="submit" class="btn-primary" disabled={isClosingPpCat}>{isClosingPpCat ? '處理中…' : '確認關閉'}</button>
        </div>
      </form>
    </div>
  </div>
{/if}
