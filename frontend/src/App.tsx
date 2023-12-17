import { useState } from 'react';
import { DownloadEz } from "../wailsjs/go/main/App";
import './App.css';

function App() {
    const [resultText, setResultText] = useState("Zpěvník není inicializován");
    const [name, setName] = useState('');
    const updateName = (e: any) => setName(e.target.value);
    const updateResultText = (result: string) => setResultText(result);


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
            <button className="btn" onClick={download}>Stáhnout data z internetu</button>
        </div>
    )
}

export default App
