import React, { ChangeEvent, useContext, useEffect, useMemo, useState } from 'react';

import { Option } from 'react-dropdown';
import { AppStatus, SortingOption } from '../../AppStatus';
import { DataContext } from '../../main';
import styles from "./index.module.less";

interface Props {
  // status: AppStatus
  // isProgress: boolean
  loadSongs: () => void
  setFilter: (filter: string) => void
}

export function InfoBox(props: Props) {
  const { status, updateStatus } = useContext(DataContext);

  const [resultText, setResultText] = useState("Zpěvník není inicializován");
  const [buttonText, setButtonText] = useState("Stáhnout data z internetu");
  const [, setError] = useState(false);

  const isButtonVisible = useMemo(() => {
    return !(status.DatabaseReady && status.SongsReady);
  }, [status.DatabaseReady, status.SongsReady]);

  function fetchStatus() {
    try {
      // Assume fetchData returns a Promise
      const result = { ...status };
      console.log("Status fetched", result);
      if (result.DatabaseReady) {
        setResultText("Data jsou připravena");
      } else if (result.SongsReady) {
        setResultText("Data jsou stažena, ale nejsou naimportována do interní datbáze");
        setButtonText("Importovat data");
      }
    } catch (error) {
      setError(true);
    }
  }

  useEffect(() => {
    // Use a timer to debounce the onChange event
    const timer = setTimeout(() => {
      props.setFilter(status.SearchPattern);
    }, 500); // Adjust the delay as needed (e.g., 1000ms for 1 second)

    // Clear the timer if the component unmounts or if the input value changes before the timer expires
    return () => clearTimeout(timer);
  }, [status.SearchPattern]);

  // useEffect with an empty dependency array runs once when the component mounts
  useEffect(() => {
    // Delay action after page render (500ms delay in this case)
    const timer = setTimeout(() => {
      console.log('This runs after 600ms delay');
      fetchStatus()
    }, 600);

    // Cleanup function to clear the timeout if the component unmounts
    return () => clearTimeout(timer);
  }, []);

  const sorting = [
    { value: 'entry' as SortingOption, label: 'čísla' },
    { value: 'title' as SortingOption, label: 'názvu' },
    { value: 'authorMusic' as SortingOption, label: 'autora hudby' },
    { value: 'authorLyric' as SortingOption, label: 'autora textu' }
  ];

  function _onSelectSorting(arg: Option): void {
    console.info(arg)
    const stat = { ...status }
    stat.Sorting = arg.value as SortingOption
    updateStatus(stat)
  }

  function _on(event: ChangeEvent<HTMLSelectElement>): void {
    console.info(event.target.value)
    const stat = { ...status }
    stat.Sorting = event.target.value as SortingOption
    updateStatus(stat)
  }

  return (
    <div className="InfoBox">
      {isButtonVisible && <div>
        {resultText} &gt;&gt;&gt;
        {!status.IsProgress && <button className="btn" onClick={() => {
          props.loadSongs()
        }}>{buttonText}</button>}
      </div>
      }
      {status.IsProgress && <div>Připravuji data, vyčkejte ....</div>}
      <div>
        Upozorňujeme, že materiály stahované z <a href='https://www.evangelickyzpevnik.cz/zpevnik/kapitoly-a-pisne/' target="_blank">www.evangelickyzpevnik.cz</a> slouží pouze pro vlastní potřebu a k případnému dalšímu užití je třeba uzavřít licenční smlouvu s nositeli autorských práv.
      </div>
      <div style={{ display: 'flex', gap: '10px', alignItems: 'center', paddingTop: '1em' }}>
        <input id="search-box" className={styles.sorting} type="text" onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
          const newStatus: AppStatus = {
            ...status,
            SearchPattern: e.target.value
          };
          updateStatus(newStatus);
        }} placeholder="Hledat text ..." />
        <select className={styles.sorting} value={status.Sorting} onChange={_on}>
          <option value="" disabled>Řadit podle</option>
          {sorting.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
      </div>
    </div>
  );

}
