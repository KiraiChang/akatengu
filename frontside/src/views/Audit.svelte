<script lang="ts">
  import {
    getAggregateVersions, getEventStorePaged,
    getCheckpoints, getSnapshots, replayProjection,
    exportEvents, importEvents,
  } from '../api/audit';
  import type { AggregateVersionAudit, EventStoreAudit, CheckpointAudit, SnapshotAudit } from '../types/audit';

  type AuditTab = 'versions' | 'events' | 'checkpoints' | 'snapshots';

  const PAGE_SIZE = 10;

  let isLoading    = $state(true);
  let error        = $state('');
  let activeTab    = $state<AuditTab>('versions');

  let versions     = $state<AggregateVersionAudit[]>([]);
  let checkpoints  = $state<CheckpointAudit[]>([]);
  let snapshots    = $state<SnapshotAudit[]>([]);
  let events       = $state<EventStoreAudit[]>([]);
  let eventTotal   = $state(0);
  let eventPage    = $state(1);

  let isLoadingEvents = $state(false);
  let isRebuilding    = $state(false);
  let rebuildMsg      = $state('');
  let showConfirm     = $state(false);

  let showExportModal = $state(false);
  let showImportModal = $state(false);
  let exportPassword  = $state('');
  let importPassword  = $state('');
  let importFile      = $state<File | null>(null);
  let isExporting     = $state(false);
  let isImporting     = $state(false);
  let exportMsg       = $state('');
  let importMsg       = $state('');

  const eventTotalPages = $derived(Math.max(1, Math.ceil(eventTotal / PAGE_SIZE)));

  $effect(() => {
    void Promise.all([
      getAggregateVersions(),
      getCheckpoints(),
      getSnapshots(),
      getEventStorePaged(1, PAGE_SIZE),
    ])
      .then(([v, c, s, ep]) => {
        versions    = v;
        checkpoints = c;
        snapshots   = s;
        events      = ep.data ?? [];
        eventTotal  = ep.meta.total_count;
      })
      .catch(err => { error = err instanceof Error ? err.message : '載入失敗'; })
      .finally(() => { isLoading = false; });
  });

  async function gotoEventPage(p: number): Promise<void> {
    if (isLoadingEvents || p < 1 || p > eventTotalPages) return;
    isLoadingEvents = true;
    try {
      const ep = await getEventStorePaged(p, PAGE_SIZE);
      events     = ep.data ?? [];
      eventTotal = ep.meta.total_count;
      eventPage  = p;
    } catch (err) {
      error = err instanceof Error ? err.message : '載入事件失敗';
    } finally {
      isLoadingEvents = false;
    }
  }

  async function rebuild(): Promise<void> {
    showConfirm  = false;
    isRebuilding = true;
    rebuildMsg   = '';
    try {
      const result = await replayProjection();
      rebuildMsg = `已重播 ${result.replayed_count} 筆事件`;
      await Promise.all([
        getAggregateVersions().then(v => { versions = v; }),
        getCheckpoints().then(c => { checkpoints = c; }),
        getSnapshots().then(s => { snapshots = s; }),
        getEventStorePaged(1, PAGE_SIZE).then(ep => {
          events = ep.data ?? []; eventTotal = ep.meta.total_count; eventPage = 1;
        }),
      ]);
    } catch (err) {
      rebuildMsg = err instanceof Error ? err.message : '重建失敗';
    } finally {
      isRebuilding = false;
    }
  }

  async function doExport(): Promise<void> {
    isExporting = true;
    exportMsg   = '';
    try {
      await exportEvents(exportPassword || undefined);
      showExportModal = false;
      exportPassword  = '';
      exportMsg = '匯出成功';
    } catch (err) {
      exportMsg = err instanceof Error ? err.message : '匯出失敗';
    } finally {
      isExporting = false;
    }
  }

  async function doImport(): Promise<void> {
    if (!importFile) return;
    isImporting = true;
    importMsg   = '';
    try {
      const result = await importEvents(importFile, importPassword || undefined);
      showImportModal = false;
      importPassword  = '';
      importFile      = null;
      rebuildMsg = `已匯入 ${result.imported} 筆事件`;
      await Promise.all([
        getAggregateVersions().then(v => { versions = v; }),
        getCheckpoints().then(c => { checkpoints = c; }),
        getSnapshots().then(s => { snapshots = s; }),
        getEventStorePaged(1, PAGE_SIZE).then(ep => {
          events = ep.data ?? []; eventTotal = ep.meta.total_count; eventPage = 1;
        }),
      ]);
    } catch (err) {
      importMsg = err instanceof Error ? err.message : '匯入失敗';
    } finally {
      isImporting = false;
    }
  }
</script>

<div class="content-header">
  <h1 class="content-title">稽核查詢</h1>
  <div class="audit-actions">
    <button
      class="audit-export-btn"
      onclick={() => { showExportModal = true; exportMsg = ''; }}
      disabled={isExporting}
    >匯出事件</button>
    <button
      class="audit-import-btn"
      onclick={() => { showImportModal = true; importMsg = ''; }}
      disabled={isImporting}
    >匯入事件</button>
    <button
      class="audit-rebuild-btn"
      onclick={() => showConfirm = true}
      disabled={isRebuilding}
    >{isRebuilding ? '重建中…' : '重建 Projection'}</button>
  </div>
</div>

{#if rebuildMsg}
  <p class:audit-msg--ok={!rebuildMsg.includes('失敗')} class:audit-msg--err={rebuildMsg.includes('失敗')}>{rebuildMsg}</p>
{/if}
{#if exportMsg}
  <p class:audit-msg--ok={!exportMsg.includes('失敗')} class:audit-msg--err={exportMsg.includes('失敗')}>{exportMsg}</p>
{/if}

{#if error}
  <p class="query-error" role="alert">{error}</p>
{/if}

<div class="tab-nav">
  <button class="tab-btn" class:active={activeTab === 'versions'}    onclick={() => activeTab = 'versions'}>彙總版本</button>
  <button class="tab-btn" class:active={activeTab === 'events'}      onclick={() => activeTab = 'events'}>事件紀錄</button>
  <button class="tab-btn" class:active={activeTab === 'checkpoints'} onclick={() => activeTab = 'checkpoints'}>Checkpoint</button>
  <button class="tab-btn" class:active={activeTab === 'snapshots'}   onclick={() => activeTab = 'snapshots'}>快照</button>
</div>

<!-- ── 彙總版本 ── -->
{#if activeTab === 'versions'}
  <section class="section">
    <div class="table-wrap">
      <table class="data-table" aria-label="彙總版本">
        <thead>
          <tr>
            <th>Aggregate Type</th>
            <th>Aggregate ID</th>
            <th class="num">版本號</th>
          </tr>
        </thead>
        <tbody>
          {#if isLoading}
            <tr><td colspan="3" class="table-empty">載入中…</td></tr>
          {:else if versions.length === 0}
            <tr><td colspan="3" class="table-empty">無資料</td></tr>
          {:else}
            {#each versions as v (`${v.aggregate_type}:${v.aggregate_id}`)}
              <tr>
                <td class="mono" style="font-size:11px">{v.aggregate_type}</td>
                <td class="mono" style="font-size:11px;color:#5c6278">{v.aggregate_id || '—'}</td>
                <td class="mono num">{v.current_version}</td>
              </tr>
            {/each}
          {/if}
        </tbody>
      </table>
    </div>
  </section>

<!-- ── 事件紀錄 ── -->
{:else if activeTab === 'events'}
  <section class="section">
    <div class="table-wrap">
      <table class="data-table" aria-label="事件紀錄">
        <thead>
          <tr>
            <th class="num">Event ID</th>
            <th>日期</th>
            <th>Aggregate Type</th>
            <th>Aggregate ID</th>
            <th class="num">版本</th>
            <th>Event Type</th>
            <th>操作者</th>
          </tr>
        </thead>
        <tbody>
          {#if isLoading || isLoadingEvents}
            <tr><td colspan="7" class="table-empty">載入中…</td></tr>
          {:else if events.length === 0}
            <tr><td colspan="7" class="table-empty">無事件資料</td></tr>
          {:else}
            {#each events as ev (ev.event_id)}
              <tr>
                <td class="mono num">{ev.event_id}</td>
                <td class="mono" style="font-size:11px">{ev.occurred_at}</td>
                <td class="mono" style="font-size:11px">{ev.aggregate_type}</td>
                <td class="mono" style="font-size:11px;color:#5c6278">{ev.aggregate_id || '—'}</td>
                <td class="mono num" style="font-size:11px">{ev.aggregate_version}</td>
                <td style="font-size:11px">{ev.event_type}</td>
                <td style="font-size:11px;color:#5c6278">{ev.updated_by ?? '—'}</td>
              </tr>
            {/each}
          {/if}
        </tbody>
      </table>
    </div>
    <div class="pagination">
      <button class="page-btn" disabled={eventPage === 1} onclick={() => gotoEventPage(eventPage - 1)}>‹ 上一頁</button>
      <span class="page-info">第 {eventPage} / {eventTotalPages} 頁（共 {eventTotal} 筆）</span>
      <button class="page-btn" disabled={eventPage >= eventTotalPages} onclick={() => gotoEventPage(eventPage + 1)}>下一頁 ›</button>
    </div>
  </section>

<!-- ── Checkpoint ── -->
{:else if activeTab === 'checkpoints'}
  <section class="section">
    <div class="table-wrap">
      <table class="data-table" aria-label="Checkpoint">
        <thead>
          <tr>
            <th>Projection 名稱</th>
            <th class="num">最後 Event ID</th>
            <th>更新時間</th>
          </tr>
        </thead>
        <tbody>
          {#if isLoading}
            <tr><td colspan="3" class="table-empty">載入中…</td></tr>
          {:else if checkpoints.length === 0}
            <tr><td colspan="3" class="table-empty">無資料</td></tr>
          {:else}
            {#each checkpoints as cp (cp.projection_name)}
              <tr>
                <td style="font-size:12px">{cp.projection_name}</td>
                <td class="mono num">{cp.last_event_id}</td>
                <td class="mono" style="font-size:11px;color:#5c6278">{cp.updated_at ?? '—'}</td>
              </tr>
            {/each}
          {/if}
        </tbody>
      </table>
    </div>
  </section>

<!-- ── 快照 ── -->
{:else}
  <section class="section">
    <div class="table-wrap">
      <table class="data-table" aria-label="快照">
        <thead>
          <tr>
            <th class="num">Snapshot ID</th>
            <th>Aggregate Type</th>
            <th>Aggregate ID</th>
            <th class="num">版本</th>
            <th>建立時間</th>
          </tr>
        </thead>
        <tbody>
          {#if isLoading}
            <tr><td colspan="5" class="table-empty">載入中…</td></tr>
          {:else if snapshots.length === 0}
            <tr><td colspan="5" class="table-empty">無快照資料</td></tr>
          {:else}
            {#each snapshots as sn (sn.snapshot_id)}
              <tr>
                <td class="mono num">{sn.snapshot_id}</td>
                <td class="mono" style="font-size:11px">{sn.aggregate_type}</td>
                <td class="mono" style="font-size:11px;color:#5c6278">{sn.aggregate_id || '—'}</td>
                <td class="mono num" style="font-size:11px">{sn.at_version}</td>
                <td class="mono" style="font-size:11px;color:#5c6278">{sn.created_at}</td>
              </tr>
            {/each}
          {/if}
        </tbody>
      </table>
    </div>
  </section>
{/if}

<!-- ── 重建確認 Modal ─────────────────────── -->
{#if showConfirm}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) showConfirm = false; }}>
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="confirm-modal-title">
      <div class="modal-header">
        <h2 class="modal-title" id="confirm-modal-title">確認重建 Projection</h2>
        <button class="modal-close" onclick={() => showConfirm = false} aria-label="關閉">×</button>
      </div>
      <div class="modal-body">
        <p style="font-size:13px;color:#c0bdb4;line-height:1.7;margin:0 0 8px;">
          此操作將清除當前商戶的所有 projection 資料，並全量重播事件重建。
        </p>
        <p style="font-size:12px;color:#c07070;margin:0;">
          此動作無法復原，請確認後繼續。
        </p>
      </div>
      <div class="modal-footer" style="padding:0;margin-top:8px;">
        <button class="btn-ghost" onclick={() => showConfirm = false}>取消</button>
        <button class="btn-primary" onclick={() => void rebuild()}>確認重建</button>
      </div>
    </div>
  </div>
{/if}

<!-- ── 匯出 Modal ─────────────────────────── -->
{#if showExportModal}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) { showExportModal = false; exportPassword = ''; } }}>
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="export-modal-title">
      <div class="modal-header">
        <h2 class="modal-title" id="export-modal-title">匯出事件資料</h2>
        <button class="modal-close" onclick={() => { showExportModal = false; exportPassword = ''; }} aria-label="關閉">×</button>
      </div>
      <div class="modal-body">
        <p style="font-size:13px;color:#c0bdb4;line-height:1.7;margin:0 0 12px;">
          不填密碼則匯出為未加密 JSON；填入密碼後以 AES-256-GCM 加密。
        </p>
        <div class="form-group">
          <label class="form-label" for="export-password">密碼（選填）</label>
          <input
            id="export-password"
            class="form-input"
            type="password"
            placeholder="留空則不加密"
            bind:value={exportPassword}
            disabled={isExporting}
          />
        </div>
        {#if exportMsg && showExportModal}
          <p class:audit-msg--err={exportMsg.includes('失敗')} style="margin:8px 0 0;">{exportMsg}</p>
        {/if}
      </div>
      <div class="modal-footer" style="padding:0;margin-top:8px;">
        <button class="btn-ghost" onclick={() => { showExportModal = false; exportPassword = ''; }} disabled={isExporting}>取消</button>
        <button class="btn-primary" onclick={() => void doExport()} disabled={isExporting}>
          {isExporting ? '匯出中…' : '確認匯出'}
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- ── 匯入 Modal ─────────────────────────── -->
{#if showImportModal}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) { showImportModal = false; importPassword = ''; importFile = null; } }}>
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="import-modal-title">
      <div class="modal-header">
        <h2 class="modal-title" id="import-modal-title">匯入事件資料</h2>
        <button class="modal-close" onclick={() => { showImportModal = false; importPassword = ''; importFile = null; }} aria-label="關閉">×</button>
      </div>
      <div class="modal-body">
        <p style="font-size:12px;color:#c07070;line-height:1.7;margin:0 0 12px;">
          此操作將清除所有現有事件與快照，以匯入資料完整替換，無法復原。
        </p>
        <div class="form-group" style="margin-bottom:10px;">
          <label class="form-label" for="import-file">事件檔案（.json）</label>
          <input
            id="import-file"
            class="form-input"
            type="file"
            accept=".json"
            disabled={isImporting}
            onchange={(e) => { const t = e.currentTarget as HTMLInputElement; importFile = t.files?.[0] ?? null; }}
          />
        </div>
        <div class="form-group">
          <label class="form-label" for="import-password">密碼（加密檔才需填）</label>
          <input
            id="import-password"
            class="form-input"
            type="password"
            placeholder="未加密檔案請留空"
            bind:value={importPassword}
            disabled={isImporting}
          />
        </div>
        {#if importMsg}
          <p class:audit-msg--err={importMsg.includes('失敗')} style="margin:8px 0 0;">{importMsg}</p>
        {/if}
      </div>
      <div class="modal-footer" style="padding:0;margin-top:8px;">
        <button class="btn-ghost" onclick={() => { showImportModal = false; importPassword = ''; importFile = null; }} disabled={isImporting}>取消</button>
        <button class="btn-primary" onclick={() => void doImport()} disabled={isImporting || !importFile}>
          {isImporting ? '匯入中…' : '確認匯入'}
        </button>
      </div>
    </div>
  </div>
{/if}
