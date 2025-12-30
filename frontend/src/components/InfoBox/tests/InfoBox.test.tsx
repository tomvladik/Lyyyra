import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { InfoBox } from '../index';
import { DataContext } from '../../../context';
import { describe, it, expect, vi, beforeEach } from 'vitest';

describe('<InfoBox />', () => {
  const mockLoadSongs = vi.fn();
  const mockSetFilter = vi.fn();
  const mockUpdateStatus = vi.fn();

  const defaultStatus = {
    DatabaseReady: false,
    SongsReady: false,
    WebResourcesReady: false,
    IsProgress: false,
    LastSave: '',
    SearchPattern: '',
    Sorting: 'entry' as const,
  };

  const renderInfoBox = (status = defaultStatus) => {
    return render(
      <DataContext.Provider value={{ status, updateStatus: mockUpdateStatus }}>
        <InfoBox loadSongs={mockLoadSongs} setFilter={mockSetFilter} />
      </DataContext.Provider>
    );
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders initial state with download button', () => {
    renderInfoBox();
    expect(screen.getByText(/Zpěvník není inicializován/)).toBeInTheDocument();
    expect(screen.getByText('Stáhnout data z internetu')).toBeInTheDocument();
  });

  it('hides button when database and songs are ready', () => {
    const readyStatus = {
      ...defaultStatus,
      DatabaseReady: true,
      SongsReady: true,
    };
    renderInfoBox(readyStatus);
    
    const button = screen.queryByText('Stáhnout data z internetu');
    expect(button).not.toBeInTheDocument();
  });

  it('calls setFilter after debounce delay', async () => {
    const statusWithPattern = {
      ...defaultStatus,
      SearchPattern: 'test search',
    };
    renderInfoBox(statusWithPattern);

    await waitFor(() => {
      expect(mockSetFilter).toHaveBeenCalledWith('test search');
    }, { timeout: 600 });
  });

  it('shows button when only songs are ready', () => {
    const songsReadyStatus = {
      ...defaultStatus,
      SongsReady: true,
      DatabaseReady: false,
    };
    renderInfoBox(songsReadyStatus);
    
    const button = screen.getByRole('button');
    expect(button).toBeInTheDocument();
  });
});

