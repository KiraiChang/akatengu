<script lang="ts">
  import type { Account } from '../types/account';

  interface BreadcrumbItem {
    id:   string | null;
    name: string;
  }

  interface Props {
    accounts:    Account[];
    value:       string;
    placeholder?: string;
    disabled?:   boolean;
    onselect:    (id: string) => void;
  }

  let {
    accounts,
    value,
    placeholder = '選擇會計科目…',
    disabled = false,
    onselect,
  }: Props = $props();

  let open            = $state(false);
  let searchQuery     = $state('');
  let currentParentId = $state<string | null>(null);
  let breadcrumb      = $state<BreadcrumbItem[]>([{ id: null, name: '根目錄' }]);
  let rootEl          = $state<HTMLDivElement | null>(null);
  let searchEl        = $state<HTMLInputElement | null>(null);

  // 查詢選定的科目
  const selected = $derived(accounts.find(a => a.account_id === value) ?? null);

  // 當前層級的科目
  const currentItems = $derived(
    accounts.filter(a => (a.parent_id ?? null) === currentParentId),
  );

  // 文字搜尋結果（含父科目名稱供顯示）
  const searchResults = $derived((() => {
    const q = searchQuery.trim().toLowerCase();
    if (!q) return null;
    return accounts
      .filter(a =>
        a.account_id.toLowerCase().includes(q) ||
        a.name.toLowerCase().includes(q),
      )
      .slice(0, 40);
  })());

  // 快速查詢父科目名稱
  const accountNameMap = $derived(
    new Map(accounts.map(a => [a.account_id, a.name])),
  );

  // 點擊外部關閉
  $effect(() => {
    if (!open) return;
    function onOutside(e: MouseEvent): void {
      if (rootEl && !rootEl.contains(e.target as Node)) close();
    }
    document.addEventListener('mousedown', onOutside);
    return () => document.removeEventListener('mousedown', onOutside);
  });

  // 開啟時自動 focus 搜尋框
  $effect(() => {
    if (open) {
      setTimeout(() => searchEl?.focus(), 30);
    }
  });

  function toggle(): void {
    if (disabled) return;
    if (open) { close(); } else { open = true; }
  }

  function close(): void {
    open           = false;
    searchQuery    = '';
    currentParentId = null;
    breadcrumb     = [{ id: null, name: '根目錄' }];
  }

  function select(account: Account): void {
    onselect(account.account_id);
    close();
  }

  function clearValue(e: MouseEvent): void {
    e.stopPropagation();
    onselect('');
  }

  // 往子層鑽入
  function drillDown(e: MouseEvent, account: Account): void {
    e.stopPropagation();
    breadcrumb      = [...breadcrumb, { id: account.account_id, name: `${account.account_id} ${account.name}` }];
    currentParentId = account.account_id;
    searchQuery     = '';
  }

  // 麵包屑跳回指定層
  function navigateTo(index: number): void {
    breadcrumb      = breadcrumb.slice(0, index + 1);
    currentParentId = breadcrumb[index].id;
    searchQuery     = '';
  }

  // 從搜尋結果跳到父科目所在層
  function jumpToParent(e: MouseEvent, parentId: string): void {
    e.stopPropagation();
    // 建立到該父科目的麵包屑路徑
    const path: BreadcrumbItem[] = [{ id: null, name: '根目錄' }];
    const ancestors: Account[] = [];
    let cur: string | null = parentId;
    while (cur) {
      const a = accounts.find(x => x.account_id === cur);
      if (!a) break;
      ancestors.unshift(a);
      cur = a.parent_id ?? null;
    }
    for (const a of ancestors) {
      path.push({ id: a.account_id, name: `${a.account_id} ${a.name}` });
    }
    breadcrumb      = path;
    currentParentId = parentId;
    searchQuery     = '';
  }

  const displayItems = $derived(searchResults ?? currentItems);
  const isSearchMode = $derived(searchQuery.trim().length > 0);

  const listboxId = `acct-lb-${Math.random().toString(36).slice(2, 7)}`;
</script>

<div class="acct-select" bind:this={rootEl}>
  <!-- Trigger button -->
  <div
    class="acct-trigger"
    class:acct-trigger--disabled={disabled}
    class:acct-trigger--open={open}
    onclick={toggle}
    role="combobox"
    aria-expanded={open}
    aria-controls={listboxId}
    aria-haspopup="listbox"
    tabindex={disabled ? -1 : 0}
    onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); toggle(); } }}
  >
    {#if selected}
      <span class="acct-trigger-value">
        <span class="acct-trigger-id">{selected.account_id}</span>
        <span class="acct-trigger-name">{selected.name}</span>
      </span>
    {:else}
      <span class="acct-trigger-placeholder">{placeholder}</span>
    {/if}
    <span class="acct-trigger-icons">
      {#if value && !disabled}
        <button
          class="acct-clear"
          type="button"
          onclick={clearValue}
          tabindex="-1"
          aria-label="清除"
        >×</button>
      {/if}
      <span class="acct-chevron" aria-hidden="true">{open ? '▲' : '▼'}</span>
    </span>
  </div>

  <!-- Dropdown panel -->
  {#if open}
    <div class="acct-panel" role="dialog">
      <!-- Search input -->
      <div class="acct-search-wrap">
        <input
          bind:this={searchEl}
          class="acct-search"
          type="text"
          bind:value={searchQuery}
          placeholder="搜尋科目編號或名稱…"
          autocomplete="off"
        />
      </div>

      <!-- Breadcrumb (tree mode only) -->
      {#if !isSearchMode}
        <div class="acct-breadcrumb" aria-label="目前位置">
          {#each breadcrumb as crumb, i (crumb.id ?? '__root__')}
            {#if i > 0}<span class="acct-breadcrumb-sep" aria-hidden="true">›</span>{/if}
            <button
              class="acct-breadcrumb-item"
              class:acct-breadcrumb-item--current={i === breadcrumb.length - 1}
              type="button"
              onclick={() => navigateTo(i)}
              disabled={i === breadcrumb.length - 1}
            >{crumb.name}</button>
          {/each}
        </div>
      {/if}

      <!-- Account list -->
      <ul class="acct-list" role="listbox" id={listboxId}>
        {#if displayItems.length === 0}
          <li class="acct-empty">無符合科目</li>
        {:else}
          {#each displayItems as account (account.account_id)}
            <li
              class="acct-item"
              class:acct-item--selected={account.account_id === value}
              role="option"
              aria-selected={account.account_id === value}
            >
              <button
                class="acct-item-main"
                type="button"
                onclick={() => select(account)}
              >
                <span class="acct-item-id">{account.account_id}</span>
                <span class="acct-item-name">{account.name}</span>
              </button>
              {#if isSearchMode && account.parent_id}
                <span class="acct-item-parent">
                  ↑
                  <button
                    class="acct-item-parent-link"
                    type="button"
                    onclick={(e) => jumpToParent(e, account.parent_id!)}
                  >{accountNameMap.get(account.parent_id) ?? account.parent_id}</button>
                </span>
              {/if}
              {#if account.has_child && !isSearchMode}
                <button
                  class="acct-drill"
                  type="button"
                  onclick={(e) => drillDown(e, account)}
                  aria-label="展開 {account.name} 的子科目"
                >›</button>
              {/if}
            </li>
          {/each}
        {/if}
      </ul>
    </div>
  {/if}
</div>
