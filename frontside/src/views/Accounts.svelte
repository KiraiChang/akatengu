<script lang="ts">
  import { getAccountPaged } from '../api/account';
  import type { Account } from '../types/account';
  import type { PaginatedMeta } from '../types/pagination';

  const PAGE_SIZE = 20;

  const ACCOUNT_TYPE_LABELS: Record<string, string> = {
    ASSET:     '資產',
    LIABILITY: '負債',
    EQUITY:    '權益',
    INCOME:    '收入',
    EXPENSE:   '費用',
  };

  const NORMAL_BALANCE_LABELS: Record<string, string> = {
    DEBIT:  '借',
    CREDIT: '貸',
  };

  let page      = $state(1);
  let accounts  = $state<Account[]>([]);
  let meta      = $state<PaginatedMeta | null>(null);
  let isLoading = $state(false);
  let error     = $state('');

  $effect(() => {
    void load(page);
  });

  async function load(p: number): Promise<void> {
    isLoading = true;
    error = '';
    try {
      const result = await getAccountPaged({ page: p, pageSize: PAGE_SIZE });
      accounts = result.data;
      meta = result.meta;
    } catch (err) {
      error = err instanceof Error ? err.message : '查詢失敗，請稍後再試。';
    } finally {
      isLoading = false;
    }
  }

  function labelOf(map: Record<string, string>, key: string): string {
    return map[key] ?? key;
  }
</script>

<div class="content-header">
  <h1 class="content-title">帳目查詢</h1>
  {#if meta}
    <span class="content-date">共 {meta.total_count} 筆</span>
  {/if}
</div>

{#if error}
  <p class="query-error" role="alert">{error}</p>
{/if}

<section class="section">
  <header class="section-header">
    <h2 class="section-title">科目列表</h2>
    {#if isLoading}
      <span class="query-loading">
        <span class="spinner" aria-hidden="true"></span>
        載入中
      </span>
    {/if}
  </header>

  <div class="table-wrap">
    <table class="data-table" aria-label="科目列表">
      <thead>
        <tr>
          <th>科目編號</th>
          <th>科目名稱</th>
          <th>類型</th>
          <th>正常餘額</th>
          <th>幣別</th>
          <th>摘要科目</th>
          <th>狀態</th>
          <th>備註</th>
        </tr>
      </thead>
      <tbody>
        {#if isLoading && accounts.length === 0}
          <tr><td colspan="8" class="table-empty">載入中...</td></tr>
        {:else if accounts.length === 0}
          <tr><td colspan="8" class="table-empty">無資料</td></tr>
        {:else}
          {#each accounts as account (account.account_id)}
            <tr>
              <td class="mono">{account.account_id}</td>
              <td>{account.name}</td>
              <td>{labelOf(ACCOUNT_TYPE_LABELS, account.type)}</td>
              <td>{labelOf(NORMAL_BALANCE_LABELS, account.normal_balance)}</td>
              <td>{account.currency}</td>
              <td>
                {#if account.is_summary}
                  <span class="badge approved">是</span>
                {:else}
                  <span class="badge-plain">否</span>
                {/if}
              </td>
              <td>
                {#if account.is_active}
                  <span class="badge approved">啟用</span>
                {:else}
                  <span class="badge pending">停用</span>
                {/if}
              </td>
              <td class="note-cell">{account.note ?? '—'}</td>
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  </div>

  {#if meta && meta.total_pages > 1}
    <div class="pagination">
      <button
        class="page-btn"
        onclick={() => { page -= 1; }}
        disabled={page <= 1 || isLoading}
        aria-label="上一頁"
      >←</button>
      <span class="page-info">
        第 <span class="page-num">{page}</span> 頁 &nbsp;/&nbsp; 共 {meta.total_pages} 頁
      </span>
      <button
        class="page-btn"
        onclick={() => { page += 1; }}
        disabled={page >= meta.total_pages || isLoading}
        aria-label="下一頁"
      >→</button>
    </div>
  {/if}
</section>
