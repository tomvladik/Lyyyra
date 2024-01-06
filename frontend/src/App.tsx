import React, { useState, useEffect } from 'react';
import './App.less';
import { dtoSong, AppStatus } from './models';
import { DownloadEz, GetSongs, GetStatus } from '../wailsjs/go/main/App';
import { SongList } from './pages/SongList';
import { Search } from './components/search';

export const StatusContext = React.createContext({} as AppStatus);

function App() {
    const [songs, setSongs] = useState(new Array<dtoSong>());
    const [status, setStatus] = useState({} as AppStatus);
    const [error, setError] = useState(false);
    const [resultText, setResultText] = useState("Zpěvník není inicializován");
    const [isButtonVisible, setIsButtonVisible] = useState(true);
    const [filterText, setFilterText] = useState("");
    const [debouncedInputValue, setDebouncedInputValue] = useState("");

    const loadSongs = () => {
        DownloadEz().then(() => {
            fetchData()
        }).catch(error => {
            console.error("Error during download:", error);
        });


    }
    const fetchData = async () => {
        try {
            // Assume fetchData returns a Promise
            const result = await GetSongs();
            fetchStatus()
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
            if (result.DatabaseReady) {
                setResultText("Data jsou připravena")
                fetchData()
                setIsButtonVisible(false)
            } else if (result.SongsReady) {
                setResultText("Data jsou stažena, ale nejsou naimportována do interní datbáze")
            }
        } catch (error) {
            setError(true);
        }
    };


    // useEffect with an empty dependency array runs once when the component mounts
    useEffect(() => {
        fetchStatus()
    }, []);

    useEffect(() => {
        // Use a timer to debounce the onChange event
        const timer = setTimeout(() => {
            setDebouncedInputValue(filterText);
        }, 500); // Adjust the delay as needed (e.g., 1000ms for 1 second)

        // Clear the timer if the component unmounts or if the input value changes before the timer expires
        return () => clearTimeout(timer);
    }, [filterText]);

    return (
        <div id="App">
            <div className="InfoBox">
                {isButtonVisible && <div>
                    {resultText}
                    <button className="btn" onClick={loadSongs}>Stáhnout data z internetu</button></div>}
                <div>
                    Upozorňujeme, že materiály stahované z <a href='https://www.evangelickyzpevnik.cz/zpevnik/kapitoly-a-pisne/' target="_blank">www.evangelickyzpevnik.cz</a> slouží pouze pro vlastní potřebu a k případnému dalšímu užití je třeba uzavřít licenční smlouvu s nositeli autorských práv.
                </div>
                <Search
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                        setFilterText(e.target.value);
                    }}
                />
            </div>
            <div className="ScrollablePart">
                <StatusContext.Provider value={status}>
                    <SongList songs={songs} inputValue={debouncedInputValue} />
                </StatusContext.Provider>
            </div>
        </div>
    )
}

export default App
