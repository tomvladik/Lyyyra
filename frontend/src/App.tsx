import { useEffect, useState } from 'react';
import * as go from '../wailsjs/go/main/App';
import './App.less';
import { AppStatus, isEqualAppStatus, SortingOption } from "./AppStatus";
import { InfoBox } from './components/InfoBox';
import StatusPanel from './components/StatusPanel';
import { DataContext } from './main';
import { dtoSong, SelectParams } from './models';
import { SongList } from './pages/SongList';


function App() {
    const [status, setStatus] = useState({} as AppStatus);

    const [songs, setSongs] = useState(new Array<dtoSong>());
    const [, setError] = useState(false);
    const [filterValue, setFilterValue] = useState("");
    const [, setSelectParams] = useState({} as SelectParams);

    const loadSongs = () => {
        const stat = { ...status }
        stat.IsProgress = true
        setStatus(stat)
        go.DownloadEz().then(() => {
            fetchStatus()
            fetchData()
        }).catch(error => {
            console.error("Error during download:", error);
        });
        stat.IsProgress = false
        setStatus({ ...stat })
    }

    const fetchData = async () => {
        try {
            // Assume fetchData returns a Promise
            const songs = await go.GetSongs("title");
            setSongs(songs);
        } catch (error) {
            console.log(error)
            setError(true);
        }
    };

    const fetchStatus = async () => {
        try {
            // Assume fetchData returns a Promise
            const goStatus = await go.GetStatus();
            const newStatus: AppStatus = {
                ...goStatus,
                Sorting: goStatus.Sorting as SortingOption
            };
            newStatus.IsProgress = false
            if (!isEqualAppStatus(newStatus, status)) {
                setStatus(newStatus)
            }
        } catch (error) {
            setError(true);
        }
    };

    useEffect(() => {
        fetchData()
    }, [filterValue]);

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
            <div id="App">
                <InfoBox loadSongs={loadSongs} setFilter={setFilterValue} />
                <div className="ScrollablePart">
                    <SongList songs={songs} filter={filterValue} />
                </div>
                <StatusPanel />
            </div>
        </DataContext.Provider>
    )
}

export default App
