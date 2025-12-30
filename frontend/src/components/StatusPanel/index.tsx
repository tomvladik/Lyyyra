import React, { useContext } from "react";
import styles from "./index.module.less";

import { DataContext } from "../../main";

interface StatusPanelProps {
    onHide?: () => void;
}

const StatusPanel: React.FC<StatusPanelProps> = ({ onHide }) => {
    const { status: data } = useContext(DataContext);

    const handleClick = (event: React.MouseEvent<HTMLDivElement>) => {
        event.stopPropagation();
        onHide?.();
    };

    const handleDoubleClick = (event: React.MouseEvent<HTMLDivElement>) => {
        event.stopPropagation();
    };

    return (
        <div
            className={styles.statusPanel}
            onClick={handleClick}
            onDoubleClick={handleDoubleClick}
        >
            {JSON.stringify(data)}
        </div>
    );
};

export default StatusPanel;
