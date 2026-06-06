import React from 'react'
import {createRoot} from 'react-dom/client'
import './style.css'
import App from './App'

const container = document.getElementById('root')

function escapeHtml(value: string) {
    return value
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#039;')
}

function showStartupError(error: unknown) {
    const message = error instanceof Error ? error.message : String(error)
    if (container) {
        container.innerHTML = `<div style="min-height:100vh;padding:24px;font-family:Segoe UI,Arial,sans-serif;background:#fff5f5;color:#7f1d1d"><h1 style="margin:0 0 8px;font-size:20px">md-preview failed to start</h1><pre style="white-space:pre-wrap;margin:0">${escapeHtml(message)}</pre></div>`
    }
}

window.addEventListener('error', (event) => showStartupError(event.error || event.message))
window.addEventListener('unhandledrejection', (event) => showStartupError(event.reason))

try {
    if (!container) {
        throw new Error('Root element was not found')
    }

    const root = createRoot(container)

    root.render(
        <React.StrictMode>
            <App/>
        </React.StrictMode>
    )
} catch (error) {
    showStartupError(error)
}
