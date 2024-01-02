import React, { useContext, useState } from 'react';

import styles from './index.module.less';
import { DownloadEz } from '../../../wailsjs/go/main/App';
import { DataContext } from '../../App';

export interface InfoBoxProps {
  prop?: string;
}

export function InfoBox({ prop = 'default value' }: InfoBoxProps) {
  const [resultText, setResultText] = useState("Zpěvník není inicializován");
  const [isVisible, setIsVisible] = useState(true);
  const data = useContext(DataContext);


  function download() {
    setResultText("Stahuji data")
    DownloadEz().then(() => {
      setResultText("Data jsou připravena")
      //!!!      setIsVisible(false)
      data.setSongs()
    }).catch(error => {
      setResultText("Problém během stahování:" + error)
      console.error("Error during download:", error);
    });

  }

  return (
    <div className={styles.InfoBox}>
      <div>{resultText}</div>
      Upozorňujeme, že materiály stahované z <a href='https://www.evangelickyzpevnik.cz/zpevnik/kapitoly-a-pisne/' target="_blank">www.evangelickyzpevnik.cz</a> slouží pouze pro vlastní potřebu a k případnému dalšímu užití je třeba uzavřít licenční smlouvu s nositeli autorských práv.
      {isVisible && <button className="btn" onClick={download}>Stáhnout data z internetu</button>}
    </div>
  );

}
