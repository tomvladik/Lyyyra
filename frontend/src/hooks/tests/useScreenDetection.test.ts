import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useScreenDetection } from '../useScreenDetection';

describe('useScreenDetection', () => {
    beforeEach(() => {
        // Ensure getScreenDetails is not defined by default (jsdom doesn't have it)
        // @ts-expect-error intentional deletion for tests
        delete (window as Window & { getScreenDetails?: unknown }).getScreenDetails;

        // Reset screen mock
        Object.defineProperty(window, 'screen', {
            value: {
                width: 1280,
                height: 800,
                isExtended: false,
            },
            writable: true,
            configurable: true,
        });
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    it('returns primary screen via fallback when getScreenDetails is not available', async () => {
        const { result } = renderHook(() => useScreenDetection());

        await act(async () => {
            await Promise.resolve();
        });

        expect(result.current).toHaveLength(1);
        expect(result.current[0]).toMatchObject({
            left: 0,
            top: 0,
            width: 1280,
            height: 800,
            isPrimary: true,
            label: 'Primary Display',
        });
    });

    it('returns two screens when isExtended is true', async () => {
        Object.defineProperty(window, 'screen', {
            value: { width: 1920, height: 1080, isExtended: true },
            writable: true,
            configurable: true,
        });

        const { result } = renderHook(() => useScreenDetection());

        await act(async () => {
            await Promise.resolve();
        });

        expect(result.current).toHaveLength(2);
        expect(result.current[0].isPrimary).toBe(true);
        expect(result.current[1].isPrimary).toBe(false);
        expect(result.current[1].left).toBe(1920);
    });

    it('returns single screen with label "Display 1" when getScreenDetails throws', async () => {
        (window as Window & { getScreenDetails: () => Promise<unknown> }).getScreenDetails = vi
            .fn()
            .mockRejectedValue(new Error('Permission denied'));

        const { result } = renderHook(() => useScreenDetection());

        await act(async () => {
            await Promise.resolve();
        });

        expect(result.current).toHaveLength(1);
        expect(result.current[0].label).toBe('Display 1');
        expect(result.current[0].isPrimary).toBe(true);
    });

    it('uses Screen Detection API when getScreenDetails is available', async () => {
        const mockScreenDetails = {
            screens: [
                { left: 0, top: 0, width: 2560, height: 1440, isPrimary: true, label: 'Built-in Display' },
                { left: 2560, top: 0, width: 1920, height: 1080, isPrimary: false, label: 'External Monitor' },
            ],
        };
        (window as Window & { getScreenDetails: () => Promise<unknown> }).getScreenDetails = vi
            .fn()
            .mockResolvedValue(mockScreenDetails);

        const { result } = renderHook(() => useScreenDetection());

        await act(async () => {
            await Promise.resolve();
        });

        expect(result.current).toHaveLength(2);
        expect(result.current[0].width).toBe(2560);
        expect(result.current[0].label).toBe('Built-in Display');
        expect(result.current[1].left).toBe(2560);
        expect(result.current[1].label).toBe('External Monitor');
    });

    it('falls back to index-based label when screen label is empty', async () => {
        const mockScreenDetails = {
            screens: [
                { left: 0, top: 0, width: 1920, height: 1080, isPrimary: true, label: '' },
            ],
        };
        (window as Window & { getScreenDetails: () => Promise<unknown> }).getScreenDetails = vi
            .fn()
            .mockResolvedValue(mockScreenDetails);

        const { result } = renderHook(() => useScreenDetection());

        await act(async () => {
            await Promise.resolve();
        });

        expect(result.current[0].label).toBe('Display 1');
    });
});
