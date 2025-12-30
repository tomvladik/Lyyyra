import { render, screen, waitFor, act, fireEvent } from '@testing-library/react';
import { SongCard } from '../index';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import * as AppModule from '../../../../wailsjs/go/main/App';
import { dtoSong } from '../../../models';
import { DataContext } from '../../../context';
import { SelectionContext, SelectionContextValue } from '../../../selectionContext';
import { createMockStatus } from '../../../test/testHelpers';
import { AppStatus } from '../../../AppStatus';

vi.mock('../../../../wailsjs/go/main/App', () => ({
  GetSongAuthors: vi.fn(),
}));

describe('<SongCard />', () => {
  const mockSong: dtoSong = {
    Id: 1,
    Entry: 123,
    Title: 'Test Song Title',
    Verses: 'First verse\nSecond verse\nThird verse',
    AuthorMusic: '',
    AuthorLyric: '',
    KytaraFile: '',
  };

  const renderWithContext = async (
    song: dtoSong,
    overrides: Partial<AppStatus> = {},
    selectionOverrides: Partial<SelectionContextValue> = {}
  ) => {
    const mockContext = {
      status: createMockStatus({
        DatabaseReady: true,
        SongsReady: true,
        WebResourcesReady: true,
        ...overrides,
      }),
      updateStatus: vi.fn(),
    };

    const selectionContextValue: SelectionContextValue = {
      selectedSongs: [],
      addSongToSelection: vi.fn(),
      removeSongFromSelection: vi.fn(),
      clearSelection: vi.fn(),
      ...selectionOverrides,
    };

    let result;
    await act(async () => {
      result = render(
        <DataContext.Provider value={mockContext}>
          <SelectionContext.Provider value={selectionContextValue}>
            <SongCard data={song} />
          </SelectionContext.Provider>
        </DataContext.Provider>
      );
      await Promise.resolve();
    });
    return {
      ...result!,
      selectionContextValue,
    };
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders song entry number and title', async () => {
    vi.mocked(AppModule.GetSongAuthors).mockResolvedValue([]);
    
    await renderWithContext(mockSong);
    
    expect(screen.getByText('123:', { exact: false })).toBeInTheDocument();
    expect(screen.getByText(/Test Song Title/)).toBeInTheDocument();
  });

  it('renders song verses', async () => {
    vi.mocked(AppModule.GetSongAuthors).mockResolvedValue([]);
    
    await renderWithContext(mockSong);
    
    expect(screen.getByText('First verse')).toBeInTheDocument();
    expect(screen.getByText('Second verse')).toBeInTheDocument();
    expect(screen.getByText('Third verse')).toBeInTheDocument();
  });

  it('fetches and displays authors', async () => {
    const mockAuthors = [
      { Type: 'words', Value: 'John Doe' },
      { Type: 'music', Value: 'Jane Smith' },
    ];
    vi.mocked(AppModule.GetSongAuthors).mockResolvedValue(mockAuthors);
    
    await renderWithContext(mockSong);
    
    await waitFor(() => {
      expect(screen.getByText('John Doe')).toBeInTheDocument();
      expect(screen.getByText('Jane Smith')).toBeInTheDocument();
    });
  });

  it('calls GetSongAuthors with correct song ID', async () => {
    vi.mocked(AppModule.GetSongAuthors).mockResolvedValue([]);
    
    await renderWithContext(mockSong);
    
    await waitFor(() => {
      expect(AppModule.GetSongAuthors).toHaveBeenCalledWith(1);
    });
  });

  it('displays T: prefix for word authors', async () => {
    const mockAuthors = [
      { Type: 'words', Value: 'Lyricist Name' },
    ];
    vi.mocked(AppModule.GetSongAuthors).mockResolvedValue(mockAuthors);
    
    await renderWithContext(mockSong);
    
    await waitFor(() => {
      const authorElement = screen.getByText('Lyricist Name').parentElement;
      expect(authorElement?.textContent).toContain('T:');
    });
  });

  it('displays M: prefix for music authors', async () => {
    const mockAuthors = [
      { Type: 'music', Value: 'Composer Name' },
    ];
    vi.mocked(AppModule.GetSongAuthors).mockResolvedValue(mockAuthors);
    
    await renderWithContext(mockSong);
    
    await waitFor(() => {
      const authorElement = screen.getByText('Composer Name').parentElement;
      expect(authorElement?.textContent).toContain('M:');
    });
  });

  it('highlights matches in title and authors when searching', async () => {
    const mockAuthors = [
      { Type: 'words', Value: 'Test Lyricist' },
      { Type: 'music', Value: 'Test Composer' },
    ];
    vi.mocked(AppModule.GetSongAuthors).mockResolvedValue(mockAuthors);

    const { container } = await renderWithContext(mockSong, { SearchPattern: 'test' });

    await waitFor(() => {
      const marks = container.querySelectorAll('mark');
      expect(marks.length).toBeGreaterThanOrEqual(3);
    });
  });

  it('adds song to selection when clipboard icon is clicked', async () => {
    vi.mocked(AppModule.GetSongAuthors).mockResolvedValue([]);
    const songWithPdf: dtoSong = { ...mockSong, KytaraFile: '123.pdf' };

    const { selectionContextValue } = await renderWithContext(songWithPdf);

    const clipboardButton = screen.getByTitle('Přidat do výběru');
    fireEvent.click(clipboardButton);

    expect(selectionContextValue.addSongToSelection).toHaveBeenCalledWith({
      id: songWithPdf.Id,
      entry: songWithPdf.Entry,
      title: songWithPdf.Title,
      filename: songWithPdf.KytaraFile,
    });
  });

  it('disables selection when song already added', async () => {
    vi.mocked(AppModule.GetSongAuthors).mockResolvedValue([]);
    const songWithPdf: dtoSong = { ...mockSong, KytaraFile: '123.pdf' };
    const preselected = [{
      id: songWithPdf.Id,
      entry: songWithPdf.Entry,
      title: songWithPdf.Title,
      filename: songWithPdf.KytaraFile!,
    }];

    const { selectionContextValue } = await renderWithContext(songWithPdf, {}, { selectedSongs: preselected });

    const clipboardButton = screen.getByTitle('Skladba už je ve výběru');
    expect(clipboardButton).toHaveAttribute('aria-disabled', 'true');

    fireEvent.click(clipboardButton);
    expect(selectionContextValue.addSongToSelection).not.toHaveBeenCalled();
  });
});
