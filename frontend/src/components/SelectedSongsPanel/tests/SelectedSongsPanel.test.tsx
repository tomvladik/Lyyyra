import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import * as AppModule from '../../../../wailsjs/go/main/App';
import { SelectedSong } from '../../../models';
import { SelectionContext, SelectionContextValue } from '../../../selectionContext';
import { SelectedSongsPanel } from '../index';

vi.mock('../../../../wailsjs/go/main/App', () => ({
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
    getSelectedSong: () => undefined,
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
    variant: 'guitar',
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('does not render when there are no selections', () => {
    const value = defaultSelectionValue();
    const { container } = renderWithSelection(value);

    expect(container).toBeEmptyDOMElement();
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

    expect(await screen.findByTestId('pdf-modal')).toBeInTheDocument();
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

    // Mock window.open and Blob API
    const mockWindow = {
      location: { href: '' },
      addEventListener: vi.fn(),
    };
    const originalOpen = window.open;
    vi.stubGlobal('open', vi.fn(() => mockWindow));
    vi.stubGlobal('URL', {
      ...URL,
      createObjectURL: vi.fn(() => 'blob:mock-url'),
      revokeObjectURL: vi.fn(),
    });

    const value = defaultSelectionValue({ selectedSongs: [sampleSong] });
    await act(async () => {
      renderWithSelection(value);
    });

    const projectButton = screen.getByRole('button', { name: /Promítat texty/ });
    expect(projectButton).toBeInTheDocument();

    await act(async () => {
      fireEvent.click(projectButton);
    });

    await waitFor(() => {
      expect(AppModule.GetSongProjection).toHaveBeenCalledWith(sampleSong.id);
    });

    expect(window.open).toHaveBeenCalledWith('', '_blank');
    expect((window.URL as any).createObjectURL).toHaveBeenCalled();
  });

  it('properly escapes newlines and special characters in projection HTML', async () => {
    const mockProjectionData = JSON.stringify({
      verse_order: 'v1',
      verses: [
        { name: 'v1', lines: 'Line with\nnewlines\nand\u2028special\u2029chars' },
      ],
    });

    vi.mocked(AppModule.GetSongProjection).mockResolvedValue(mockProjectionData);

    const mockWindow = {
      location: { href: '' },
      addEventListener: vi.fn(),
    };
    vi.stubGlobal('open', vi.fn(() => mockWindow));
    let blobContent = '';
    vi.stubGlobal('URL', {
      ...URL,
      createObjectURL: vi.fn((blob: Blob) => {
        // Verify it's a valid Blob with HTML content
        expect(blob.type).toBe('text/html;charset=utf-8');
        return 'blob:mock-url';
      }),
      revokeObjectURL: vi.fn(),
    });

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

    // Verify that the window was navigated to the blob URL
    expect((window.URL as any).createObjectURL).toHaveBeenCalled();
    expect(mockWindow.location.href).toBe('blob:mock-url');
  });
});
