import '@testing-library/jest-dom'
import { vi } from 'vitest'

// Mock Wails runtime
vi.mock('../../wailsjs/go/app/App', () => ({
  GetSongs2: vi.fn(),
  GetSongAuthors: vi.fn(),
  DownloadSongBase: vi.fn(),
  GetFilteredSongs: vi.fn(),
}))
