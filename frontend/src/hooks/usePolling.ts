import { useEffect } from 'react';

/**
 * Custom hook for polling - executes callback at regular intervals when condition is true
 * @param callback Function to execute on each interval
 * @param interval Milliseconds between each execution
 * @param condition Whether polling should be active
 */
export function usePolling(callback: () => void, interval: number, condition: boolean) {
    useEffect(() => {
        if (!condition) return;
        const id = setInterval(callback, interval);
        return () => clearInterval(id);
    }, [callback, interval, condition]);
}

/**
 * Custom hook for delayed execution - runs callback once after a delay
 * @param callback Function to execute after delay
 * @param delay Milliseconds to wait before execution
 * @param deps Dependency array that triggers re-execution when changed
 */
export function useDelayedEffect(callback: () => void, delay: number, deps: React.DependencyList = []) {
    useEffect(() => {
        const timer = setTimeout(callback, delay);
        return () => clearTimeout(timer);
    }, [callback, delay, ...deps]);
}
