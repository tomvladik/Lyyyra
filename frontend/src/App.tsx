import React, { useState, useEffect } from 'react';
import './App.less';
import { dtoSong, AppStatus } from './models';
import { DownloadEz, GetSongs, GetStatus } from '../wailsjs/go/main/App';
import { SongList } from './pages/SongList';
import { InfoBox } from './components/InfoBox';

export const StatusContext = React.createContext({} as AppStatus);

function App() {
    const [songs, setSongs] = useState(new Array<dtoSong>());
    const [status, setStatus] = useState({} as AppStatus);
    const [error, setError] = useState(false);

    const loadSongs = () => {
        DownloadEz().then(() => {
            fetchData()
            fetchStatus()
        }).catch(error => {
            console.error("Error during download:", error);
        });


    }
    const fetchData = async () => {
        try {
            // Assume fetchData returns a Promise
            const result = await GetSongs();
            console.log("Data fetched", result)
            setSongs(result);
        } catch (error) {
            console.log(error)
            setError(true);
        }
    };

    const fetchStatus = async () => {
        try {
            // Assume fetchData returns a Promise
            const result = await GetStatus();
            console.log("Status fetched", result)
            setStatus(result);
        } catch (error) {
            setError(true);
        }
    };

    // useEffect with an empty dependency array runs once when the component mounts
    useEffect(() => {
    }, []);
    //fetchData()
    //fetchStatus()


    return (
        <div id="App">
            <StatusContext.Provider value={status}>
                <InfoBox loadFunction={loadSongs} />
                <SongList songs={songs} />
            </StatusContext.Provider>
        </div>
    )
}

export default App
