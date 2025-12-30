import { createContext } from 'react'
import { AppStatus } from './AppStatus'

export type GlobalData = {
    status: AppStatus
    updateStatus: (d: AppStatus) => void
}

export const DataContext = createContext({} as GlobalData)
