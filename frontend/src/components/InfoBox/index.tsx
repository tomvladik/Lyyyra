import React, { useContext, useEffect, useState } from 'react';

import styles from './index.module.less';
import { Search } from '../search';
import { AppStatus } from "../../AppStatus";
import { Option } from 'react-dropdown';
import Dropdown from 'react-dropdown';
import useDataContext from '../../DataProvider';
//import { useDataContext } from '../../App';

interface Props {
  // status: AppStatus
  // isProgress: boolean
  loadSongs: () => void
  setFilter: (filter: string) => void
}
export function InfoBox(props: Props) {
  const [resultText, setResultText] = useState("Zpěvník není inicializován");
  const [buttonText, setButtonText] = useState("Stáhnout data z internetu");
  const [isButtonVisible, setIsButtonVisible] = useState(true);
  const [filterText, setFilterText] = useState("");
  const [error, setError] = useState(false);
  const { data, updateData } = useDataContext();
  const [isProgress, setIsProgress] = useState(false);


  const fetchStatus = () => {
    try {
      // Assume fetchData returns a Promise
      const result = { ...data };
      setIsProgress(result.IsProgress);
      console.log("Status fetched", result)
      if (result.DatabaseReady) {
        setResultText("Data jsou připravena")
        setIsButtonVisible(false)
      } else if (result.SongsReady) {
        setResultText("Data jsou stažena, ale nejsou naimportována do interní datbáze")
        setButtonText("Importovat data")
      }
    } catch (error) {
      setError(true);
    }
  };

  useEffect(() => {
    setIsProgress(data.IsProgress);
  }, [data.IsProgress])

  useEffect(() => {
    // Use a timer to debounce the onChange event
    const timer = setTimeout(() => {
      props.setFilter(filterText);
    }, 500); // Adjust the delay as needed (e.g., 1000ms for 1 second)

    // Clear the timer if the component unmounts or if the input value changes before the timer expires
    return () => clearTimeout(timer);
  }, [filterText]);

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
        {!isProgress && <button className="btn" onClick={() => {
          setIsProgress(true)
          props.loadSongs()
        }}>{buttonText}</button>}
      </div>
      }
      {isProgress && <div>Připravuji data, vyčkejte ....</div>}
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
