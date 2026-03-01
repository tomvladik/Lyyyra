import { createContext } from 'react'
import { AppStatus } from './AppStatus'

export type GlobalData = {
    status: AppStatus
    updateStatus: (d: Partial<AppStatus>) => void
    sourceFilter: string
    setSourceFilter: (f: string) => void
}

export const DataContext = createContext({} as GlobalData)
