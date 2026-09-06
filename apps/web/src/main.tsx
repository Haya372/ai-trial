import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'

const rootElement = document.getElementById('root')
if (!rootElement) {
  throw new Error(
    'Mount point #root not found. Check that index.html contains <div id="root"></div>.',
  )
}
createRoot(rootElement).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
