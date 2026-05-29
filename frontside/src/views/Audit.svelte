<script lang="ts">
  import {
    getAggregateVersions, getEventStorePaged,
    getCheckpoints, getSnapshots, replayProjection,
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
</script>

<div class="content-header">
  <h1 class="content-title">稽核查詢</h1>
  <button
    class="audit-rebuild-btn"
    onclick={() => showConfirm = true}
    disabled={isRebuilding}
  >{isRebuilding ? '重建中…' : '重建 Projection'}</button>
</div>

{#if rebuildMsg}
  <p class:audit-msg--ok={!rebuildMsg.includes('失敗')} class:audit-msg--err={rebuildMsg.includes('失敗')}>{rebuildMsg}</p>
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
