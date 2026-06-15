<svelte:head>
  <link rel="preconnect" href="https://fonts.googleapis.com" />
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="" />
  <link
    href="https://fonts.googleapis.com/css2?family=Cormorant+Garamond:ital,wght@0,300;0,500;1,300&family=JetBrains+Mono:wght@300;400;500&display=swap"
    rel="stylesheet"
  />
</svelte:head>

<script lang="ts">
  import Router from 'svelte-spa-router';
  import { authStore } from '../stores/auth.svelte';
  import Dashboard        from '../views/Dashboard.svelte';
  import Accounts         from '../views/Accounts.svelte';
  import Ledger           from '../views/Ledger.svelte';
  import JournalEntry     from '../views/JournalEntry.svelte';
  import Period           from '../views/Period.svelte';
  import Investment       from '../views/Investment.svelte';
  import Installment      from '../views/Installment.svelte';
  import BalanceSheet     from '../views/BalanceSheet.svelte';
  import IncomeStatement  from '../views/IncomeStatement.svelte';
  import CashFlowStatement  from '../views/CashFlowStatement.svelte';
  import EquityStatement    from '../views/EquityStatement.svelte';
  import Settings           from '../views/Settings.svelte';
  import Prepaid            from '../views/Prepaid.svelte';
  import FixedAsset         from '../views/FixedAsset.svelte';
  import Template           from '../views/Template.svelte';
  import AccountAnalysis    from '../views/AccountAnalysis.svelte';
  import Audit              from '../views/Audit.svelte';
  import ExchangeRate       from '../views/ExchangeRate.svelte';
  import Merchants             from '../views/Merchants.svelte';
  import BankStatementTemplate from '../views/BankStatementTemplate.svelte';
  import BankPdfTemplate       from '../views/BankPdfTemplate.svelte';
  import BankStatement         from '../views/BankStatement.svelte';
  import BankStatementResult   from '../views/BankStatementResult.svelte';
  import BankStatementReview   from '../views/BankStatementReview.svelte';
  import CFCategoryReview      from '../views/CFCategoryReview.svelte';

  interface SubMenuItem { label: string; path: string; }
  interface MenuItem    { label: string; path: string; key?: string; children?: SubMenuItem[]; }

  const menuItems: MenuItem[] = [
    { label: '儀表板',   path: '/home/dashboard' },
    { label: '傳票管理', path: '/home/journal-entry' },
    { label: '會計科目', path: '/home/accounts' },
    { label: '帳戶管理', path: '/home/ledger' },
    { label: '會計期間', path: '/home/period' },
    { label: '財務報表', path: '/home/reports', key: 'reports', children: [
      { label: '資產負債表', path: '/home/reports/balance-sheet' },
      { label: '損益表',     path: '/home/reports/income-statement' },
      { label: '現金流量表', path: '/home/reports/cash-flow' },
      { label: '權益變動表', path: '/home/reports/equity-statement' },
      { label: 'CF 分類審視', path: '/home/reports/cf-category' },
    ]},
    { label: '科目分析', path: '/home/account-analysis' },
    { label: '投資管理', path: '/home/investment' },
    { label: '銀行對帳', path: '/home/bank-statement', key: 'bank-statement', children: [
      { label: '匯入清單',     path: '/home/bank-statement' },
      { label: 'CSV 範本管理', path: '/home/bank-statement/template' },
      { label: 'PDF 範本管理', path: '/home/bank-statement/pdf-template' },
    ]},
    { label: '分期管理', path: '/home/installment', key: 'amortization', children: [
      { label: '分期付款', path: '/home/installment' },
      { label: '預付費用', path: '/home/prepaid' },
      { label: '固定資產', path: '/home/fixed-asset' },
    ]},
    { label: '系統設定', path: '/home/settings', key: 'settings', children: [
      { label: '系統科目對應', path: '/home/settings' },
      { label: '範本管理',     path: '/home/settings/template' },
      { label: '稽核查詢',     path: '/home/settings/audit' },
      { label: '匯率管理',     path: '/home/settings/exchange-rate' },
      { label: '商戶管理',     path: '/home/merchants' },
    ]},
  ];

  const routes = {
    '/home':                          Dashboard,
    '/home/dashboard':                Dashboard,
    '/home/journal-entry':            JournalEntry,
    '/home/accounts':                 Accounts,
    '/home/ledger':                   Ledger,
    '/home/period':                   Period,
    '/home/reports':                  BalanceSheet,
    '/home/reports/balance-sheet':    BalanceSheet,
    '/home/reports/income-statement': IncomeStatement,
    '/home/reports/cash-flow':        CashFlowStatement,
    '/home/reports/equity-statement': EquityStatement,
    '/home/reports/cf-category':      CFCategoryReview,
    '/home/account-analysis':         AccountAnalysis,
    '/home/investment':                Investment,
    '/home/bank-statement':                    BankStatement,
    '/home/bank-statement/:import_id/result':  BankStatementResult,
    '/home/bank-statement/:import_id/review':  BankStatementReview,
    '/home/bank-statement/template':           BankStatementTemplate,
    '/home/bank-statement/pdf-template':       BankPdfTemplate,
    '/home/installment':              Installment,
    '/home/prepaid':                  Prepaid,
    '/home/fixed-asset':              FixedAsset,
    '/home/settings':                 Settings,
    '/home/settings/template':        Template,
    '/home/settings/audit':           Audit,
    '/home/settings/exchange-rate':   ExchangeRate,
    '/home/merchants':                Merchants,
  };

  let currentPath     = $state(window.location.hash.replace(/^#/, '') || '/home');
  let sidebarOpen     = $state(false);
  let expandedParents = $state(new Set<string>());

  const parentChildPaths: Record<string, string[]> = {
    reports:          ['/home/reports'],
    'bank-statement': ['/home/bank-statement'],
    settings:         ['/home/settings', '/home/settings/audit', '/home/settings/exchange-rate', '/home/merchants'],
    amortization:     ['/home/installment', '/home/prepaid', '/home/fixed-asset'],
  };

  function isExpanded(key: string): boolean {
    return expandedParents.has(key) || (parentChildPaths[key] ?? []).some(p => currentPath.startsWith(p));
  }

  $effect(() => {
    const onHashChange = (): void => {
      currentPath = window.location.hash.replace(/^#/, '') || '/home';
      sidebarOpen = false;
    };
    window.addEventListener('hashchange', onHashChange);
    return () => window.removeEventListener('hashchange', onHashChange);
  });

  function isActive(path: string): boolean {
    if (currentPath === '/home') return path === '/home/dashboard';
    return currentPath.startsWith(path);
  }

  function toggleParent(key: string): void {
    const next = new Set(expandedParents);
    if (next.has(key)) next.delete(key); else next.add(key);
    expandedParents = next;
  }

  function logout(): void {
    authStore.clearToken();
    window.location.hash = '#/';
  }
</script>

<div class="app-shell">

  <!-- ── Header ──────────────────────────── -->
  <header class="app-header">
    <div class="header-brand">
      <button
        class="header-menu-btn"
        aria-label={sidebarOpen ? '關閉選單' : '開啟選單'}
        aria-expanded={sidebarOpen}
        onclick={() => sidebarOpen = !sidebarOpen}
      >☰</button>
      <div class="header-brand-mark" aria-hidden="true">赤</div>
      <div class="header-brand-text">
        <span class="header-brand-name">AKATENGU</span>
        <span class="header-brand-sub">會計管理系統</span>
      </div>
    </div>
    <div class="header-right">
      <div class="header-user">
        <span class="header-user-dot" aria-hidden="true"></span>
        <span class="header-user-label">管理員</span>
      </div>
      <button class="header-logout" onclick={logout}>登 出</button>
    </div>
  </header>

  <!-- ── Body ───────────────────────────── -->
  <div class="app-body">

    <!-- Mobile backdrop -->
    {#if sidebarOpen}
      <div
        class="sidebar-backdrop"
        role="presentation"
        onclick={() => sidebarOpen = false}
      ></div>
    {/if}

    <!-- Left sidebar -->
    <nav class="app-sidebar" class:sidebar-open={sidebarOpen} aria-label="主選單">
      <ul class="sidebar-nav">
        {#each menuItems as item, i}
          {#if i === 9}
            <hr class="sidebar-sep" />
          {/if}
          <li>
            {#if item.children}
              {@const key = item.key ?? item.path}
              {@const expanded = isExpanded(key)}
              <button
                class="sidebar-item sidebar-parent-btn"
                class:active={isActive(item.path)}
                aria-expanded={expanded}
                onclick={() => toggleParent(key)}
              >
                <span>{item.label}</span>
                <span class="sidebar-expand-icon" aria-hidden="true">
                  {expanded ? '▾' : '▸'}
                </span>
              </button>
              {#if expanded}
                <ul class="sidebar-sub-nav">
                  {#each item.children as child}
                    <li>
                      <a
                        href="#{child.path}"
                        class="sidebar-sub-item"
                        class:active={currentPath === child.path}
                        aria-current={currentPath === child.path ? 'page' : undefined}
                      >{child.label}</a>
                    </li>
                  {/each}
                </ul>
              {/if}
            {:else}
              <a
                href="#{item.path}"
                class="sidebar-item"
                class:active={isActive(item.path)}
                aria-current={isActive(item.path) ? 'page' : undefined}
              >
                {item.label}
              </a>
            {/if}
          </li>
        {/each}
      </ul>
      <div class="sidebar-bottom">
        <span class="sidebar-ver">v0.1.0-alpha</span>
      </div>
    </nav>

    <!-- Main content (swapped by router) -->
    <main class="app-content">
      <Router {routes} />
    </main>
  </div>

  <!-- ── Footer ─────────────────────────── -->
  <footer class="app-footer">
    <span>© 2026 AKATENGU 會計管理系統</span>
    <span class="footer-sep" aria-hidden="true">|</span>
    <span>版本 0.1.0</span>
  </footer>

</div>
