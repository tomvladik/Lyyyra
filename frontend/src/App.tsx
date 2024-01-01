import { useState } from 'react';
import { DownloadEz } from "../wailsjs/go/main/App";
import './App.less';
import { SongList } from './pages/SongList';

function App() {
    const [resultText, setResultText] = useState("Zpěvník není inicializován");



    function download() {
        setResultText("Stahuji data")
        DownloadEz().then(() => {
            setResultText("Data jsou připravena")
        }).catch(error => {
            setResultText("Problém během stahování:" + error)
            console.error("Error during download:", error);
        });

    }

    return (
        <div id="App">
            {/* <img src={logo} id="logo" alt="logo"/> */}
            <div id="result" className="result">{resultText}</div>
            <div className="result">
                Upozorňujeme, že materiály stahované z <a href='https://www.evangelickyzpevnik.cz/zpevnik/kapitoly-a-pisne/' target="_blank">www.evangelickyzpevnik.cz</a> slouží pouze pro vlastní potřebu a k případnému dalšímu užití je třeba uzavřít licenční smlouvu s nositeli autorských práv.
                <button className="btn" onClick={download}>Stáhnout data z internetu</button>
            </div>
            <SongList />
        </div>
    )
}

export default App
