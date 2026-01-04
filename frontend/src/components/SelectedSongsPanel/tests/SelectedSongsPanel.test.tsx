import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import * as AppModule from '../../../../wailsjs/go/app/App';
import { SelectedSong } from '../../../models';
import { SelectionContext, SelectionContextValue } from '../../../selectionContext';
import { SelectedSongsPanel } from '../index';

vi.mock('../../../../wailsjs/go/app/App', () => ({
  GetCombinedPdf: vi.fn(),
  GetSongProjection: vi.fn(),
  GetSongVerses: vi.fn(),
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
    hasNotes: true,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    // Mock window.open to return a mock window object
    vi.spyOn(window, 'open').mockReturnValue({
      location: { href: '' },
      document: {
        open: vi.fn(),
        write: vi.fn(),
        close: vi.fn(),
      },
      addEventListener: vi.fn(),
      closed: false,
    } as unknown as Window);

    // Mock URL.createObjectURL and revokeObjectURL
    global.URL.createObjectURL = vi.fn(() => 'blob:mock-url');
    global.URL.revokeObjectURL = vi.fn();
  });

  it('does not render when there are no selections', () => {
    const value = defaultSelectionValue();
    const { container } = renderWithSelection(value);

    expect(container.firstChild).toBeNull();
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

    const clearButton = screen.getByRole('button', { name: /Zrušit/ });
    fireEvent.click(clearButton);

    expect(value.clearSelection).toHaveBeenCalledTimes(1);
  });

  it('requests combined PDF and opens modal', async () => {
    vi.mocked(AppModule.GetCombinedPdf).mockResolvedValue('data:pdf');
    const value = defaultSelectionValue({ selectedSongs: [sampleSong] });

    await act(async () => {
      renderWithSelection(value);
    });

    const combineButton = screen.getByRole('button', { name: /Zobrazit připravené noty/ });
    fireEvent.click(combineButton);

    await waitFor(() => {
      expect(AppModule.GetCombinedPdf).toHaveBeenCalledWith([sampleSong.filename]);
    });

    expect(await screen.findByTestId('pdf-modal')).toBeTruthy();
  });

  it('renders projection button and opens projection window', async () => {
    const mockProjectionData = JSON.stringify({
      verse_order: 'v1 v2 v1',
      verses: [
        { name: 'v1', lines: 'Line 1\nLine 2\nLine 3' },
        { name: 'v2', lines: 'Verse text\nWith multiple\nLines' },
      ],
    });

    vi.mocked(AppModule.GetSongProjection).mockResolvedValue(mockProjectionData);

    const value = defaultSelectionValue({ selectedSongs: [sampleSong] });
    await act(async () => {
      renderWithSelection(value);
    });

    const projectButton = screen.getByRole('button', { name: /Promítat texty/ });
    expect(projectButton).toBeTruthy();

    await act(async () => {
      fireEvent.click(projectButton);
    });

    await waitFor(() => {
      expect(AppModule.GetSongProjection).toHaveBeenCalledWith(sampleSong.id);
    });
  });

  it('loads songs into projection context', async () => {
    const mockProjectionData = JSON.stringify({
      verse_order: 'v1',
      verses: [
        { name: 'v1', lines: 'Line with\nnewlines' },
      ],
    });

    vi.mocked(AppModule.GetSongProjection).mockResolvedValue(mockProjectionData);

    const value = defaultSelectionValue({ selectedSongs: [sampleSong] });
    await act(async () => {
      renderWithSelection(value);
    });

    const projectButton = screen.getByRole('button', { name: /Promítat texty/ });
    await act(async () => {
      fireEvent.click(projectButton);
    });

    await waitFor(() => {
      expect(AppModule.GetSongProjection).toHaveBeenCalled();
    });
  });

  describe('Projection Window Management', () => {
    it('displays control panel when projection opens', async () => {
      const mockProjectionData = JSON.stringify({
        verse_order: 'v1',
        verses: [{ name: 'v1', lines: 'Test verse' }],
      });

      vi.mocked(AppModule.GetSongProjection).mockResolvedValue(mockProjectionData);

      const value = defaultSelectionValue({ selectedSongs: [sampleSong] });
      await act(async () => {
        renderWithSelection(value);
      });

      const projectButton = screen.getByRole('button', { name: /Promítat texty/ });

      await act(async () => {
        fireEvent.click(projectButton);
      });

      await waitFor(() => {
        // Control panel buttons should appear
        expect(screen.getByTitle('Předchozí píseň')).toBeTruthy();
      }, { timeout: 2000 });
    });

    it('prevents opening multiple projection windows', async () => {
      const mockProjectionData = JSON.stringify({
        verse_order: 'v1',
        verses: [{ name: 'v1', lines: 'Test verse' }],
      });

      vi.mocked(AppModule.GetSongProjection).mockResolvedValue(mockProjectionData);

      const value = defaultSelectionValue({ selectedSongs: [sampleSong] });
      await act(async () => {
        renderWithSelection(value);
      });

      const projectButton = screen.getByRole('button', { name: /Promítat texty/ }) as HTMLButtonElement;

      // First click
      await act(async () => {
        fireEvent.click(projectButton);
      });

      await waitFor(() => {
        expect(AppModule.GetSongProjection).toHaveBeenCalledTimes(1);
      }, { timeout: 2000 });

      // Button is hidden once projection is open to avoid duplicate launches
      await waitFor(() => {
        expect(screen.queryByRole('button', { name: /Promítat texty/ })).toBeNull();
      }, { timeout: 2000 });

      // GetSongProjection should still only be called once (not twice)
      expect(AppModule.GetSongProjection).toHaveBeenCalledTimes(1);
    });
  });
});
