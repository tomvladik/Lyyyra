import '@testing-library/jest-dom';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import * as AppModule from '../../../../wailsjs/go/app/App';
import { AppStatus } from '../../../AppStatus';
import { DataContext } from '../../../context';
import { createMockStatus } from '../../../test/testHelpers';
import StatusPanel from '../index';

vi.mock('../../../../wailsjs/go/app/App', () => ({
    GetStatus: vi.fn(),
    ResetData: vi.fn(),
}));

describe('<StatusPanel />', () => {
    const mockUpdateStatus = vi.fn();
    const mockOnHide = vi.fn();

    const renderWithContext = (statusOverrides: Partial<AppStatus> = {}) => {
        const mockContext = {
            status: createMockStatus({
                DatabaseReady: true,
                SongsReady: true,
                WebResourcesReady: true,
                BuildVersion: '1.0.0',
                LastSave: '2026-01-15T10:30:00Z',
                ...statusOverrides,
            }),
            updateStatus: mockUpdateStatus,
        };

        return render(
            <DataContext.Provider value={mockContext}>
                <StatusPanel onHide={mockOnHide} />
            </DataContext.Provider>
        );
    };

    beforeEach(() => {
        vi.clearAllMocks();
        vi.mocked(AppModule.GetStatus).mockResolvedValue(createMockStatus());
        vi.mocked(AppModule.ResetData).mockResolvedValue();
    });

    it('renders version information', () => {
        renderWithContext({ BuildVersion: '1.2.3' });
        expect(screen.getByText('1.2.3')).toBeInTheDocument();
    });

    it('renders all status indicators', () => {
        renderWithContext();
        expect(screen.getByText('Verze:')).toBeInTheDocument();
        expect(screen.getByText('Uloženo:')).toBeInTheDocument();
        expect(screen.getByText('Podklady:')).toBeInTheDocument();
        expect(screen.getByText('Databáze:')).toBeInTheDocument();
        expect(screen.getByText('Skladby:')).toBeInTheDocument();
    });

    it('displays OK when all resources are ready', () => {
        renderWithContext({
            WebResourcesReady: true,
            DatabaseReady: true,
            SongsReady: true,
        });

        const okElements = screen.getAllByText('OK');
        expect(okElements).toHaveLength(3); // Podklady, Databáze, Skladby
    });

    it('displays čekám when resources are not ready', () => {
        renderWithContext({
            WebResourcesReady: false,
            DatabaseReady: false,
            SongsReady: false,
        });

        const waitingElements = screen.getAllByText('čekám');
        expect(waitingElements).toHaveLength(3);
    });

    it('shows progress when IsProgress is true', () => {
        renderWithContext({
            IsProgress: true,
            ProgressMessage: 'Loading data...',
            ProgressPercent: 45,
        });

        expect(screen.getByText('Loading data...')).toBeInTheDocument();
        expect(screen.getByText('45%')).toBeInTheDocument();
    });

    it('shows progress when not all resources are ready', () => {
        renderWithContext({
            IsProgress: false,
            WebResourcesReady: false,
            DatabaseReady: true,
            SongsReady: true,
        });

        // Should still show progress indicators when not all ready
        expect(screen.getByText(/Pracuji\.\.\.|Hotovo/)).toBeInTheDocument();
    });

    it('formats date correctly', () => {
        renderWithContext({
            LastSave: '2026-01-15T14:30:00Z',
        });

        // Date should be formatted in Czech locale
        const formattedDate = screen.getByTitle(/15\. 01\. 14:30|15\.01\. 14:30/);
        expect(formattedDate).toBeInTheDocument();
    });

    it('shows --- when LastSave is empty', () => {
        renderWithContext({ LastSave: '' });
        expect(screen.getByText('---')).toBeInTheDocument();
    });

    it('calls onHide when clicked', () => {
        renderWithContext();
        const panel = screen.getByText('Verze:').closest('div');
        if (panel) {
            fireEvent.click(panel);
            expect(mockOnHide).toHaveBeenCalledTimes(1);
        }
    });

    it('stops propagation on double click', () => {
        renderWithContext();
        const panel = screen.getByText('Verze:').closest('div');
        if (panel) {
            fireEvent.doubleClick(panel);
            // Should not call onHide on double click
            expect(mockOnHide).not.toHaveBeenCalled();
        }
    });

    it('disables reset button when busy', () => {
        renderWithContext({ IsProgress: true });
        const resetButton = screen.getByTestId('reset-data-button');
        expect(resetButton).toBeDisabled();
    });

    it('enables reset button when not busy', () => {
        renderWithContext({ IsProgress: false });
        const resetButton = screen.getByTestId('reset-data-button');
        expect(resetButton).not.toBeDisabled();
    });

    it('handles reset data flow with confirmation', async () => {
        const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true);

        renderWithContext({ IsProgress: false });
        const resetButton = screen.getByTestId('reset-data-button');

        await act(async () => {
            fireEvent.click(resetButton);
        });

        expect(confirmSpy).toHaveBeenCalled();

        await waitFor(() => {
            expect(AppModule.ResetData).toHaveBeenCalled();
            expect(AppModule.GetStatus).toHaveBeenCalled();
            expect(mockUpdateStatus).toHaveBeenCalled();
        });

        confirmSpy.mockRestore();
    });

    it('cancels reset when user declines confirmation', async () => {
        const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(false);

        renderWithContext({ IsProgress: false });
        const resetButton = screen.getByTestId('reset-data-button');

        await act(async () => {
            fireEvent.click(resetButton);
        });

        expect(confirmSpy).toHaveBeenCalled();
        expect(AppModule.ResetData).not.toHaveBeenCalled();

        confirmSpy.mockRestore();
    });

    it('handles reset error gracefully', async () => {
        const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => { });
        const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true);
        vi.mocked(AppModule.ResetData).mockRejectedValue(new Error('Reset failed'));

        renderWithContext({ IsProgress: false });
        const resetButton = screen.getByTestId('reset-data-button');

        await act(async () => {
            fireEvent.click(resetButton);
        });

        await waitFor(() => {
            expect(consoleErrorSpy).toHaveBeenCalled();
        });

        // Error message should be displayed
        await waitFor(() => {
            expect(screen.getByText(/Nepodařilo se smazat data/)).toBeInTheDocument();
        });

        confirmSpy.mockRestore();
        consoleErrorSpy.mockRestore();
    });

    it('prevents multiple simultaneous resets', async () => {
        const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true);
        let resolveReset: () => void;
        const resetPromise = new Promise<void>(resolve => {
            resolveReset = resolve;
        });
        vi.mocked(AppModule.ResetData).mockReturnValue(resetPromise);

        renderWithContext({ IsProgress: false });
        const resetButton = screen.getByTestId('reset-data-button');

        // First click starts reset
        fireEvent.click(resetButton);

        // Second click should be ignored (button should be disabled or flag should prevent it)
        fireEvent.click(resetButton);

        // Resolve the reset
        resolveReset!();

        // Should only be called once despite two clicks
        await waitFor(() => {
            expect(AppModule.ResetData).toHaveBeenCalledTimes(1);
        });

        confirmSpy.mockRestore();
    });
});
