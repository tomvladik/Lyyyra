import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { SelectedSongsPanel } from '../index';
import { SelectionContext, SelectionContextValue } from '../../../selectionContext';
import { SelectedSong } from '../../../models';
import * as AppModule from '../../../../wailsjs/go/main/App';

vi.mock('../../../../wailsjs/go/main/App', () => ({
  GetCombinedPdf: vi.fn(),
}));

vi.mock('../../PdfModal', () => ({
  PdfModal: ({ isOpen }: { isOpen: boolean }) => (isOpen ? <div data-testid="pdf-modal">modal</div> : null),
}));

describe('<SelectedSongsPanel />', () => {
  const defaultSelectionValue = (overrides: Partial<SelectionContextValue> = {}): SelectionContextValue => ({
    selectedSongs: [],
    addSongToSelection: vi.fn(),
    removeSongFromSelection: vi.fn(),
    clearSelection: vi.fn(),
    isSongSelected: () => false,
    ...overrides,
  });

  const renderWithSelection = (value: SelectionContextValue) => {
    const renderResult = render(
      <SelectionContext.Provider value={value}>
        <SelectedSongsPanel />
      </SelectionContext.Provider>
    );

    return {
      ...renderResult,
      selectionValue: value,
    };
  };

  const sampleSong: SelectedSong = {
    id: 1,
    entry: 10,
    title: 'Test song',
    filename: 'song.pdf',
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('shows empty state and disables controls with no selection', () => {
    const value = defaultSelectionValue();
    renderWithSelection(value);

    expect(screen.getByText(/Výběr je prázdný/)).toBeInTheDocument();
    expect(screen.getByText(/Klepněte na ikonu/)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Vyčistit/ })).toBeDisabled();
    expect(screen.getByRole('button', { name: /Zobrazit společné PDF/ })).toBeDisabled();
  });

  it('calls remove handler when clicking remove button', () => {
    const value = defaultSelectionValue({ selectedSongs: [sampleSong] });
    renderWithSelection(value);

    const removeButton = screen.getByTitle('Odebrat ze seznamu');
    fireEvent.click(removeButton);

    expect(value.removeSongFromSelection).toHaveBeenCalledWith(sampleSong.id);
  });

  it('invokes clearSelection when clear button pressed', () => {
    const value = defaultSelectionValue({ selectedSongs: [sampleSong] });
    renderWithSelection(value);

    const clearButton = screen.getByRole('button', { name: /Vyčistit/ });
    fireEvent.click(clearButton);

    expect(value.clearSelection).toHaveBeenCalledTimes(1);
  });

  it('requests combined PDF and opens modal', async () => {
    vi.mocked(AppModule.GetCombinedPdf).mockResolvedValue('data:pdf');
    const value = defaultSelectionValue({ selectedSongs: [sampleSong] });

    await act(async () => {
      renderWithSelection(value);
    });

    const combineButton = screen.getByRole('button', { name: /Zobrazit společné PDF/ });
    fireEvent.click(combineButton);

    await waitFor(() => {
      expect(AppModule.GetCombinedPdf).toHaveBeenCalledWith([sampleSong.filename]);
    });

    expect(await screen.findByTestId('pdf-modal')).toBeInTheDocument();
  });
});
