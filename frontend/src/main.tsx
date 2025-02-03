import React, { createContext } from 'react'
import { createRoot } from 'react-dom/client'
import App from './App'
import { AppStatus } from './AppStatus'
import './style.less'

const container = document.getElementById('root')
const root = createRoot(container!)

export type GlobalData = {
    status: AppStatus
    updateStatus: (d: AppStatus) => void
}

export const DataContext = createContext({} as GlobalData)

root.render(
    <React.StrictMode>
        <App />
    </React.StrictMode>
)
