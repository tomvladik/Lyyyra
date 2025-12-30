import { createContext } from 'react';
import { SelectedSong } from './models';

export type SelectionContextValue = {
    selectedSongs: SelectedSong[];
    addSongToSelection: (song: SelectedSong) => void;
    removeSongFromSelection: (id: number) => void;
    clearSelection: () => void;
};

export const SelectionContext = createContext<SelectionContextValue>({
    selectedSongs: [],
    addSongToSelection: () => undefined,
    removeSongFromSelection: () => undefined,
    clearSelection: () => undefined,
});
