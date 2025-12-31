import React, { useContext } from "react";
import styles from "./index.module.less";

import { DataContext } from "../../main";

interface StatusPanelProps {
    onHide?: () => void;
}

const StatusPanel: React.FC<StatusPanelProps> = ({ onHide }) => {
    const { status: data } = useContext(DataContext);

    const formatDate = (value: string) => {
        if (!value) {
            return "---";
        }
        const parsed = new Date(value);
        if (Number.isNaN(parsed.getTime())) {
            return value;
        }
        return parsed.toLocaleString("cs-CZ", {
            hour: "2-digit",
            minute: "2-digit",
            day: "2-digit",
            month: "2-digit",
        });
    };

    const progressPercent = Math.max(0, Math.min(100, data.ProgressPercent || 0));
    const isBusy = data.IsProgress && progressPercent < 100;
    const progressMessage = data.ProgressMessage?.trim() || (isBusy ? "Pracuji..." : "Hotovo");
    const versionLabel = data.BuildVersion || "dev";
    const lastSaveLabel = formatDate(data.LastSave);
    const allReady = data.WebResourcesReady && data.DatabaseReady && data.SongsReady;
    const showProgress = isBusy || !allReady;

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
            <div className={styles.statusRow}>
                <span className={styles.statusLabel}>Verze:</span>
                <span className={styles.statusValueTight} title={versionLabel}>
                    {versionLabel}
                </span>
            </div>
            {showProgress && (
                <div className={styles.progressRow}>
                    <span className={styles.statusLabel}>{progressMessage}</span>
                    <span className={styles.statusValue}>{progressPercent}%</span>
                </div>
            )}
            <div className={styles.statusRow}>
                <span className={styles.statusLabel}>Uloženo:</span>
                <span className={styles.statusValueTight} title={lastSaveLabel}>
                    {lastSaveLabel}
                </span>
            </div>
            <div className={styles.statusRow}>
                <span className={styles.statusLabel}>Podklady:</span>
                <span className={styles.statusValue}>{data.WebResourcesReady ? "OK" : "čekám"}</span>
            </div>
            <div className={styles.statusRow}>
                <span className={styles.statusLabel}>Databáze:</span>
                <span className={styles.statusValue}>{data.DatabaseReady ? "OK" : "čekám"}</span>
            </div>
            <div className={styles.statusRow}>
                <span className={styles.statusLabel}>Skladby:</span>
                <span className={styles.statusValue}>{data.SongsReady ? "OK" : "čekám"}</span>
            </div>
        </div>
    );
};

export default StatusPanel;
