import React, { useContext, useEffect, useMemo, useState } from 'react';

import Dropdown, { Option } from 'react-dropdown';
import { DataContext } from '../../main';
import { Search } from '../search';
import styles from './index.module.less';

interface Props {
  // status: AppStatus
  // isProgress: boolean
  loadSongs: () => void
  setFilter: (filter: string) => void
}
export function InfoBox(props: Props) {
  const { status } = useContext(DataContext);

  const [resultText, setResultText] = useState("Zpěvník není inicializován");
  const [buttonText, setButtonText] = useState("Stáhnout data z internetu");
  const [filterText, setFilterText] = useState("");
  const [, setError] = useState(false);

  const isButtonVisible = useMemo(() => {
    return !(status.DatabaseReady && status.SongsReady && status.WebResourcesReady);
  }, [status.DatabaseReady, status.SongsReady, status.WebResourcesReady]);

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
      props.setFilter(filterText);
    }, 500); // Adjust the delay as needed (e.g., 1000ms for 1 second)

    // Clear the timer if the component unmounts or if the input value changes before the timer expires
    return () => clearTimeout(timer);
  }, [filterText]);

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
    { value: 'entry', label: 'čísla' },
    { value: 'title', label: 'názvu' },
    { value: 'authorMusic', label: 'autora hudby' },
    { value: 'authorLyric', label: 'autora textu' }
  ];
  const sortingDdefault = sorting[1]

  function _onSelectSorting(arg: Option): void {
    console.info(arg)
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
      <Search
        onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
          setFilterText(e.target.value);
        }
        } />
      <Dropdown className={styles.sorting} options={sorting} onChange={_onSelectSorting} value={sortingDdefault} placeholder="Select an option" />;
    </div>
  );

}
