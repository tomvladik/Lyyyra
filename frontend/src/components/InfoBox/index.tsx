import React, { ChangeEvent, useContext, useEffect, useMemo, useState } from 'react';
import { Trans, useTranslation } from 'react-i18next';
import { SortingOption } from '../../AppStatus';
import { DEBOUNCE_DELAY } from '../../constants';
import { DataContext } from '../../context';
import { LanguageSwitcher } from '../LanguageSwitcher';
import styles from "./index.module.less";

interface Props {
  // status: AppStatus
  // isProgress: boolean
  loadSongs: () => void
}

export function InfoBox(props: Props) {
  const { t } = useTranslation();
  const { status, updateStatus, sourceFilter, setSourceFilter } = useContext(DataContext);

  const [resultText, setResultText] = useState("Zpěvník není inicializován");
  const [buttonText, setButtonText] = useState("Stáhnout data z internetu");
  const [searchValue, setSearchValue] = useState(status.SearchPattern || '');

  const isButtonVisible = useMemo(() => {
    return !(status.DatabaseReady && status.SongsReady);
  }, [status.DatabaseReady, status.SongsReady]);

  useEffect(() => {
    if (status.DatabaseReady) {
      setResultText(t('infoBox.dataReady'));
      setButtonText(t('infoBox.downloadData'));
    } else if (status.SongsReady) {
      setResultText(t('infoBox.dataDownloadedNotImported'));
      setButtonText(t('infoBox.importData'));
    } else {
      setResultText(t('infoBox.notInitialized'));
      setButtonText(t('infoBox.downloadData'));
    }
  }, [status.DatabaseReady, status.SongsReady, t]);

  useEffect(() => {
    setSearchValue(status.SearchPattern || '');
  }, [status.SearchPattern]);

  useEffect(() => {
    if (searchValue === (status.SearchPattern || '')) {
      return;
    }
    const timer = setTimeout(() => {
      updateStatus({ SearchPattern: searchValue });
    }, DEBOUNCE_DELAY);

    return () => clearTimeout(timer);
  }, [searchValue, status.SearchPattern, updateStatus]);

  // Initial status check delay removed - no longer needed

  const sorting = [
    { value: 'entry' as SortingOption, label: t('infoBox.sortOptions.entry') },
    { value: 'title' as SortingOption, label: t('infoBox.sortOptions.title') },
    { value: 'authorMusic' as SortingOption, label: t('infoBox.sortOptions.authorMusic') },
    { value: 'authorLyric' as SortingOption, label: t('infoBox.sortOptions.authorLyric') }
  ];

  const sourceOptions = [
    { value: '', label: t('infoBox.sourceOptions.both') },
    { value: 'EZ', label: t('infoBox.sourceOptions.ez') },
    { value: 'KK', label: t('infoBox.sourceOptions.kk') },
  ];

  function _on(event: ChangeEvent<HTMLSelectElement>): void {
    console.info(event.target.value)
    const stat = { ...status }
    stat.Sorting = event.target.value as SortingOption
    updateStatus(stat)
  }

  return (
    <div className="InfoBox">
      <LanguageSwitcher />
      {isButtonVisible && <div>
        {resultText} &gt;&gt;&gt;
        {!status.IsProgress && <button className={styles.actionButton} onClick={() => {
          props.loadSongs()
        }}>{buttonText}</button>}
      </div>
      }
      {status.IsProgress && <div>
        <div style={{ marginBottom: '12px' }}>{status.ProgressMessage || t('infoBox.preparingData')}</div>
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
        <Trans
          i18nKey="infoBox.copyrightNotice"
          components={{
            1: <a href='https://www.evangelickyzpevnik.cz/zpevnik/kapitoly-a-pisne/' target="_blank" rel="noreferrer" />
          }}
        />
      </div>
      <div style={{ marginTop: '12px' }}>
        <div style={{ display: 'flex', gap: '20px', flexWrap: 'wrap' }}>
          <div style={{ flex: '1', minWidth: '250px' }}>
            <label htmlFor="search-box" style={{ display: 'block', marginBottom: '6px', fontSize: '15px', fontWeight: '500', textAlign: 'left' }}>{t('infoBox.searchLabel')}</label>
            <div className={styles.searchWrapper}>
              <input
                id="search-box"
                className={styles.searchInput}
                type="text"
                value={searchValue}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                  setSearchValue(e.target.value);
                }}
                placeholder={t('infoBox.searchPlaceholder')}
              />
              {searchValue && (
                <button
                  className={styles.clearButton}
                  onClick={() => setSearchValue('')}
                  aria-label={t('infoBox.clearSearch')}
                  title={t('infoBox.clearSearch')}
                  type="button"
                >
                  ×
                </button>
              )}
            </div>
          </div>
          <div style={{ flex: '1', minWidth: '200px' }}>
            <label htmlFor="sort-select" style={{ display: 'block', marginBottom: '6px', fontSize: '15px', fontWeight: '500', textAlign: 'left' }}>{t('infoBox.sortLabel')}</label>
            <select id="sort-select" className={styles.sorting} value={status.Sorting} onChange={_on}>
              <option value="" disabled>{t('infoBox.sortPlaceholder')}</option>
              {sorting.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </div>
          <div style={{ flex: '1', minWidth: '160px' }}>
            <label htmlFor="source-select" style={{ display: 'block', marginBottom: '6px', fontSize: '15px', fontWeight: '500', textAlign: 'left' }}>{t('infoBox.sourceFilterLabel')}</label>
            <select
              id="source-select"
              className={styles.sorting}
              value={sourceFilter}
              onChange={(e) => setSourceFilter(e.target.value)}
            >
              {sourceOptions.map((option) => (
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
