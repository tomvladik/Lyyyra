import React, { ChangeEvent, useContext, useEffect, useMemo, useState } from 'react';
import { SortingOption } from '../../AppStatus';
import { DEBOUNCE_DELAY, FETCH_STATUS_DELAY } from '../../constants';
import { DataContext } from '../../context';
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
  const [searchValue, setSearchValue] = useState(status.SearchPattern || '');
  const [, setError] = useState(false);

  const isButtonVisible = useMemo(() => {
    return !(status.DatabaseReady && status.SongsReady);
  }, [status.DatabaseReady, status.SongsReady]);

  useEffect(() => {
    try {
      if (status.DatabaseReady) {
        setResultText("Data jsou připravena");
        setButtonText("Stáhnout data z internetu");
      } else if (status.SongsReady) {
        setResultText("Data jsou stažena, ale nejsou naimportována do interní databáze");
        setButtonText("Importovat data");
      } else {
        setResultText("Zpěvník není inicializován");
        setButtonText("Stáhnout data z internetu");
      }
    } catch (error) {
      setError(true);
    }
  }, [status.DatabaseReady, status.SongsReady]);

  useEffect(() => {
    setSearchValue(status.SearchPattern || '');
  }, [status.SearchPattern]);

  useEffect(() => {
    if (searchValue === (status.SearchPattern || '')) {
      return;
    }
    const timer = setTimeout(() => {
      props.setFilter(searchValue);
      updateStatus({ SearchPattern: searchValue });
    }, DEBOUNCE_DELAY);

    return () => clearTimeout(timer);
  }, [searchValue, status.SearchPattern, props.setFilter, updateStatus]);

  // maintain backwards compatibility with previous delayed logging for debugging
  useEffect(() => {
    const timer = setTimeout(() => {
      console.log('Status check after delay');
    }, FETCH_STATUS_DELAY);
    return () => clearTimeout(timer);
  }, []);

  const sorting = [
    { value: 'entry' as SortingOption, label: 'čísla' },
    { value: 'title' as SortingOption, label: 'názvu' },
    { value: 'authorMusic' as SortingOption, label: 'autora hudby' },
    { value: 'authorLyric' as SortingOption, label: 'autora textu' }
  ];

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
        {!status.IsProgress && <button className={styles.actionButton} onClick={() => {
          props.loadSongs()
        }}>{buttonText}</button>}
      </div>
      }
      {status.IsProgress && <div>
        <div style={{ marginBottom: '12px' }}>{status.ProgressMessage || 'Připravuji data, vyčkejte ....'}</div>
        {status.ProgressPercent > 0 && (
          <div>
            <div style={{
              width: '100%',
              backgroundColor: '#e0e0e0',
              borderRadius: '4px',
              height: '20px',
              overflow: 'hidden'
            }}>
              <div style={{
                width: `${status.ProgressPercent}%`,
                backgroundColor: '#a67460',
                height: '100%',
                transition: 'width 0.3s ease',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: 'white',
                fontSize: '12px',
                fontWeight: 'bold'
              }}>
                {status.ProgressPercent}%
              </div>
            </div>
          </div>
        )}
      </div>}
      <div style={{ marginTop: '12px' }}>
        Upozorňujeme, že materiály stahované z <a href='https://www.evangelickyzpevnik.cz/zpevnik/kapitoly-a-pisne/' target="_blank">www.evangelickyzpevnik.cz</a> slouží pouze pro vlastní potřebu a k případnému dalšímu užití je třeba uzavřít licenční smlouvu s nositeli autorských práv.
      </div>
      <div style={{ marginTop: '12px' }}>
        <div style={{ display: 'flex', gap: '20px', flexWrap: 'wrap' }}>
          <div style={{ flex: '1', minWidth: '250px' }}>
            <label htmlFor="search-box" style={{ display: 'block', marginBottom: '6px', fontSize: '15px', fontWeight: '500', textAlign: 'left' }}>Hledat v textu</label>
            <input
              id="search-box"
              className={styles.sorting}
              type="text"
              value={searchValue}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                setSearchValue(e.target.value);
              }}
              placeholder="Hledat text ..."
            />
          </div>
          <div style={{ flex: '1', minWidth: '200px' }}>
            <label htmlFor="sort-select" style={{ display: 'block', marginBottom: '6px', fontSize: '15px', fontWeight: '500', textAlign: 'left' }}>Řadit podle</label>
            <select id="sort-select" className={styles.sorting} value={status.Sorting} onChange={_on}>
              <option value="" disabled>Vyberte možnost</option>
              {sorting.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </div>
        </div>
      </div>
    </div>
  );

}
