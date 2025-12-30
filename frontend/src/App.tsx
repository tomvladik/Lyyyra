import { MouseEvent, useEffect, useState } from 'react';
import * as go from '../wailsjs/go/main/App';
import './App.less';
import { AppStatus, isEqualAppStatus, SortingOption } from "./AppStatus";
import { InfoBox } from './components/InfoBox';
import StatusPanel from './components/StatusPanel';
import { INITIAL_LOAD_DELAY, STATUS_POLL_INTERVAL } from './constants';
import { DataContext } from './main';
import { dtoSong, SelectedSong } from './models';
import { SelectionContext } from './selectionContext';
import { SelectedSongsPanel } from './components/SelectedSongsPanel';
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
    const [isStatusPanelVisible, setIsStatusPanelVisible] = useState(false);
    const [selectedSongs, setSelectedSongs] = useState<SelectedSong[]>([]);

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

    const handleBackgroundDoubleClick = (_event: MouseEvent<HTMLDivElement>) => {
        if (!isStatusPanelVisible) {
            setIsStatusPanelVisible(true);
        }
    };

    const addSongToSelection = (song: SelectedSong) => {
        setSelectedSongs(prev => {
            if (prev.some(existing => existing.id === song.id)) {
                return prev;
            }
            return [...prev, song];
        });
    };

    const removeSongFromSelection = (id: number) => {
        setSelectedSongs(prev => prev.filter(song => song.id !== id));
    };

    const clearSelection = () => setSelectedSongs([]);

    return (
        <DataContext.Provider value={{ status: status, updateStatus: updateStatus }}>
            <SelectionContext.Provider value={{ selectedSongs, addSongToSelection, removeSongFromSelection, clearSelection }}>
                <div
                    id="App"
                    style={{ display: 'flex', flexDirection: 'column', height: '100%' }}
                    onDoubleClick={handleBackgroundDoubleClick}
                >
                    <header className="header" style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                        <InfoBox loadSongs={loadSongs} setFilter={setFilterValue} />
                    </header>
                    <main className="ScrollablePart">
                        <div className="ContentLayout">
                            <div className="SongListArea">
                                <SongList />
                            </div>
                            <SelectedSongsPanel />
                        </div>
                    </main>
                    {isStatusPanelVisible && (
                        <footer className="footer">
                            <StatusPanel onHide={() => setIsStatusPanelVisible(false)} />
                        </footer>
                    )}
                </div>
            </SelectionContext.Provider>
        </DataContext.Provider>
    )
}

export default App
