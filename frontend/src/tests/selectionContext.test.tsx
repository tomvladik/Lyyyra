import { renderHook } from '@testing-library/react';
import { useContext } from 'react';
import { describe, expect, it } from 'vitest';
import { SelectionContext } from '../selectionContext';

describe('SelectionContext defaults', () => {
    it('has empty selectedSongs by default', () => {
        const { result } = renderHook(() => useContext(SelectionContext));
        expect(result.current.selectedSongs).toEqual([]);
    });

    it('default addSongToSelection is a no-op', () => {
        const { result } = renderHook(() => useContext(SelectionContext));
        expect(() =>
            result.current.addSongToSelection({ id: 1, entry: 1, title: 'Test', filename: '', hasNotes: false })
        ).not.toThrow();
    });

    it('default removeSongFromSelection is a no-op', () => {
        const { result } = renderHook(() => useContext(SelectionContext));
        expect(() => result.current.removeSongFromSelection(1)).not.toThrow();
    });

    it('default clearSelection is a no-op', () => {
        const { result } = renderHook(() => useContext(SelectionContext));
        expect(() => result.current.clearSelection()).not.toThrow();
    });

    it('default isSongSelected returns false', () => {
        const { result } = renderHook(() => useContext(SelectionContext));
        expect(result.current.isSongSelected(42)).toBe(false);
    });
});
