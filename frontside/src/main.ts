import { mount } from 'svelte'
import './app.css'
import './styles/login.css'
import './styles/home.css'
import App from './App.svelte'

mount(App, {
  target: document.getElementById('app')!,
})
