import { render, screen, waitFor, act } from '@testing-library/react';
import { SongCard } from '../index';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import * as AppModule from '../../../../wailsjs/go/main/App';
import { dtoSong } from '../../../models';
import { DataContext } from '../../../context';
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
  };

  const renderWithContext = async (song: dtoSong, overrides: Partial<AppStatus> = {}) => {
    const mockContext = {
      status: createMockStatus({
        DatabaseReady: true,
        SongsReady: true,
        WebResourcesReady: true,
        ...overrides,
      }),
      updateStatus: vi.fn(),
    };

    let result;
    await act(async () => {
      result = render(
        <DataContext.Provider value={mockContext}>
          <SongCard data={song} />
        </DataContext.Provider>
      );
      await Promise.resolve();
    });
    return result!;
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
});
