import React, { useState, useEffect } from 'react';
import './App.less';
import { dtoSong } from './models';
import { GetSongs } from '../wailsjs/go/main/App';
import { SongList } from './pages/SongList';
import { InfoBox } from './components/InfoBox';

export const DataContext = React.createContext({
    songs: new Array<dtoSong>(),
    setSongs: () => { }
});

function App() {
    const [songs, setSongs] = useState(new Array<dtoSong>());

    const [error, setError] = useState(false);

    const fetchData = async () => {
        try {
            // Assume fetchData returns a Promise
            const result = await GetSongs();
            setSongs(result);
        } catch (error) {
            setError(true);
        }
    };
    const value = {
        songs: songs,
        setSongs: () => { fetchData() }
    }

    // useEffect with an empty dependency array runs once when the component mounts
    useEffect(() => {
        fetchData();
    }, []);


    return (
        <div id="App">
            <DataContext.Provider value={value}>
                <InfoBox />
                <SongList />
            </DataContext.Provider>
        </div>
    )
}

export default App
