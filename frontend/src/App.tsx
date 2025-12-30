import { useEffect, useState } from 'react';
import * as go from '../wailsjs/go/main/App';
import './App.less';
import { AppStatus, isEqualAppStatus, SortingOption } from "./AppStatus";
import { InfoBox } from './components/InfoBox';
import StatusPanel from './components/StatusPanel';
import { DataContext } from './main';
import { dtoSong } from './models';
import { SongList } from './pages/SongList';


function App() {
    const [status, setStatus] = useState({} as AppStatus);

    const [songs, setSongs] = useState(new Array<dtoSong>());
    const [filterValue, setFilterValue] = useState("");

    const loadSongs = () => {
        const stat = { ...status }
        stat.IsProgress = true
        setStatus(stat)
        
        // Poll for status updates while in progress
        const pollInterval = setInterval(() => {
            fetchStatus();
        }, 500);
        
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
        // Delay action after page render (500ms delay in this case)
        const timer = setTimeout(() => {
            console.log('This runs after 500ms delay');
            fetchStatus()
        }, 500);

        // Cleanup function to clear the timeout if the component unmounts
        return () => clearTimeout(timer);
    }, []);

    const updateStatus = (newStatus: Partial<AppStatus>) => {
        setStatus(prevStatus => ({ ...prevStatus, ...newStatus }));
        go.SaveSorting(newStatus.Sorting);
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
