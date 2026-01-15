import '@testing-library/jest-dom';
import { render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import * as AppModule from '../../../../wailsjs/go/app/App';
import { AppStatus } from '../../../AppStatus';
import { DataContext } from '../../../context';
import { dtoSong } from '../../../models';
import { createMockStatus } from '../../../test/testHelpers';
import { SongList } from '../index';

vi.mock('../../../../wailsjs/go/app/App', () => ({
    GetSongs: vi.fn(),
}));

// Mock the polling hooks
vi.mock('../../../hooks/usePolling', () => ({
    usePolling: vi.fn((callback: () => void, _interval: number, condition: boolean) => {
        // Only execute callback if condition is true
        if (condition) {
            // Execute immediately for testing
            callback();
        }
    }),
    useDelayedEffect: vi.fn((callback: () => void) => {
        // Execute immediately for testing
        callback();
    }),
}));

describe('<SongList />', () => {
    const mockUpdateStatus = vi.fn();

    const mockSongs: dtoSong[] = [
        {
            Id: 1,
            Entry: 123,
            Title: 'Test Song 1',
            Verses: 'Verse 1\nVerse 2',
            AuthorMusic: 'Composer 1',
            AuthorLyric: 'Lyricist 1',
            KytaraFile: 'file1.pdf',
        },
        {
            Id: 2,
            Entry: 456,
            Title: 'Test Song 2',
            Verses: 'Verse A\nVerse B',
            AuthorMusic: 'Composer 2',
            AuthorLyric: 'Lyricist 2',
            KytaraFile: 'file2.pdf',
        },
        {
            Id: 3,
            Entry: 789,
            Title: 'Test Song 3',
            Verses: 'Verse X\nVerse Y',
            AuthorMusic: 'Composer 3',
            AuthorLyric: 'Lyricist 3',
            KytaraFile: '',
        },
    ];

    const renderWithContext = (statusOverrides: Partial<AppStatus> = {}) => {
        const mockContext = {
            status: createMockStatus({
                DatabaseReady: true,
                SongsReady: true,
                WebResourcesReady: true,
                SearchPattern: '',
                Sorting: 'entry',
                ...statusOverrides,
            }),
            updateStatus: mockUpdateStatus,
        };

        return render(
            <DataContext.Provider value={mockContext}>
                <SongList />
            </DataContext.Provider>
        );
    };

    beforeEach(() => {
        vi.clearAllMocks();
        vi.mocked(AppModule.GetSongs).mockResolvedValue(mockSongs);
    });

    it('renders without crashing', async () => {
        renderWithContext();
        await waitFor(() => {
            expect(screen.getByText('Test Song 1')).toBeInTheDocument();
        });
    });

    it('fetches and displays songs on mount', async () => {
        renderWithContext();

        await waitFor(() => {
            expect(AppModule.GetSongs).toHaveBeenCalled();
        });

        expect(screen.getByText('Test Song 1')).toBeInTheDocument();
        expect(screen.getByText('Test Song 2')).toBeInTheDocument();
        expect(screen.getByText('Test Song 3')).toBeInTheDocument();
    });

    it('fetches songs with correct sorting option', async () => {
        renderWithContext({ Sorting: 'title' });

        await waitFor(() => {
            expect(AppModule.GetSongs).toHaveBeenCalledWith('title', '');
        });
    });

    it('fetches songs with search pattern', async () => {
        renderWithContext({ SearchPattern: 'test search' });

        await waitFor(() => {
            expect(AppModule.GetSongs).toHaveBeenCalledWith('entry', 'test search');
        });
    });

    it('refetches songs when sorting changes', async () => {
        const { rerender } = renderWithContext({ Sorting: 'entry' });

        await waitFor(() => {
            expect(AppModule.GetSongs).toHaveBeenCalledWith('entry', '');
        });

        // Change sorting
        const mockContext = {
            status: createMockStatus({
                DatabaseReady: true,
                SongsReady: true,
                WebResourcesReady: true,
                SearchPattern: '',
                Sorting: 'authorMusic',
            }),
            updateStatus: mockUpdateStatus,
        };

        rerender(
            <DataContext.Provider value={mockContext}>
                <SongList />
            </DataContext.Provider>
        );

        await waitFor(() => {
            expect(AppModule.GetSongs).toHaveBeenCalledWith('authorMusic', '');
        });
    });

    it('refetches songs when search pattern changes', async () => {
        const { rerender } = renderWithContext({ SearchPattern: '' });

        await waitFor(() => {
            expect(AppModule.GetSongs).toHaveBeenCalledWith('entry', '');
        });

        // Change search pattern
        const mockContext = {
            status: createMockStatus({
                DatabaseReady: true,
                SongsReady: true,
                WebResourcesReady: true,
                SearchPattern: 'new search',
                Sorting: 'entry',
            }),
            updateStatus: mockUpdateStatus,
        };

        rerender(
            <DataContext.Provider value={mockContext}>
                <SongList />
            </DataContext.Provider>
        );

        await waitFor(() => {
            expect(AppModule.GetSongs).toHaveBeenCalledWith('entry', 'new search');
        });
    });

    it('renders all songs returned from API', async () => {
        renderWithContext();

        await waitFor(() => {
            const songCards = screen.getAllByText(/Test Song \d/);
            expect(songCards).toHaveLength(3);
        });
    });

    it('handles empty song list', async () => {
        vi.mocked(AppModule.GetSongs).mockResolvedValue([]);

        renderWithContext();

        await waitFor(() => {
            expect(AppModule.GetSongs).toHaveBeenCalled();
        });

        // Should not render any song cards
        expect(screen.queryByText(/Test Song/)).not.toBeInTheDocument();
    });

    it('handles API error gracefully', async () => {
        const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => { });
        vi.mocked(AppModule.GetSongs).mockRejectedValue(new Error('API Error'));

        renderWithContext();

        await waitFor(() => {
            expect(consoleErrorSpy).toHaveBeenCalledWith(
                'Failed to fetch songs:',
                expect.any(Error)
            );
        });

        consoleErrorSpy.mockRestore();
    });

    it('renders SongCard components with correct keys', async () => {
        renderWithContext();

        await waitFor(() => {
            expect(AppModule.GetSongs).toHaveBeenCalled();
        });

        // Verify that each song is rendered (keys are not visible in DOM, but we can verify rendering)
        mockSongs.forEach(song => {
            expect(screen.getByText(song.Title)).toBeInTheDocument();
        });
    });

    it('polls for songs when database is being filled', () => {
        renderWithContext({
            IsProgress: true,
            DatabaseReady: false,
            SongsReady: true,
        });

        // The usePolling mock will execute callback immediately
        // Should have called GetSongs at least once
        expect(AppModule.GetSongs).toHaveBeenCalled();
    });

    it('does not poll when not in progress', () => {
        vi.clearAllMocks();

        renderWithContext({
            IsProgress: false,
            DatabaseReady: true,
            SongsReady: true,
        });

        // Only initial fetch should occur, no polling
        expect(AppModule.GetSongs).toHaveBeenCalledTimes(2); // Initial + delayed effect
    });

    it('does not poll when database is ready', () => {
        vi.clearAllMocks();

        renderWithContext({
            IsProgress: true,
            DatabaseReady: true,
            SongsReady: true,
        });

        // Should not poll because DatabaseReady is true
        expect(AppModule.GetSongs).toHaveBeenCalledTimes(2); // Initial + delayed effect only
    });

    it('does not poll when songs are not ready', () => {
        vi.clearAllMocks();

        renderWithContext({
            IsProgress: true,
            DatabaseReady: false,
            SongsReady: false,
        });

        // Should not poll because SongsReady is false
        expect(AppModule.GetSongs).toHaveBeenCalledTimes(2); // Initial + delayed effect only
    });
});
