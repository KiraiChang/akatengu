<script lang="ts">
  import SysAccountConfig        from '../components/SysAccountConfig.svelte';
  import FixedAssetCategoryConfig from '../components/FixedAssetCategoryConfig.svelte';
  import PrepaidCategoryConfig    from '../components/PrepaidCategoryConfig.svelte';

  type SettingsTab = 'sys' | 'fa-cat' | 'pp-cat';
  let activeTab = $state<SettingsTab>('sys');
</script>

<div class="content-header">
  <h1 class="content-title">系統設定</h1>
</div>

<div class="settings-tabs">
  <button
    class="settings-tab"
    class:settings-tab--active={activeTab === 'sys'}
    onclick={() => activeTab = 'sys'}
  >系統科目對應</button>
  <button
    class="settings-tab"
    class:settings-tab--active={activeTab === 'fa-cat'}
    onclick={() => activeTab = 'fa-cat'}
  >固定資產類別</button>
  <button
    class="settings-tab"
    class:settings-tab--active={activeTab === 'pp-cat'}
    onclick={() => activeTab = 'pp-cat'}
  >預付費用類別</button>
</div>

{#if activeTab === 'sys'}
  <SysAccountConfig />
{:else if activeTab === 'fa-cat'}
  <FixedAssetCategoryConfig />
{:else}
  <PrepaidCategoryConfig />
{/if}
