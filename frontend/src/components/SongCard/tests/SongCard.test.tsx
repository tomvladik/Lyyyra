import '@testing-library/jest-dom';
import type { RenderResult } from '@testing-library/react';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import * as AppModule from '../../../../wailsjs/go/main/App';
import { AppStatus } from '../../../AppStatus';
import { DataContext } from '../../../context';
import { dtoSong } from '../../../models';
import { SelectionContext, SelectionContextValue } from '../../../selectionContext';
import { createMockStatus } from '../../../test/testHelpers';
import { SongCard } from '../index';

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
      isSongSelected: () => false,
      ...selectionOverrides,
    };

    let result: RenderResult | undefined;
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
    if (!result) {
      throw new Error('Failed to render SongCard');
    }
    return {
      ...result,
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

  it('renders song verses (compact, single paragraph)', async () => {
    vi.mocked(AppModule.GetSongAuthors).mockResolvedValue([]);

    await renderWithContext(mockSong);

    // Verses are collapsed into a single line for compact card display
    expect(screen.getByText(/First verse Second verse Third verse/)).toBeInTheDocument();
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

    const { selectionContextValue } = await renderWithContext(
      songWithPdf,
      {},
      {
        selectedSongs: preselected,
        isSongSelected: (id: number) => preselected.some(song => song.id === id),
      }
    );

    const clipboardButton = screen.getByTitle('Skladba už je ve výběru');
    expect(clipboardButton).toHaveAttribute('aria-disabled', 'true');

    fireEvent.click(clipboardButton);
    expect(selectionContextValue.addSongToSelection).not.toHaveBeenCalled();
  });

  it('collapses internal newlines in verses for compact SongCard display', async () => {
    // Test verse with internal newlines (e.g., from <br /> tags in XML)
    const songWithNewlines: dtoSong = {
      ...mockSong,
      Verses: 'First line\nSecond line\nThird line',
    };
    vi.mocked(AppModule.GetSongAuthors).mockResolvedValue([]);

    const { container } = await renderWithContext(songWithNewlines);

    // Verses should be rendered, but newlines within a single verse block should be collapsed
    // The text should appear as a single line when split by \n\n is empty
    const verseTexts = Array.from(container.querySelectorAll('div')).map(el => el.textContent);
    const hasCollapsedText = verseTexts.some(text => text?.includes('First line') && text?.includes('Second line'));

    expect(hasCollapsedText).toBeTruthy();
  });

  it('preserves paragraph breaks between verses in SongCard rendering', async () => {
    // Test verses separated by paragraph breaks (double newlines)
    const songWithParagraphs: dtoSong = {
      ...mockSong,
      Verses: 'First verse line 1\nFirst verse line 2\n\nSecond verse line 1\nSecond verse line 2',
    };
    vi.mocked(AppModule.GetSongAuthors).mockResolvedValue([]);

    const { container } = await renderWithContext(songWithParagraphs);

    const lyricsDiv = container.querySelector('[class*="lyrics"]');
    expect(lyricsDiv).not.toBeNull();

    // Each paragraph break should yield a separate child div
    const childDivs = lyricsDiv ? Array.from(lyricsDiv.querySelectorAll(':scope > div')) : [];
    expect(childDivs.length).toBeGreaterThanOrEqual(2);
  });

  it('handles verses with both internal newlines and paragraph breaks', async () => {
    // Simulate XML parsing: <br/> becomes \n, verses separated by paragraph breaks
    const complexVerse = `Line 1\nLine 2\nLine 3\n\nVerse 2 Line 1\nVerse 2 Line 2`;
    const songComplex: dtoSong = {
      ...mockSong,
      Verses: complexVerse,
    };
    vi.mocked(AppModule.GetSongAuthors).mockResolvedValue([]);

    await renderWithContext(songComplex);

    // First verse text should be rendered (with internal newlines collapsed)
    expect(screen.getByText(/Line 1.*Line 2.*Line 3/)).toBeInTheDocument();
    // Second verse text should be rendered separately
    expect(screen.getByText(/Verse 2 Line 1.*Verse 2 Line 2/)).toBeInTheDocument();
  });
});
