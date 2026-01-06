import { useEffect, useState } from "react";

export interface ScreenDetailed {
    left: number;
    top: number;
    width: number;
    height: number;
    isPrimary: boolean;
    label: string;
}

interface ScreenDetails {
    screens: Array<{
        left: number;
        top: number;
        width: number;
        height: number;
        isPrimary: boolean;
        label?: string;
    }>;
}

interface ExtendedScreen {
    width: number;
    height: number;
    isExtended?: boolean;
}

/**
 * Custom hook for detecting available screens/displays.
 * Uses Screen Detection API when available, falls back to basic screen info.
 * 
 * @returns Array of available screens with position and size information
 */
export function useScreenDetection(): ScreenDetailed[] {
    const [availableScreens, setAvailableScreens] = useState<ScreenDetailed[]>([]);

    useEffect(() => {
        const detectScreens = async () => {
            try {
                // Try modern Screen Detection API
                if ('getScreenDetails' in window) {
                    const screenDetails = await (window as Window & { getScreenDetails: () => Promise<ScreenDetails> }).getScreenDetails();
                    const screens: ScreenDetailed[] = screenDetails.screens.map((s, idx: number) => ({
                        left: s.left,
                        top: s.top,
                        width: s.width,
                        height: s.height,
                        isPrimary: s.isPrimary,
                        label: s.label || `Display ${idx + 1}`
                    }));
                    setAvailableScreens(screens);
                } else {
                    // Fallback: use window.screen
                    const primaryScreen: ScreenDetailed = {
                        left: 0,
                        top: 0,
                        width: window.screen.width,
                        height: window.screen.height,
                        isPrimary: true,
                        label: 'Primary Display'
                    };
                    // Check if extended display might exist
                    if ((window.screen as ExtendedScreen).isExtended) {
                        const secondaryScreen: ScreenDetailed = {
                            left: window.screen.width,
                            top: 0,
                            width: window.screen.width,
                            height: window.screen.height,
                            isPrimary: false,
                            label: 'Secondary Display'
                        };
                        setAvailableScreens([primaryScreen, secondaryScreen]);
                    } else {
                        setAvailableScreens([primaryScreen]);
                    }
                }
            } catch (err) {
                // Screen detection not available or denied - default to single screen
                setAvailableScreens([{
                    left: 0,
                    top: 0,
                    width: window.screen.width,
                    height: window.screen.height,
                    isPrimary: true,
                    label: 'Display 1'
                }]);
            }
        };
        detectScreens();
    }, []);

    return availableScreens;
}
