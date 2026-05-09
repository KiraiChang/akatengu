<script lang="ts">
  import {
    getPeriodPaged,
    startMonthPeriod,
    closeMonthPeriod,
    reopenMonthPeriod,
    startAnnualPeriod,
    closeAnnualPeriod,
    reopenAnnualPeriod,
  } from '../api/period';
  import type { PeriodClosing, PeriodStatus } from '../types/period';

  const PAGE_SIZE = 20;

  type Tab = 'MONTHLY' | 'ANNUAL';

  let activeTab   = $state<Tab>('MONTHLY');
  let page        = $state(1);
  let total       = $state(0);
  let periods     = $state<PeriodClosing[]>([]);
  let isLoading   = $state(false);
  let loadError   = $state('');

  // ── 新增期間 modal ─────────────────────────
  let showCreateModal = $state(false);
  let createMonthStart = $state(currentYearMonth());
  let createYear       = $state(new Date().getFullYear());
  let isCreating       = $state(false);
  let createError      = $state('');

  // ── 關帳 modal ──────────────────────────────
  let closingTarget = $state<PeriodClosing | null>(null);
  let closeDate     = $state(today());
  let isClosing     = $state(false);
  let closeError    = $state('');

  // ── 重開 modal ──────────────────────────────
  let reopenTarget  = $state<PeriodClosing | null>(null);
  let reopenDate    = $state(today());
  let reopenReason  = $state('');
  let isReopening   = $state(false);
  let reopenError   = $state('');

  $effect(() => {
    void loadPage(activeTab, page);
  });

  async function loadPage(tab: Tab, p: number): Promise<void> {
    isLoading = true;
    loadError = '';
    try {
      const res = await getPeriodPaged(tab, { page: p - 1, pageSize: PAGE_SIZE });
      periods = res.data ?? [];
      total   = res.meta.total_pages ?? 0;
    } catch (err) {
      loadError = err instanceof Error ? err.message : '載入失敗';
    } finally {
      isLoading = false;
    }
  }

  function switchTab(tab: Tab): void {
    if (activeTab === tab) return;
    activeTab = tab;
    page      = 1;
  }

  function today(): string {
    return new Date().toISOString().slice(0, 10);
  }

  function currentYearMonth(): string {
    const d = new Date();
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`;
  }

  // ── 新增 ──────────────────────────────────

  function openCreateModal(): void {
    createMonthStart = currentYearMonth();
    createYear       = new Date().getFullYear();
    createError      = '';
    showCreateModal  = true;
  }

  async function handleCreate(e: Event): Promise<void> {
    e.preventDefault();
    isCreating  = true;
    createError = '';
    try {
      if (activeTab === 'MONTHLY') {
        await startMonthPeriod({ period_start: `${createMonthStart}` });
      } else {
        await startAnnualPeriod({ year: createYear });
      }
      showCreateModal = false;
      page = 1;
      await loadPage(activeTab, 1);
    } catch (err) {
      createError = err instanceof Error ? err.message : '新增失敗';
    } finally {
      isCreating = false;
    }
  }

  // ── 關帳 ──────────────────────────────────

  function openCloseModal(p: PeriodClosing): void {
    closingTarget = p;
    closeDate     = today();
    closeError    = '';
  }

  async function handleClose(e: Event): Promise<void> {
    e.preventDefault();
    if (!closingTarget) return;
    isClosing  = true;
    closeError = '';
    try {
      if (activeTab === 'MONTHLY') {
        await closeMonthPeriod({ closing_id: closingTarget.closing_id, closed_at: closeDate });
      } else {
        await closeAnnualPeriod({ closing_id: closingTarget.closing_id, closed_at: closeDate });
      }
      closingTarget = null;
      await loadPage(activeTab, page);
    } catch (err) {
      closeError = err instanceof Error ? err.message : '關帳失敗';
    } finally {
      isClosing = false;
    }
  }

  // ── 重開 ──────────────────────────────────

  function openReopenModal(p: PeriodClosing): void {
    reopenTarget = p;
    reopenDate   = today();
    reopenReason = '';
    reopenError  = '';
  }

  async function handleReopen(e: Event): Promise<void> {
    e.preventDefault();
    if (!reopenTarget) return;
    isReopening  = true;
    reopenError  = '';
    try {
      if (activeTab === 'MONTHLY') {
        await reopenMonthPeriod({ closing_id: reopenTarget.closing_id, reason: reopenReason, reopened_at: reopenDate });
      } else {
        await reopenAnnualPeriod({ closing_id: reopenTarget.closing_id, reason: reopenReason, reopened_at: reopenDate });
      }
      reopenTarget = null;
      await loadPage(activeTab, page);
    } catch (err) {
      reopenError = err instanceof Error ? err.message : '重開失敗';
    } finally {
      isReopening = false;
    }
  }

  const totalPages = $derived(Math.max(1, Math.ceil(total / PAGE_SIZE)));

  const STATUS_LABEL: Record<PeriodStatus, string> = {
    OPEN:     '開帳中',
    CLOSED:   '已關帳',
    REOPENED: '已重開',
  };

  const TAB_LABEL: Record<Tab, string> = {
    MONTHLY: '月結',
    ANNUAL:  '年結',
  };
</script>

<div class="content-header">
  <h1 class="content-title">會計期間</h1>
  <button class="btn-primary" onclick={openCreateModal}>
    ＋ 新增{activeTab === 'MONTHLY' ? '月期' : '年期'}
  </button>
</div>

{#if loadError}
  <p class="query-error" role="alert">{loadError}</p>
{/if}

<div class="period-tabs">
  {#each (['MONTHLY', 'ANNUAL'] as Tab[]) as tab}
    <button
      class="period-tab {activeTab === tab ? 'period-tab--active' : ''}"
      onclick={() => switchTab(tab)}
    >{TAB_LABEL[tab]}</button>
  {/each}
</div>

<section class="section">
  <div class="period-table">
    <div class="period-thead">
      <span>開始日期</span>
      <span>結束日期</span>
      <span>狀態</span>
      <span>關帳日</span>
      <span>備註</span>
      <span></span>
    </div>

    {#if isLoading}
      <div class="period-empty">載入中…</div>
    {:else if periods.length === 0}
      <div class="period-empty">尚無期間記錄</div>
    {:else}
      {#each periods as p (p.closing_id)}
        <div class="period-row">
          <span class="period-cell">{p.period_start}</span>
          <span class="period-cell">{p.period_end}</span>
          <span class="period-cell">
            <span class="period-status period-status--{p.status.toLowerCase()}">
              {STATUS_LABEL[p.status]}
            </span>
          </span>
          <span class="period-cell period-dim">{p.closed_at ?? '—'}</span>
          <span class="period-cell period-dim">{p.note ?? ''}</span>
          <span class="period-cell period-actions">
            {#if p.status === 'OPEN' || p.status === 'REOPENED'}
              <button class="period-action-btn" onclick={() => openCloseModal(p)}>
                關帳
              </button>
            {/if}
            {#if p.status === 'CLOSED'}
              <button class="period-action-btn period-action-btn--ghost" onclick={() => openReopenModal(p)}>
                重新開帳
              </button>
            {/if}
          </span>
        </div>
      {/each}
    {/if}
  </div>

  {#if totalPages > 1}
    <div class="pagination">
      <button class="page-btn" disabled={page <= 1} onclick={() => { page -= 1; }}>‹</button>
      <span class="page-info">{page} / {totalPages}</span>
      <button class="page-btn" disabled={page >= totalPages} onclick={() => { page += 1; }}>›</button>
    </div>
  {/if}
</section>

<!-- ── 新增期間 modal ─────────────────────── -->
{#if showCreateModal}
  <div class="modal-overlay" role="dialog" aria-modal="true">
    <div class="modal">
      <header class="modal-header">
        <h2 class="modal-title">新增{TAB_LABEL[activeTab]}期間</h2>
        <button class="modal-close" onclick={() => { showCreateModal = false; }} aria-label="關閉">×</button>
      </header>
      <form class="modal-body" onsubmit={handleCreate}>
        {#if createError}
          <p class="query-error" style="margin-bottom:16px;" role="alert">{createError}</p>
        {/if}

        {#if activeTab === 'MONTHLY'}
          <div class="form-group">
            <label class="form-label" for="p-month-start">期間年月 *</label>
            <input
              id="p-month-start"
              class="form-input"
              type="month"
              bind:value={createMonthStart}
              required
            />
          </div>
        {:else}
          <div class="form-group">
            <label class="form-label" for="p-year">年份 *</label>
            <input
              id="p-year"
              class="form-input"
              type="number"
              min="2000"
              max="2099"
              bind:value={createYear}
              required
            />
          </div>
        {/if}

        <div class="je-actions">
          <button type="button" class="btn-ghost" onclick={() => { showCreateModal = false; }} disabled={isCreating}>取消</button>
          <button type="submit" class="btn-primary" disabled={isCreating}>
            {isCreating ? '建立中…' : '確認新增'}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- ── 關帳 modal ─────────────────────────── -->
{#if closingTarget}
  <div class="modal-overlay" role="dialog" aria-modal="true">
    <div class="modal">
      <header class="modal-header">
        <h2 class="modal-title">關帳確認</h2>
        <button class="modal-close" onclick={() => { closingTarget = null; }} aria-label="關閉">×</button>
      </header>
      <form class="modal-body" onsubmit={handleClose}>
        {#if closeError}
          <p class="query-error" style="margin-bottom:16px;" role="alert">{closeError}</p>
        {/if}

        <p class="period-confirm-text">
          確認對期間 <strong>{closingTarget.period_start} ～ {closingTarget.period_end}</strong> 執行關帳？
        </p>

        <div class="form-group">
          <label class="form-label" for="p-close-date">關帳日期 *</label>
          <input
            id="p-close-date"
            class="form-input"
            type="date"
            bind:value={closeDate}
            required
          />
        </div>

        <div class="je-actions">
          <button type="button" class="btn-ghost" onclick={() => { closingTarget = null; }} disabled={isClosing}>取消</button>
          <button type="submit" class="btn-primary" disabled={isClosing}>
            {isClosing ? '處理中…' : '確認關帳'}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- ── 重開 modal ─────────────────────────── -->
{#if reopenTarget}
  <div class="modal-overlay" role="dialog" aria-modal="true">
    <div class="modal">
      <header class="modal-header">
        <h2 class="modal-title">重新開帳</h2>
        <button class="modal-close" onclick={() => { reopenTarget = null; }} aria-label="關閉">×</button>
      </header>
      <form class="modal-body" onsubmit={handleReopen}>
        {#if reopenError}
          <p class="query-error" style="margin-bottom:16px;" role="alert">{reopenError}</p>
        {/if}

        <p class="period-confirm-text">
          期間 <strong>{reopenTarget.period_start} ～ {reopenTarget.period_end}</strong>
        </p>

        <div class="form-group">
          <label class="form-label" for="p-reopen-date">重開日期 *</label>
          <input
            id="p-reopen-date"
            class="form-input"
            type="date"
            bind:value={reopenDate}
            required
          />
        </div>

        <div class="form-group">
          <label class="form-label" for="p-reopen-reason">重開原因 *</label>
          <input
            id="p-reopen-reason"
            class="form-input"
            type="text"
            placeholder="說明重開原因"
            bind:value={reopenReason}
            required
          />
        </div>

        <div class="je-actions">
          <button type="button" class="btn-ghost" onclick={() => { reopenTarget = null; }} disabled={isReopening}>取消</button>
          <button type="submit" class="btn-primary" disabled={isReopening || !reopenReason.trim()}>
            {isReopening ? '處理中…' : '確認重開'}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}
