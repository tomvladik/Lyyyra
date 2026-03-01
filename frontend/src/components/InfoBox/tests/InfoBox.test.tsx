import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { DEBOUNCE_DELAY } from '../../../constants';
import { DataContext } from '../../../context';
import { createMockStatus } from '../../../test/testHelpers';
import { InfoBox } from '../index';

describe('<InfoBox />', () => {
  const mockLoadSongs = vi.fn();
  const mockUpdateStatus = vi.fn();

  const mockSetSourceFilter = vi.fn();

  const renderInfoBox = (statusOverrides = {}) => {
    const status = createMockStatus(statusOverrides);
    return render(
      <DataContext.Provider value={{ status, updateStatus: mockUpdateStatus, sourceFilter: '', setSourceFilter: mockSetSourceFilter }}>
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

  it('calls loadSongs when action button is clicked', async () => {
    const user = userEvent.setup();
    renderInfoBox();
    const button = screen.getByRole('button', { name: /Stáhnout data/ });
    await user.click(button);
    expect(mockLoadSongs).toHaveBeenCalledTimes(1);
  });

  it('hides action button while in progress', () => {
    renderInfoBox({ IsProgress: true, ProgressMessage: 'Zpracovávám...', ProgressPercent: 0 });
    expect(screen.queryByRole('button', { name: /Stáhnout/ })).not.toBeInTheDocument();
  });

  it('shows progress message when IsProgress is true', () => {
    renderInfoBox({ IsProgress: true, ProgressMessage: 'Nahrávám data...', ProgressPercent: 0 });
    expect(screen.getByText('Nahrávám data...')).toBeInTheDocument();
  });

  it('shows progress bar when ProgressPercent > 0', () => {
    renderInfoBox({ IsProgress: true, ProgressMessage: 'Nahrávám...', ProgressPercent: 42 });
    expect(screen.getByText('42%')).toBeInTheDocument();
  });

  it('renders the search input', () => {
    renderInfoBox({ DatabaseReady: true, SongsReady: true });
    const input = screen.getByRole('textbox');
    expect(input).toBeInTheDocument();
  });

  it('updates search value on input change', async () => {
    renderInfoBox({ DatabaseReady: true, SongsReady: true });
    const input = screen.getByRole('textbox');
    fireEvent.change(input, { target: { value: 'Ježíš' } });
    expect(input).toHaveValue('Ježíš');

    await waitFor(() => {
      expect(mockUpdateStatus).toHaveBeenCalledWith(expect.objectContaining({ SearchPattern: 'Ježíš' }));
    }, { timeout: DEBOUNCE_DELAY + 300 });
  });

  it('shows clear button when search has value', async () => {
    renderInfoBox({ DatabaseReady: true, SongsReady: true, SearchPattern: 'test' });
    const clearBtn = await screen.findByRole('button', { name: /Vymazat/i });
    expect(clearBtn).toBeInTheDocument();
  });

  it('clears search when clear button is clicked', async () => {
    renderInfoBox({ DatabaseReady: true, SongsReady: true, SearchPattern: 'test' });
    const clearBtn = await screen.findByRole('button', { name: /Vymazat/i });
    fireEvent.click(clearBtn);
    const input = screen.getByRole('textbox');
    expect(input).toHaveValue('');

    await waitFor(() => {
      expect(mockUpdateStatus).toHaveBeenCalledWith(expect.objectContaining({ SearchPattern: '' }));
    }, { timeout: DEBOUNCE_DELAY + 300 });
  });

  it('renders the sort dropdown with options', () => {
    renderInfoBox({ DatabaseReady: true, SongsReady: true });
    const selects = screen.getAllByRole('combobox');
    const sortSelect = selects[0];
    expect(sortSelect).toBeInTheDocument();
    expect(screen.getByText('čísla')).toBeInTheDocument();
    expect(screen.getByText('názvu')).toBeInTheDocument();
  });

  it('calls updateStatus with new sorting when dropdown changes', async () => {
    const user = userEvent.setup();
    const status = createMockStatus({ DatabaseReady: true, SongsReady: true, Sorting: 'entry' });
    render(
      <DataContext.Provider value={{ status, updateStatus: mockUpdateStatus, sourceFilter: '', setSourceFilter: vi.fn() }}>
        <InfoBox loadSongs={mockLoadSongs} />
      </DataContext.Provider>
    );
    const sortSelect = screen.getAllByRole('combobox')[0];
    await user.selectOptions(sortSelect, 'title');
    expect(mockUpdateStatus).toHaveBeenCalledWith(expect.objectContaining({ Sorting: 'title' }));
  });

  it('shows data ready state correctly', async () => {
    renderInfoBox({ DatabaseReady: true, SongsReady: true });
    // Button should not be visible when both are ready
    expect(screen.queryByRole('button', { name: /Stáhnout/ })).not.toBeInTheDocument();
  });
});

