import React, { useContext, useEffect, useState } from 'react';

import styles from './index.module.less';
import { DownloadEz } from '../../../wailsjs/go/main/App';
import { StatusContext } from '../../App';

interface ChildProps {
  loadFunction: Function
}
export function InfoBox(props: ChildProps) {
  const [resultText, setResultText] = useState("Zpěvník není inicializován");
  const [isButtonVisible, setIsButtonVisible] = useState(true);
  const status = useContext(StatusContext);
  //console.log(status)

  // function download() {
  //   setResultText("Stahuji data")
  //   DownloadEz().then(() => {
  //     setResultText("Data jsou připravena")
  //     setIsButtonVisible(false)
  //   }).catch(error => {
  //     setResultText("Problém během stahování:" + error)
  //     console.error("Error during download:", error);
  //   });

  // }
  // useEffect with an empty dependency array runs once when the component mounts

  // useEffect with an empty dependency array runs once when the component mounts
  useEffect(() => {
    if (status.DatabaseReady) {
      setResultText("Data jsou připravena")
      //    setIsButtonVisible(false)
    } else if (status.SongsReady) {
      setResultText("Data jsou stažena, ale nejsou naimportována do interní datbáze")
    }
  }, []); // Empty dependency array means it runs once when the component mounts



  return (
    <div className={styles.InfoBox}>
      {isButtonVisible && <div>{resultText}</div>}
      Upozorňujeme, že materiály stahované z <a href='https://www.evangelickyzpevnik.cz/zpevnik/kapitoly-a-pisne/' target="_blank">www.evangelickyzpevnik.cz</a> slouží pouze pro vlastní potřebu a k případnému dalšímu užití je třeba uzavřít licenční smlouvu s nositeli autorských práv.
      {isButtonVisible && <button className="btn" onClick={props.loadFunction()}>Stáhnout data z internetu</button>}

      {/* <script src="/wails/ipc.js"></script>
      <script src="/wails/runtime.js"></script> */}

    </div>
  );

}
