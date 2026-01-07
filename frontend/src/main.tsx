import React from 'react'
import { createRoot } from 'react-dom/client'
import App from './App'
import './i18n/config'
import './style.less'

const container = document.getElementById('root')
const root = createRoot(container!)

root.render(
    <React.StrictMode>
        <App />
    </React.StrictMode>
)

export { DataContext } from './context'
export type { GlobalData } from './context'

