import React, { useState, useEffect, createContext, useContext } from 'react';
import './App.less';
import { SelectParams, dtoSong } from './models';
import { AppStatus, isEqualAppStatus } from "./AppStatus";
import * as go from '../wailsjs/go/main/App';
import { SongList } from './pages/SongList';
import { InfoBox } from './components/InfoBox';
import _ from 'lodash';
import { DataContext } from './DataProvider';

// interface DataContextProps {
//     data: AppStatus;
//     updateData: (newData: AppStatus) => void;
// }
// const DataContext = createContext<DataContextProps | undefined>(undefined);
// export const useDataContext = () => {
//     const context = useContext(DataContext);
//     if (!context) {
//         throw new Error('useDataContext must be used within a DataProvider');
//     }
//     return context;
// };


function App() {
    const [songs, setSongs] = useState(new Array<dtoSong>());
    //const [status, setStatus] = useState({} as AppStatus);
    // const [isProgress, setIsProgress] = useState(false);
    const [error, setError] = useState(false);
    const [filterValue, setFilterValue] = useState("");
    const [selectParams, setSelectParams] = useState({} as SelectParams);
    const [data, updateData] = useState<AppStatus>({} as AppStatus);

    const loadSongs = () => {
        const stat = { ...data }
        stat.IsProgress = true
        updateData(stat)
        //       setIsProgress(true);
        go.DownloadEz().then(() => {
            fetchStatus()
            fetchData()
        }).catch(error => {
            console.error("Error during download:", error);
        });
        //        setIsProgress(false);
        stat.IsProgress = false
        updateData({ ...stat })
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
            const newStatus = await go.GetStatus();
            newStatus.IsProgress = data.IsProgress
            if (!isEqualAppStatus(newStatus, data)) {
                updateData(newStatus)
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
        fetchStatus()
    }, []);

    return (
        <div id="App">
            <DataContext.Provider value={{ data, updateData }}>
                <InfoBox loadSongs={loadSongs} setFilter={setFilterValue} />
                <div className="ScrollablePart">
                    <SongList songs={songs} filter={filterValue} />
                </div>
            </DataContext.Provider>
        </div>
    )
}

export default App
