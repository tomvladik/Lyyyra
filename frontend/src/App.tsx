import { useEffect, useState } from 'react';
import * as go from '../wailsjs/go/main/App';
import './App.less';
import { AppStatus, isEqualAppStatus, SortingOption } from "./AppStatus";
import { InfoBox } from './components/InfoBox';
import StatusPanel from './components/StatusPanel';
import { INITIAL_LOAD_DELAY, STATUS_POLL_INTERVAL } from './constants';
import { DataContext } from './main';
import { dtoSong } from './models';
import { SongList } from './pages/SongList';


const createInitialStatus = (): AppStatus => ({
    SearchPattern: '',
    WebResourcesReady: false,
    SongsReady: false,
    DatabaseReady: false,
    IsProgress: true,
    ProgressMessage: 'Nacitam stav...',
    ProgressPercent: 0,
    LastSave: '',
    Sorting: 'entry' as SortingOption,
});

function App() {
    const [status, setStatus] = useState<AppStatus>(() => createInitialStatus());

    const [songs, setSongs] = useState(new Array<dtoSong>());
    const [filterValue, setFilterValue] = useState("");

    const loadSongs = () => {
        const stat = { ...status }
        stat.IsProgress = true
        setStatus(stat)
        
        // Poll for status updates while in progress
        const pollInterval = setInterval(() => {
            fetchStatus();
        }, STATUS_POLL_INTERVAL);
        
        go.DownloadEz().then(() => {
            clearInterval(pollInterval);
            fetchStatus()
        }).catch(error => {
            clearInterval(pollInterval);
            console.error("Error during download:", error);
            fetchStatus()
        });
    }

    // const fetchData = async () => {
    //     try {
    //         // Assume fetchData returns a Promise
    //         const songs = await go.GetSongs(status.Sorting, status.SearchPattern);
    //         setSongs(songs);
    //     } catch (error) {
    //         console.log(error)
    //     }
    // };

    const fetchStatus = async () => {
        try {
            // Assume fetchData returns a Promise
            const goStatus = await go.GetStatus();
            const newStatus: AppStatus = {
                ...goStatus,
                Sorting: goStatus.Sorting as SortingOption,
                ProgressMessage: goStatus.ProgressMessage || '',
                ProgressPercent: goStatus.ProgressPercent || 0
            };
            if (!isEqualAppStatus(newStatus, status)) {
                setStatus(newStatus)
            }
        } catch (error) {
            console.log(error)
        }
    };

    // useEffect(() => {
    //     fetchData()
    // }, [filterValue, status.Sorting]);

    // useEffect with an empty dependency array runs once when the component mounts
    useEffect(() => {
        fetchStatus();
        const timer = setTimeout(() => {
            console.log('Initial load after delay');
            fetchStatus();
        }, INITIAL_LOAD_DELAY);
        return () => clearTimeout(timer);
    }, []);

    const updateStatus = (newStatus: Partial<AppStatus>) => {
        setStatus(prevStatus => ({ ...prevStatus, ...newStatus }));
        if (newStatus.Sorting) {
            go.SaveSorting(newStatus.Sorting);
        }
    };

    return (
        <DataContext.Provider value={{ status: status, updateStatus: updateStatus }}>
            <div id="App" style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
                <header className="header">
                    <InfoBox loadSongs={loadSongs} setFilter={setFilterValue} />
                </header>

                <main className="ScrollablePart">
                    <SongList />
                </main>
                <footer className="footer">
                    <StatusPanel />
                </footer>
            </div>
        </DataContext.Provider>
    )
}

export default App
