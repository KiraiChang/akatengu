<script lang="ts">
  import Router         from 'svelte-spa-router'
  import Home           from './routes/Home.svelte'
  import Login          from './routes/Login.svelte'
  import MerchantSelect from './routes/MerchantSelect.svelte'
  import NotFound       from './routes/NotFound.svelte'

  const routes = {
    '/':                Login,
    '/merchant-select': MerchantSelect,
    '/home':            Home,
    '/home/*':          Home,
    '*':                NotFound,
  }
</script>

<Router {routes} />
