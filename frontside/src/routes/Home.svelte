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
  import Dashboard    from '../views/Dashboard.svelte';
  import Accounts     from '../views/Accounts.svelte';
  import Ledger       from '../views/Ledger.svelte';
  import JournalEntry from '../views/JournalEntry.svelte';
  import Period       from '../views/Period.svelte';
  import Investment   from '../views/Investment.svelte';
  import Reports      from '../views/Reports.svelte';
  import Settings     from '../views/Settings.svelte';

  interface MenuItem {
    label: string;
    path:  string;
  }

  const menuItems: MenuItem[] = [
    { label: '儀表板', path: '/home/dashboard' },
    { label: '傳票管理', path: '/home/journal-entry' },
    { label: '會計科目', path: '/home/accounts' },
    { label: '帳戶管理', path: '/home/ledger' },
    { label: '會計期間', path: '/home/period' },
    { label: '財務報表', path: '/home/reports' },
    { label: '投資管理', path: '/home/investment' },
    { label: '系統設定', path: '/home/settings' },
  ];

  const routes = {
    '/home':                Dashboard,
    '/home/dashboard':      Dashboard,
    '/home/journal-entry':  JournalEntry,
    '/home/accounts':       Accounts,
    '/home/ledger':         Ledger,
    '/home/period':         Period,
    '/home/reports':        Reports,
    '/home/investment':     Investment,
    '/home/settings':       Settings,
  };

  let currentPath = $state(window.location.hash.replace(/^#/, '') || '/home');

  $effect(() => {
    const onHashChange = (): void => {
      currentPath = window.location.hash.replace(/^#/, '') || '/home';
    };
    window.addEventListener('hashchange', onHashChange);
    return () => window.removeEventListener('hashchange', onHashChange);
  });

  function isActive(path: string): boolean {
    if (currentPath === '/home') return path === '/home/dashboard';
    return currentPath.startsWith(path);
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

    <!-- Left sidebar -->
    <nav class="app-sidebar" aria-label="主選單">
      <ul class="sidebar-nav">
        {#each menuItems as item, i}
          {#if i === 7}
            <hr class="sidebar-sep" />
          {/if}
          <li>
            <a
              href="#{item.path}"
              class="sidebar-item"
              class:active={isActive(item.path)}
              aria-current={isActive(item.path) ? 'page' : undefined}
            >
              {item.label}
            </a>
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
