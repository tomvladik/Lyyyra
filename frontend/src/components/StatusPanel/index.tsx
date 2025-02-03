import React, { useContext } from "react";
import styles from "./index.module.less";

import { DataContext } from "../../main";

const StatusPanel: React.FC = () => {
    const { status: data } = useContext(DataContext);

    return (
        <div className={styles.statusPanel}>
            {JSON.stringify(data)}
        </div>
    );
};

export default StatusPanel;
