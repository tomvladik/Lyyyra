import '@testing-library/jest-dom';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import * as AppModule from '../../../../wailsjs/go/app/App';
import { PdfModal } from '../index';

vi.mock('../../../../wailsjs/go/app/App', () => ({
    GetPdfFile: vi.fn(),
}));

const mockT = vi.fn((key: string) => {
    const translations: Record<string, string> = {
        'pdfModal.close': 'Close',
        'pdfModal.loadingPdf': 'Loading PDF...',
    };
    return translations[key] || key;
});

vi.mock('react-i18next', () => ({
    useTranslation: () => ({
        t: mockT,
    }),
}));

describe('<PdfModal />', () => {
    const mockOnClose = vi.fn();

    beforeEach(() => {
        vi.clearAllMocks();
        vi.mocked(AppModule.GetPdfFile).mockResolvedValue('data:application/pdf;base64,mock');
    });

    it('does not render when isOpen is false', () => {
        render(
            <PdfModal
                isOpen={false}
                filename="test.pdf"
                onClose={mockOnClose}
            />
        );

        expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });

    it('renders when isOpen is true', () => {
        render(
            <PdfModal
                isOpen={true}
                dataUrl="data:application/pdf;base64,mock"
                filename="test.pdf"
                onClose={mockOnClose}
            />
        );

        expect(screen.getByText('test.pdf')).toBeInTheDocument();
    });

    it('displays song number and name in title when provided', () => {
        render(
            <PdfModal
                isOpen={true}
                dataUrl="data:application/pdf;base64,mock"
                filename="song.pdf"
                songNumber={123}
                songName="Amazing Song"
                onClose={mockOnClose}
            />
        );

        expect(screen.getByText('123 – Amazing Song')).toBeInTheDocument();
    });

    it('displays only song name when number is not provided', () => {
        render(
            <PdfModal
                isOpen={true}
                dataUrl="data:application/pdf;base64,mock"
                filename="song.pdf"
                songName="Amazing Song"
                onClose={mockOnClose}
            />
        );

        expect(screen.getByText('Amazing Song')).toBeInTheDocument();
    });

    it('displays filename when no song info provided', () => {
        render(
            <PdfModal
                isOpen={true}
                dataUrl="data:application/pdf;base64,mock"
                filename="document.pdf"
                onClose={mockOnClose}
            />
        );

        expect(screen.getByText('document.pdf')).toBeInTheDocument();
    });

    it('calls onClose when close button is clicked', () => {
        render(
            <PdfModal
                isOpen={true}
                dataUrl="data:application/pdf;base64,mock"
                filename="test.pdf"
                onClose={mockOnClose}
            />
        );

        const closeButton = screen.getByTitle('Close');
        fireEvent.click(closeButton);

        expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('calls onClose when overlay is clicked', () => {
        const { container } = render(
            <PdfModal
                isOpen={true}
                dataUrl="data:application/pdf;base64,mock"
                filename="test.pdf"
                onClose={mockOnClose}
            />
        );

        // Find the modal overlay (the outermost clickable div)
        const overlay = container.firstChild;
        if (overlay) {
            fireEvent.click(overlay as Element);
            expect(mockOnClose).toHaveBeenCalledTimes(1);
        }
    });

    it('does not call onClose when modal content is clicked', () => {
        render(
            <PdfModal
                isOpen={true}
                dataUrl="data:application/pdf;base64,mock"
                filename="test.pdf"
                onClose={mockOnClose}
            />
        );

        const modalContent = screen.getByText('test.pdf').parentElement;
        if (modalContent) {
            fireEvent.click(modalContent);
            expect(mockOnClose).not.toHaveBeenCalled();
        }
    });

    it('calls onClose when Escape key is pressed', async () => {
        render(
            <PdfModal
                isOpen={true}
                dataUrl="data:application/pdf;base64,mock"
                filename="test.pdf"
                onClose={mockOnClose}
            />
        );

        fireEvent.keyDown(document, { key: 'Escape' });

        await waitFor(() => {
            expect(mockOnClose).toHaveBeenCalledTimes(1);
        });
    });

    it('does not call onClose on other key presses', () => {
        render(
            <PdfModal
                isOpen={true}
                dataUrl="data:application/pdf;base64,mock"
                filename="test.pdf"
                onClose={mockOnClose}
            />
        );

        fireEvent.keyDown(document, { key: 'Enter' });
        fireEvent.keyDown(document, { key: 'Space' });

        expect(mockOnClose).not.toHaveBeenCalled();
    });

    it('shows loading message while fetching PDF', async () => {
        vi.mocked(AppModule.GetPdfFile).mockImplementation(
            () => new Promise(resolve => setTimeout(() => resolve('data:pdf'), 100))
        );

        render(
            <PdfModal
                isOpen={true}
                filename="test.pdf"
                onClose={mockOnClose}
            />
        );

        expect(screen.getByText('Loading PDF...')).toBeInTheDocument();
    });

    it('fetches PDF file when filename is provided', async () => {
        render(
            <PdfModal
                isOpen={true}
                filename="song123.pdf"
                onClose={mockOnClose}
            />
        );

        await waitFor(() => {
            expect(AppModule.GetPdfFile).toHaveBeenCalledWith('song123.pdf');
        });
    });

    it('uses dataUrl directly when provided', async () => {
        const dataUrl = 'data:application/pdf;base64,test123';

        render(
            <PdfModal
                isOpen={true}
                dataUrl={dataUrl}
                onClose={mockOnClose}
            />
        );

        await waitFor(() => {
            const iframe = screen.getByTitle(/PDF:/);
            expect(iframe).toHaveAttribute('src', dataUrl);
        });

        // Should not call GetPdfFile when dataUrl is provided
        expect(AppModule.GetPdfFile).not.toHaveBeenCalled();
    });

    it('displays error message when PDF fetch fails', async () => {
        const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => { });
        vi.mocked(AppModule.GetPdfFile).mockRejectedValue(new Error('Network error'));

        render(
            <PdfModal
                isOpen={true}
                filename="test.pdf"
                onClose={mockOnClose}
            />
        );

        await waitFor(() => {
            expect(screen.getByText(/Failed to load PDF:/)).toBeInTheDocument();
        });

        consoleErrorSpy.mockRestore();
    });

    it('displays error when no PDF source is provided', async () => {
        render(
            <PdfModal
                isOpen={true}
                onClose={mockOnClose}
            />
        );

        await waitFor(() => {
            expect(screen.getByText('PDF source is missing')).toBeInTheDocument();
        });
    });

    it('renders iframe with correct src after successful fetch', async () => {
        const mockPdfUrl = 'data:application/pdf;base64,mockdata';
        vi.mocked(AppModule.GetPdfFile).mockResolvedValue(mockPdfUrl);

        render(
            <PdfModal
                isOpen={true}
                filename="test.pdf"
                onClose={mockOnClose}
            />
        );

        await waitFor(() => {
            const iframe = screen.getByTitle('PDF: test.pdf');
            expect(iframe).toHaveAttribute('src', mockPdfUrl);
        });
    });

    it('clears previous error when reopening modal', async () => {
        const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => { });
        vi.mocked(AppModule.GetPdfFile).mockRejectedValueOnce(new Error('Error'));

        const { rerender } = render(
            <PdfModal
                isOpen={true}
                filename="test1.pdf"
                onClose={mockOnClose}
            />
        );

        await waitFor(() => {
            expect(screen.getByText(/Failed to load PDF:/)).toBeInTheDocument();
        });

        // Close and reopen with new file
        rerender(
            <PdfModal
                isOpen={false}
                filename="test1.pdf"
                onClose={mockOnClose}
            />
        );

        vi.mocked(AppModule.GetPdfFile).mockResolvedValue('data:pdf');

        rerender(
            <PdfModal
                isOpen={true}
                filename="test2.pdf"
                onClose={mockOnClose}
            />
        );

        await waitFor(() => {
            expect(screen.queryByText(/Failed to load PDF:/)).not.toBeInTheDocument();
        });

        consoleErrorSpy.mockRestore();
    });
});
