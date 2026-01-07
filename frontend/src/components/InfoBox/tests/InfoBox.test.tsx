import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { DataContext } from '../../../context';
import { createMockStatus } from '../../../test/testHelpers';
import { InfoBox } from '../index';

describe('<InfoBox />', () => {
  const mockLoadSongs = vi.fn();
  const mockUpdateStatus = vi.fn();

  const renderInfoBox = (statusOverrides = {}) => {
    const status = createMockStatus(statusOverrides);
    return render(
      <DataContext.Provider value={{ status, updateStatus: mockUpdateStatus }}>
        <InfoBox loadSongs={mockLoadSongs} />
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

  it('shows import message and button when only songs are ready', async () => {
    renderInfoBox({
      SongsReady: true,
    });

    expect(await screen.findByText(/Data jsou stažena/)).toBeInTheDocument();
    const button = screen.getByRole('button', { name: /Importovat data/ });
    expect(button).toBeInTheDocument();
    expect(button).toHaveTextContent('Importovat data');
  });
});

