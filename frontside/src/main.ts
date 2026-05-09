import { mount } from 'svelte'
import './app.css'
import './styles/login.css'
import './styles/home.css'
import './styles/views/accounts.css'
import './styles/views/ledger.css'
import './styles/views/journal-entry.css'
import './styles/views/investment.css'
import './styles/views/period.css'
import './styles/views/installment.css'
import './styles/components/modal.css'
import App from './App.svelte'

mount(App, {
  target: document.getElementById('app')!,
})
