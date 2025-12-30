import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { InfoBox } from '../index';
import { DataContext } from '../../../context';
import { createMockStatus } from '../../../test/testHelpers';
import { describe, it, expect, vi, beforeEach } from 'vitest';

describe('<InfoBox />', () => {
  const mockLoadSongs = vi.fn();
  const mockSetFilter = vi.fn();
  const mockUpdateStatus = vi.fn();

  const renderInfoBox = (statusOverrides = {}) => {
    const status = createMockStatus(statusOverrides);
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
    renderInfoBox({
      DatabaseReady: true,
      SongsReady: true,
    });
    
    const button = screen.queryByText('Stáhnout data z internetu');
    expect(button).not.toBeInTheDocument();
  });

  it('calls setFilter after debounce delay', async () => {
    renderInfoBox({
      SearchPattern: 'test search',
    });

    await waitFor(() => {
      expect(mockSetFilter).toHaveBeenCalledWith('test search');
    }, { timeout: 600 });
  });

  it('shows button when only songs are ready', () => {
    renderInfoBox({
      SongsReady: true,
    });
    
    const button = screen.getByRole('button');
    expect(button).toBeInTheDocument();
  });
});

